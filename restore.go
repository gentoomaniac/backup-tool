package main

import (
	"fmt"

	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/rs/zerolog/log"
)

type Restore struct {
	ID     int    `short:"i" help:"ID of the backup to restore"`
	Path   string `help:"path to backup" argument:"" required:""`
	DBPath string `short:"d" help:"database file with backup meta information" type:"path"`
}

func restore(params *Restore) {
	log.Debug().Msg("restore called")
	database, _ := db.NewSQLLite(params.DBPath)

	fmt.Println(cli.Restore.ID)
	backup, _ := database.GetBackupBackupById(params.ID)

	log.Printf("Backup: %s - %s - %d\n", backup.Name, backup.Description, backup.Timestamp)
}
