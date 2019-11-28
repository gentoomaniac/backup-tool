package local

import (
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/gentoomaniac/backup-tool/src/model"
	log "github.com/sirupsen/logrus"
)

func Write(block *model.Block, basepath string) (int, error) {
	log.Debugf("Writing block: %x", block.Hash)

	blockpath := filepath.Join(basepath, hex.EncodeToString(block.Hash[0:1]), hex.EncodeToString(block.Hash[1:2]))
	os.MkdirAll(blockpath, 0755)

	if _, err := os.Stat(filepath.Join(blockpath, hex.EncodeToString(block.Hash))); os.IsNotExist(err) {
		blockfile, createErr := os.Create(filepath.Join(blockpath, hex.EncodeToString(block.Hash)))
		if createErr != nil {
			log.Error(createErr)
			err = createErr
			return 0, createErr
		}
		defer blockfile.Close()

		bytes, err := blockfile.Write(block.Data)
		return bytes, err
	}

	return 0, nil
}
