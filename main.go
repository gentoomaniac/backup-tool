package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"./lib/crypt/aes256"

	"github.com/alecthomas/kingpin"
	"github.com/gentoomaniac/backup-tool/src/model"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	file    = kingpin.Arg("file", "File to hash").Required().ExistingFile()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	// encryption / decryption
	iv, _ := aes256.GenerateIV()
	log.WithFields(log.Fields{
		"iv": base64.StdEncoding.EncodeToString(iv),
	}).Debug("iv generated")

	secret, _ := aes256.GenerateSecret()
	log.WithFields(log.Fields{
		"secret": base64.StdEncoding.EncodeToString(secret),
	}).Debug("secret generated")

	encrypted, _ := aes256.Encrypt([]byte("Hello, World!"), secret, iv)
	log.WithFields(log.Fields{
		"encrypted": base64.StdEncoding.EncodeToString(encrypted),
	}).Debug("text encrypted")

	decrypted, _ := aes256.Decrypt(encrypted, secret, iv)
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
	var blocksize = 52428800
	var blockIndex = make(map[string][]byte)

	file, err := os.Open(*file)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	filehasher := sha256.New()
	filemeta := &model.FSObject{}
	filesize := 0
	cwd, _ := os.Getwd()
	var buffer = make([]byte, blocksize)

	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			log.Debug(err)
			break
		}
		filesize += bytesread

		data := buffer[:bytesread]

		hash := sha256.Sum256(data)
		blockIndex[hex.EncodeToString(hash[:])] = data

		filehasher.Write(data)

		blockfile, err := os.Create(filepath.Join(cwd, string(hex.EncodeToString(hash[:]))))
		if err != nil {
			log.Error(err)
			return
		}
		defer blockfile.Close()
		blockbytes, _ := blockfile.Write(data)

		log.Debugf("bytes read: %d", bytesread)
		log.Debugf("block hash: %x", hash)
		log.Debugf("bytes written: %d", blockbytes)

	}

	hash := filehasher.Sum(nil)
	filemeta.Hash = hash[:]
	log.Debugf("File hash: %x", filemeta.Hash)
	log.Debugf("Filse size: %d", filesize)

	file.Seek(0, 0)
	hasher := sha256.New()
	bytesCopied, _ := io.Copy(hasher, file)
	log.Debugf("File hash: %x", hasher.Sum(nil))
	log.Debugf("Filse size: %d", bytesCopied)
}
