package local

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/rs/zerolog/log"
)

func Write(data []byte, metadata *db.BlockMeta, basepath string) (int, error) {
	log.Debug().
		Str("block_secret", base64.StdEncoding.EncodeToString(metadata.Secret)).
		Str("block_hash", fmt.Sprintf("%x", metadata.Hash)).
		Str("block_name", fmt.Sprintf("%x", metadata.Name)).
		Int("block_Size", metadata.Size).
		Msg("Writing block")

	blockpath := filepath.Join(basepath, hex.EncodeToString(metadata.Name[0:1]), hex.EncodeToString(metadata.Name[1:2]))
	os.MkdirAll(blockpath, 0755)

	blockfile, err := os.Create(filepath.Join(blockpath, hex.EncodeToString(metadata.Name)))
	if err != nil {
		log.Error().Err(err).Msg("")
		blockfile.Close()
		return 0, err
	}

	bytes, err := blockfile.Write(data)

	blockfile.Close()
	return bytes, err
}
