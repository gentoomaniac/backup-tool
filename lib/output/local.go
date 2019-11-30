package local

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/gentoomaniac/backup-tool/model"
	log "github.com/sirupsen/logrus"
)

func Write(data []byte, metadata *model.BlockMeta, basepath string) (int, error) {
	log.WithFields(log.Fields{
		"block_secret": base64.StdEncoding.EncodeToString(metadata.Secret),
		"block_hash":   metadata.Hash,
		"block_name":   metadata.Name,
		"block_Size":   metadata.Size,
	}).Debugf("Writing block: %x", metadata.Hash)

	blockpath := filepath.Join(basepath, hex.EncodeToString(metadata.Name[0:1]), hex.EncodeToString(metadata.Name[1:2]))
	os.MkdirAll(blockpath, 0755)

	blockfile, err := os.Create(filepath.Join(blockpath, hex.EncodeToString(metadata.Name)))
	if err != nil {
		log.Error(err)
		return 0, err
	}
	defer blockfile.Close()

	bytes, err := blockfile.Write(data)
	return bytes, err
}
