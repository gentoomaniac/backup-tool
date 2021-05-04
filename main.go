package main

import (
	"github.com/alecthomas/kong"
	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/gentoomaniac/logging"
	"github.com/rs/zerolog/log"
)

var (
	version = "unset"
	commit  = "unset"
	binName = "backup-tool"
	builtBy = "manual"
	date    = "unset"
)

var cli struct {
	logging.LoggingConfig

	Backup `cmd:"" embed:"" help:"Run a backup"`

	Restore struct {
		Restore `embed:""`
	} `cmd:"" help:"Run a restore"`

	Run struct {
	} `cmd:"" help:"Run the application (default)." default:"1" hidden:""`

	Version kong.VersionFlag `short:"v" help:"Display version."`
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})
	logging.Setup(&cli.LoggingConfig)

	database, _ := db.NewSQLLite(cli.DBPath)

	switch ctx.Command() {
	case "backup":
		backup(database, &cli.Backup)

	default:
		log.Info().Msg("Default command")
	}
	ctx.Exit(0)
}
