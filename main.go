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
		BlockSize   int    `short:"b" help:"Data block size in bytes" default:"52428800"`
		BlockPath   string `short:"p" help:"Where to store the blocks" default:"./blocks/"`
		Name        string `help:"name of the backup" required:""`
		Description string `help:"description for the backup"`
		DBPath      string `short:"d" help:"database file with backup meta information" type:"path"`
		Secret      string `short:"s" help:"secret"`
		Nonce       string `short:"n" help:"IV"`
		Path        string `help:"path to backup" argument:"" required:""`
	} `cmd:"" help:"Run a backup"`
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
		backup()

	default:
		log.Info().Msg("Default command")
	}
	ctx.Exit(0)
}
