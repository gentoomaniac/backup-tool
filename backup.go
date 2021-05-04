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
	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/gentoomaniac/backup-tool/pkg/output/local"
	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

type Backup struct {
	BlockSize   int    `short:"b" help:"Data block size in bytes" default:"52428800"`
	BlockPath   string `short:"p" help:"Where to store the blocks" default:"./blocks/"`
	Name        string `help:"name of the backup" required:""`
	Description string `help:"description for the backup"`
	Secret      string `short:"s" help:"secret"`
	Nonce       string `short:"n" help:"IV"`
	Path        string `help:"path to backup" argument:"" required:""`
	DBPath      string `short:"d" help:"database file with backup meta information" type:"path"`
}

func filterFSObjectsByHash(objects []*db.FSObject, hash []byte) *db.FSObject {
	for _, obj := range objects {
		if bytes.Equal(obj.Hash, hash) {
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

func backup(database db.DB, params *Backup) {
	// encryption / decryption
	var iv []byte
	if params.Nonce == "" {
		iv, _ = aes256.GenerateIV()
	} else {
		decodedNonce, _ := base64.StdEncoding.DecodeString(params.Name)
		iv = []byte(decodedNonce)
	}
	log.Debug().Str("iv", base64.StdEncoding.EncodeToString(iv)).Msg("iv loaded")

	var secretBytes []byte
	if params.Secret == "" {
		secretBytes, _ = aes256.GenerateSecret()
	} else {
		decodedSecret, _ := base64.StdEncoding.DecodeString(params.Secret)
		secretBytes = []byte(decodedSecret)
	}
	log.Debug().Str("secret", base64.StdEncoding.EncodeToString(secretBytes)).Msg("secret loaded")

	// Backup code
	backup := &db.Backup{
		Blocksize:   params.BlockSize,
		Timestamp:   0,
		Objects:     make([]*db.FSObject, 0),
		Name:        params.Name,
		Description: params.Description,
		Expiration:  999999999,
	}

	pathStat, _ := os.Stat(params.Path)

	var files []string
	if pathStat.IsDir() {
		files, _ = filePathWalkDir(params.Path)
	} else {
		files = make([]string, 0)
		files = append(files, params.Path)
	}

	var buffer = make([]byte, params.BlockSize)
	filehasher := sha256.New()

	for _, file := range files {
		log.Info().Msgf("Backing up file %s", file)

		f, err := os.Open(file)
		if err != nil {
			log.Error().Err(err)
			return
		}
		defer f.Close()

		filemeta := &db.FSObject{}
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
			blockMetadata := &db.BlockMeta{
				Hash:   hash[:],
				Name:   []byte(base64.StdEncoding.EncodeToString(encryptedHash)),
				Secret: blockSecret,
				Size:   len(data),
				IV:     iv,
			}
			filehasher.Write(data)

			meta, err := database.GetBlockMeta(blockMetadata.Hash)
			if err != nil {
				log.Error().Err(err).Msg("")
				return
			}
			if meta == nil {
				encryptedData, _ := aes256.Encrypt(data, blockSecret, iv)
				local.Write(encryptedData, blockMetadata, params.BlockPath)
				database.AddBlockToIndex(blockMetadata)
			}
			filemeta.Blocks = append(filemeta.Blocks, blockMetadata)
		}

		hash := filehasher.Sum(nil)
		filemeta.Hash = hash[:]
		log.Debug().Str("hash", fmt.Sprintf("%x", filemeta.Hash)).Int("size", int(filesize)).Msg("")

		fsObjects, err := database.GetFSObj(filemeta.Name, filemeta.Path)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		if filterFSObjectsByHash(fsObjects, filemeta.Hash) == nil {
			database.AddFileToIndex(filemeta)
		}
		backup.Objects = append(backup.Objects, filemeta)

		f.Close()
	}

	database.AddBackupToIndex(backup)
}
