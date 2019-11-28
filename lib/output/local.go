package local

import (
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/gentoomaniac/backup-tool/src/model"
	log "github.com/sirupsen/logrus"
)

func Write(block *model.Block, basepath string) (bytes int, err error) {
	log.Debugf("Writing block: %x", block.Hash)

	blockpath := filepath.Join(basepath, hex.EncodeToString(block.Hash[0:1]))
	os.Mkdir(blockpath, 0755)

	blockfile, err := os.Create(filepath.Join(blockpath, hex.EncodeToString(block.Hash)))
	if err != nil {
		log.Error(err)
		return
	}
	defer blockfile.Close()

	bytes, err = blockfile.Write(block.Data)

	return
}
