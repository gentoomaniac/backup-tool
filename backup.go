package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/gentoomaniac/backup-tool/pkg/crypt/aes256"
	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/gentoomaniac/backup-tool/pkg/output/local"
	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

type BackupArgs struct {
	BlockSize   int    `short:"b" help:"Data block size in bytes" default:"52428800"`
	BlockPath   string `short:"p" help:"Where to store the blocks" default:"./blocks/"`
	Name        string `help:"name of the backup" required:""`
	Description string `help:"description for the backup"`
	Secret      string `short:"s" help:"secret"`
	Nonce       string `short:"n" help:"IV"`
	Path        string `help:"path to backup" argument:"" required:"" type:"path"`
	DBPath      string `short:"d" help:"database file with backup meta information" type:"path" required:""`
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)

		return nil
	})
	return files, err
}

func backup(params *BackupArgs) (err error) {
	database, err := db.NewSQLLite(params.DBPath)
	if err != nil {
		log.Error().Err(err).Msg("failed opening db")
	}
	err = database.Init()
	if err != nil {
		log.Error().Err(err).Msg("failed initialising db")
		return
	}
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

	pathStat, err := os.Stat(params.Path)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	var files []string
	if pathStat.IsDir() {
		files, _ = filePathWalkDir(params.Path)
	} else {
		files = make([]string, 0)
		files = append(files, params.Path)
	}

	var buffer = make([]byte, params.BlockSize)
	filehasher := sha256.New()

	log.Debug().Msg("Adding backup to index")
	backup.ID, err = database.AddBackupToIndex(backup)
	if err != nil {
		log.Error().Err(err).Msg("Adding backup to index failed")
		return
	}
	log.Debug().Msg(backup.String())

	for _, file := range files {
		log.Info().Msgf("Backing up file %s", file)

		f, err := os.Open(file)
		if err != nil {
			log.Error().Err(err)
			return err
		}
		defer f.Close()

		filestat, _ := os.Stat(file)
		filemeta := &db.FSObject{
			Name:     filepath.Base(file),
			IsDir:    filestat.IsDir(),
			FileMode: filestat.Mode(),
			BackupID: backup.ID,
		}
		filemeta.Path, _ = filepath.Abs(filepath.Dir(file))
		if stat, ok := filestat.Sys().(*syscall.Stat_t); ok {
			filemeta.User = int(stat.Uid)
			filemeta.Group = int(stat.Gid)
		}
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
				return err
			}
			if meta == nil {
				encryptedData, _ := aes256.Encrypt(data, blockSecret, iv)
				local.Write(encryptedData, blockMetadata, params.BlockPath)
				blockMetadata.ID, err = database.AddBlockToIndex(blockMetadata)
				if err != nil {
					log.Error().Err(err).Str("block", string(blockMetadata.Name)).Msg("failed adding block to index")
				}
			} else {
				log.Debug().Int("id", int(meta.ID)).Str("name", string(meta.Name)).Msg("Block found")
			}
			filemeta.Blocks = append(filemeta.Blocks, blockMetadata)
		}

		hash := filehasher.Sum(nil)
		filemeta.Hash = hash[:]
		log.Debug().Str("hash", fmt.Sprintf("%x", filemeta.Hash)).Int("size", int(filesize)).Msg("")

		filemeta.ID, err = database.AddFileToIndex(filemeta)
		if err != nil {
			log.Error().Err(err).Str("file", path.Join(filemeta.Path, filemeta.Name)).Msg("failed adding file to index")
		}
		// for _, block := range filemeta.Blocks {
		// 	// need to add fileblocks
		// }
		backup.Objects = append(backup.Objects, filemeta)

		f.Close()
	}

	return
}
