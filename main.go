package main

import (
	"github.com/alecthomas/kong"
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

	Backup struct {
		BackupArgs `embed:""`
	} `cmd:"" help:"Run a backup"`

	Restore struct {
		RestoreArgs `embed:""`
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

	switch ctx.Command() {
	case "backup":
		err := backup(&cli.Backup.BackupArgs)
		if err != nil {
			log.Error().Err(err).Msg("oops!")
		}

	case "restore":
		restore(&cli.Restore.RestoreArgs)

	default:
		log.Info().Msg("Default command")
	}
	ctx.Exit(0)
}
