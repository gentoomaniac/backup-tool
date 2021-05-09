package main

import (
	"fmt"
	"io/ioutil"
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

	log.Debug().Msg("recreating folder structure ...")
	for _, obj := range fsobjects {
		targetPath := path.Join(params.DestinationPath, obj.Path, obj.Name)
		if obj.IsDir {
			log.Debug().Str("path", obj.String()).Msg("creating directory")
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				err = os.MkdirAll(targetPath, obj.FileMode)
				if err != nil {
					log.Error().Err(err).Msg("failed creating directory")
					return
				}
			}
		}
	}

	for _, obj := range fsobjects {
		if !obj.IsDir {

			err = restoreFSObject(obj, params.DestinationPath)
			if err != nil {
				log.Error().Err(err).Str("file", obj.String()).Msg("restoring object failed")
			}
		}
	}

	log.Debug().Msg("restoring files ...")
}

func restoreFSObject(obj *db.FSObject, destination string) error {
	log.Debug().Str("file", obj.String()).Msg("restoring file")
	var data []byte

	targetPath := path.Join(destination, obj.Path, obj.Name)

	err := ioutil.WriteFile(targetPath, data, 0644)
	return err
}
