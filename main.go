package main

import (
	"io/ioutil"
	golog "log"
	"os"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version    = "0.0.1-dev"
	githubSlug = "ahilsend/tink-infrastructure"
)

var cli struct {
	Verbose int   `short:"v" help:"Increase verbosity." type:"counter"`
	Quiet   bool  `short:"q" help:"Do not run upgrades."`
	Json    bool  `help:"Log as json"`
	Regex   regex `help:"Some parameter with custom validator" default:".*"`

	Backup struct {
		BlockSize   int    `short:"b" help:"Data block size in bytes" default:"52428800"`
		BlockPath   string `short:"p" help:"Where to store the blocks" default:"./blocks/"`
		Name        string `required help:"name of the backup"`
		Description string `help:"description for the backup"`
		DBPath      string `short:"d" help:"database file with backup meta information" type:"path"`
		Secret      string `short:"s" help:"secret"`
		Nonce       string `short:"n" help:"IV"`
		Path        string `argument required help:"path to backup"`
	} `cmd help:"Run a backup"`
	Run struct {
	} `cmd help:"Run the application (default)." default:"1" hidden`

	Version kong.VersionFlag `short:"v" help:"Display version."`
}

type regex string

func (r *regex) String() string {
	return string(*r)
}
func (r *regex) Validate() (err error) {
	_, err = regexp.Compile(r.String())
	return err
}

func setupLogging(verbosity int, logJson bool, quiet bool) {
	if !quiet {
		// 1 is zerolog.InfoLevel
		zerolog.SetGlobalLevel(zerolog.Level(1 - verbosity))
		if !logJson {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		golog.SetFlags(0)
		golog.SetOutput(ioutil.Discard)
	}
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
	})
	setupLogging(cli.Verbose, cli.Json, cli.Quiet)

	switch ctx.Command() {
	case "backup":
		backup()

	default:
		log.Info().Msg("Default command")
		log.Debug().Str("regex", cli.Regex.String()).Msg("debug message with extra values")
	}
	ctx.Exit(0)
}
