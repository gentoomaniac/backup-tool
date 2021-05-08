package main

import (
	"fmt"
	"os"
	"path"

	clitools "github.com/gentoomaniac/backup-tool/pkg/cli"
	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/rs/zerolog/log"
)

type RestoreArgs struct {
	ID              int    `short:"i" help:"ID of the backup to restore"`
	DestinationPath string `help:"path to restore to" argument:"" required:""`
	DBPath          string `short:"d" help:"database file with backup meta information" type:"path" required:""`
}

func restore(params *RestoreArgs) {
	log.Debug().Msg("restore called")
	database, _ := db.NewSQLLite(params.DBPath)

	fmt.Println(cli.Restore.ID)
	backups, err := database.GetBackups()
	if err != nil {
		log.Error().Err(err).Msg("failed getting backups")
		return
	}

	backup, err := clitools.PromptBackups(backups)
	if err != nil {
		log.Error().Err(err).Msg("failed selecting backup")
		return
	}
	log.Debug().Str("name", backup.Name).Str("description", backup.Description).Int("created", backup.Timestamp).Msg("backup selected")

	log.Debug().Msg("Loading FSObjects")
	fsobjects, err := database.GetFSObjForBackups(backup.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed getting fs objects")
		return
	}
	for _, obj := range fsobjects {
		//log.Debug().Str("path", obj.Path).Str("name", obj.Name).Str("mode", obj.FileMode.String()).Msg("")
		if obj.IsDir {
			targetPath := path.Join(params.DestinationPath, obj.Path)
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				os.MkdirAll(targetPath, obj.FileMode)
			}
		}
	}
}

func restoreFSObject() {

}
