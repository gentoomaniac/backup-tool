package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"

	aes256 "github.com/gentoomaniac/backup-tool/lib/crypt"
	sqlite "github.com/gentoomaniac/backup-tool/lib/db"
	local "github.com/gentoomaniac/backup-tool/lib/output"
	"github.com/gentoomaniac/backup-tool/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		database, _ := sqlite.InitDB(db)
		log.Debug("DB initialised")

		// encryption / decryption
		var iv []byte
		if nonce == "" {
			iv, _ = aes256.GenerateIV()
		} else {
			decodedNonce, _ := base64.StdEncoding.DecodeString(nonce)
			iv = []byte(decodedNonce)
		}
		log.WithFields(log.Fields{
			"iv": base64.StdEncoding.EncodeToString(iv),
		}).Debug("iv loaded")

		var secretBytes []byte
		if secret == "" {
			secretBytes, _ = aes256.GenerateSecret()
		} else {
			decodedSecret, _ := base64.StdEncoding.DecodeString(secret)
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

		pathStat, _ := os.Stat(path)

		var files []string
		if pathStat.IsDir() {
			files, _ = filePathWalkDir(path)
		} else {
			files = make([]string, 0)
			files = append(files, path)
		}

		var buffer = make([]byte, config.BlockSize)
		filehasher := sha256.New()

		for _, file := range files {
			fmt.Printf("Backing up file %s", file)

			f, err := os.Open(file)
			if err != nil {
				log.Error(err)
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
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	backupCmd.Flags().IntP("blocksize", "b", 52428800, "Data block size in bytes")
	backupCmd.Flags().StringP("db", "d", "backup.db", "Database file with backup meta information")
	backupCmd.Flags().StringP("path", "p", "", "path to backup")
	backupCmd.Flags().StringP("secret", "s", "", "secret")
	backupCmd.Flags().StringP("nonce", "n", "", "IV")
}
