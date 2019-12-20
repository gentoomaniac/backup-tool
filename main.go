package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
	"syscall"

	aes256 "github.com/gentoomaniac/backup-tool/lib/crypt"
	sqlite "github.com/gentoomaniac/backup-tool/lib/db"
	local "github.com/gentoomaniac/backup-tool/lib/output"
	"github.com/gentoomaniac/backup-tool/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
)

type config struct {
	DBPath    string
	BlockPath string
	BlockSize int
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func setupConfig() {
	viper.SetConfigName("config")        // name of config file (without extension)
	viper.AddConfigPath("$HOME/.backup") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		log.Panicf("fatal error config file: %s \n", err)
	}
}

func filterFSObjectsByHash(objects []*model.FSObject, hash []byte) *model.FSObject {
	for _, obj := range objects {
		if bytes.Compare(obj.Hash, hash) == 0 {
			return obj
		}
	}
	return nil
}

var (
	verbose   = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	blocksize = kingpin.Flag("blocksize", "Data block size in bytes").Short('b').Default("52428800").Int()
	db        = kingpin.Flag("database", "Database file with backup meta information").Short('d').Default("backup.db").String()
	file      = kingpin.Arg("file", "File to hash").Required().ExistingFile()
	secret    = kingpin.Flag("secret", "Base64 encoded secret").Short('s').String()
	nonce     = kingpin.Flag("nonce", "Base64 encoded nonce").Short('n').String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	config := &config{
		DBPath:    *db,
		BlockPath: "/home/marco/git-private/backup-tool/blocks",
		BlockSize: *blocksize,
	}

	database, _ := sqlite.InitDB(*db)
	log.Debug("DB initialised")

	// encryption / decryption
	var iv []byte
	if *nonce == "" {
		iv, _ = aes256.GenerateIV()
	} else {
		decodedNonce, _ := base64.StdEncoding.DecodeString(*nonce)
		iv = []byte(decodedNonce)
	}
	log.WithFields(log.Fields{
		"iv": base64.StdEncoding.EncodeToString(iv),
	}).Debug("iv loaded")

	var secretBytes []byte
	if *secret == "" {
		secretBytes, _ = aes256.GenerateSecret()
	} else {
		decodedSecret, _ := base64.StdEncoding.DecodeString(*secret)
		secretBytes = []byte(decodedSecret)
	}
	log.WithFields(log.Fields{
		"secret": base64.StdEncoding.EncodeToString(secretBytes),
	}).Debug("secret loaded")

	// Backup code

	backup := &model.Backup{
		Blocksize:   config.BlockSize,
		Timestamp:   0,
		Objects:     make([]*model.FSObject, 0),
		Name:        "test backup",
		Description: "Just a test entry",
		Expiration:  999999999,
	}

	f, err := os.Open(*file)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	filehasher := sha256.New()
	var buffer = make([]byte, config.BlockSize)

	filestat, _ := os.Stat(*file)
	filemeta := &model.FSObject{}
	filemeta.Name = filepath.Base(*file)
	filemeta.Path = filepath.Dir(*file)
	if stat, ok := filestat.Sys().(*syscall.Stat_t); ok {
		filemeta.User = int(stat.Uid)
		filemeta.Group = int(stat.Gid)
	}
	filemeta.FileMode = filestat.Mode()
	filesize := filestat.Size()

	for {
		bytesread, err := f.Read(buffer)
		if err != nil {
			log.Debug(err)
			break
		}
		filesize += int64(bytesread)

		data := buffer[:bytesread]

		blockSecret, _ := aes256.GenerateSecret()
		hash := sha256.Sum256(data)
		encryptedHash, _ := aes256.Encrypt(hash[:], blockSecret, iv)
		aes256.Encrypt(hash[:], blockSecret, iv)
		blockMetadata := &model.BlockMeta{
			Hash:   hash[:],
			Name:   []byte(base64.StdEncoding.EncodeToString(encryptedHash)),
			Secret: blockSecret,
			Size:   len(data),
			IV:     iv,
		}
		filehasher.Write(data)

		if sqlite.GetBlockMeta(database, blockMetadata.Hash) == nil {
			encryptedData, _ := aes256.Encrypt(data, blockSecret, iv)
			local.Write(encryptedData, blockMetadata, config.BlockPath)
			sqlite.AddBlockToIndex(database, blockMetadata)
		}
		filemeta.Blocks = append(filemeta.Blocks, blockMetadata)
	}

	hash := filehasher.Sum(nil)
	filemeta.Hash = hash[:]
	log.Debugf("File hash: %x", filemeta.Hash)
	log.Debugf("Filse size: %d", filesize)

	fsObjects := sqlite.GetFSObj(database, filemeta.Name, filemeta.Path)
	if filterFSObjectsByHash(fsObjects, filemeta.Hash) == nil {
		sqlite.AddFileToIndex(database, filemeta)
	}

	backup.Objects = append(backup.Objects, filemeta)

}
