package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gentoomaniac/backup-tool/pkg/crypt/aes256"
	"github.com/gentoomaniac/backup-tool/pkg/db/sqlite"
	"github.com/gentoomaniac/backup-tool/pkg/model"
	"github.com/gentoomaniac/backup-tool/pkg/output/local"
	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

func filterFSObjectsByHash(objects []*model.FSObject, hash []byte) *model.FSObject {
	for _, obj := range objects {
		if bytes.Compare(obj.Hash, hash) == 0 {
			return obj
		}
	}
	return nil
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func backup() {
	database, err := sqlite.InitDB(cli.Backup.DBPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed initialising DB")
	}
	log.Debug("DB initialised")

	// encryption / decryption
	var iv []byte
	if cli.Backup.Nonce == "" {
		iv, _ = aes256.GenerateIV()
	} else {
		decodedNonce, _ := base64.StdEncoding.DecodeString(cli.Backup.Name)
		iv = []byte(decodedNonce)
	}
	log.Debug().Str("iv", base64.StdEncoding.EncodeToString(iv)).Msg("iv loaded")

	var secretBytes []byte
	if cli.Backup.Secret == "" {
		secretBytes, _ = aes256.GenerateSecret()
	} else {
		decodedSecret, _ := base64.StdEncoding.DecodeString(cli.Backup.Secret)
		secretBytes = []byte(decodedSecret)
	}
	log.WithFields(log.Fields{
		"secret": base64.StdEncoding.EncodeToString(secretBytes),
	}).Debug("secret loaded")

	// Backup code
	backup := &model.Backup{
		Blocksize:   cli.Backup.BlockSize,
		Timestamp:   0,
		Objects:     make([]*model.FSObject, 0),
		Name:        cli.Backup.Name,
		Description: cli.Backup.Description,
		Expiration:  999999999,
	}

	pathStat, _ := os.Stat(cli.Backup.Path)

	var files []string
	if pathStat.IsDir() {
		files, _ = filePathWalkDir(cli.Backup.Path)
	} else {
		files = make([]string, 0)
		files = append(files, cli.Backup.Path)
	}

	var buffer = make([]byte, cli.Backup.BlockSize)
	filehasher := sha256.New()

	for _, file := range files {
		fmt.Info().Msgf("Backing up file %s", file)

		f, err := os.Open(file)
		if err != nil {
			log.Error().Err(err)
			return
		}
		defer f.Close()

		filemeta := &model.FSObject{}
		filestat, _ := os.Stat(file)
		filemeta.Name = filepath.Base(file)
		filemeta.Path, _ = filepath.Abs(filepath.Dir(file))
		if stat, ok := filestat.Sys().(*syscall.Stat_t); ok {
			filemeta.User = int(stat.Uid)
			filemeta.Group = int(stat.Gid)
		}
		filemeta.FileMode = filestat.Mode()
		filesize := filestat.Size()

		for {
			bytesread, err := f.Read(buffer)
			if err != nil {
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
				local.Write(encryptedData, blockMetadata, cli.Backup.BlockPath)
				sqlite.AddBlockToIndex(database, blockMetadata)
			}
			filemeta.Blocks = append(filemeta.Blocks, blockMetadata)
		}

		hash := filehasher.Sum(nil)
		filemeta.Hash = hash[:]
		log.Debug().Str("hash", filemeta.Hash).Int("size", filesize).Msg("")

		fsObjects := sqlite.GetFSObj(database, filemeta.Name, filemeta.Path)
		if filterFSObjectsByHash(fsObjects, filemeta.Hash) == nil {
			sqlite.AddFileToIndex(database, filemeta)
		}
		backup.Objects = append(backup.Objects, filemeta)

		f.Close()
	}

	sqlite.AddBackupToIndex(database, backup)
}
