package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	aes256 "github.com/gentoomaniac/backup-tool/lib/crypt"
	local "github.com/gentoomaniac/backup-tool/lib/output"
	"github.com/gentoomaniac/backup-tool/model"
	"github.com/hprose/hprose-go"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
)

type config struct {
	BlockPath  string
	BlockSize  int
	BlockIndex string
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func saveBlockIndex(path string, blockindex map[string]*model.BlockMeta) (bytesWritten int, err error) {
	serialized, err := hprose.Serialize(blockindex, true)
	if err != nil {
		log.Error(err)
		return
	}
	blockIndexFile, err := os.Create(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer blockIndexFile.Close()
	bytesWritten, err = blockIndexFile.Write(serialized)

	return
}

func loadlockIndex(path string, blockindex *map[string]*model.BlockMeta) (err error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return
	}

	err = hprose.Unserialize(raw, blockindex, true)
	if err != nil {
		log.Error(err)
	}

	return
}

var (
	verbose    = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	blocksize  = kingpin.Flag("blocksize", "Data block size in bytes").Short('b').Default("52428800").Int()
	blockindex = kingpin.Flag("blockindex", "Saved block index").Short('i').Default("blockindex").String()
	file       = kingpin.Arg("file", "File to hash").Required().ExistingFile()
	secret     = kingpin.Flag("secret", "Base64 encoded secret").Short('s').String()
	nonce      = kingpin.Flag("nonce", "Base64 encoded nonce").Short('n').String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	config := &config{
		BlockPath:  "/home/marco/git-private/backup-tool/blocks",
		BlockSize:  52428800,
		BlockIndex: *blockindex,
	}

	// encryption / decryption
	var iv []byte
	if *nonce == "" {
		iv, _ = aes256.GenerateIV()
	} else {
		iv, _ = base64.StdEncoding.DecodeString(*nonce)
	}
	log.WithFields(log.Fields{
		"iv": base64.StdEncoding.EncodeToString(iv),
	}).Debug("iv loaded")

	var secretBytes []byte
	if *secret == "" {
		secretBytes, _ = aes256.GenerateSecret()
	} else {
		secretBytes, _ = base64.StdEncoding.DecodeString(*secret)
	}
	log.WithFields(log.Fields{
		"secret": base64.StdEncoding.EncodeToString(secretBytes),
	}).Debug("secret loaded")

	encrypted, _ := aes256.Encrypt([]byte("Hello, World!"), secretBytes, iv)
	log.WithFields(log.Fields{
		"encrypted": base64.StdEncoding.EncodeToString(encrypted),
	}).Debug("text encrypted")

	decrypted, _ := aes256.Decrypt(encrypted, secretBytes, iv)
	log.WithFields(log.Fields{
		"decrypted": string(decrypted),
	}).Debug("text decrypted")

	// File stat
	fm, _ := os.Stat("/boot/kernel")
	switch mode := fm.Mode(); {
	case mode.IsRegular():
		fmt.Println("regular file")
	case mode.IsDir():
		fmt.Println("directory")
	case mode&os.ModeSymlink != 0:
		fmt.Println("symbolic link")
	case mode&os.ModeNamedPipe != 0:
		fmt.Println("named pipe")
	}
	log.Debug(fm.Name())
	log.Debug(fm.Mode())
	log.Debug(fm.ModTime())
	log.Debug(fm.Size())
	log.Debug(fm.Sys())

	// read file into data structs
	var blockIndex = make(map[string]*model.BlockMeta)
	err := loadlockIndex(config.BlockIndex, &blockIndex)
	if err != nil {
		log.Warnf("Couldn't open block index: %s", err)
	}
	jsonBlockIndex, _ := json.Marshal(blockIndex)
	fmt.Print(string(jsonBlockIndex))

	file, err := os.Open(*file)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	filehasher := sha256.New()
	filemeta := &model.FSObject{}
	filesize := 0
	var buffer = make([]byte, config.BlockSize)
	var blockbytes int

	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			log.Debug(err)
			break
		}
		filesize += bytesread

		data := buffer[:bytesread]

		blockSecret, _ := aes256.GenerateSecret()
		hash := sha256.Sum256(data)
		encryptedHash, _ := aes256.Encrypt(hash[:], blockSecret, iv)
		blockMetadata := &model.BlockMeta{
			Hash:   hash[:],
			Name:   []byte(base64.StdEncoding.EncodeToString(encryptedHash)),
			Secret: blockSecret,
			Size:   len(data),
		}
		filehasher.Write(data)

		encryptedData, _ := aes256.Encrypt(data, blockSecret, iv)
		if _, ok := blockIndex[hex.EncodeToString(hash[:])]; !ok {
			blockbytes, _ = local.Write(encryptedData, blockMetadata, "/home/marco/git-private/backup-tool/blocks")
			blockIndex[hex.EncodeToString(hash[:])] = blockMetadata
		}

		log.Debugf("bytes read: %d", bytesread)
		log.Debugf("block hash: %x", hash)
		log.Debugf("bytes written: %d", blockbytes)
		json, _ := json.Marshal(blockMetadata)
		fmt.Print(string(json))

	}

	hash := filehasher.Sum(nil)
	filemeta.Hash = hash[:]
	log.Debugf("File hash: %x", filemeta.Hash)
	log.Debugf("Filse size: %d", filesize)

	saveBlockIndex(config.BlockIndex, blockIndex)

}
