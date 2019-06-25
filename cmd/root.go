package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/structured/log"
)

type LogOpts struct {
	Level  string
	Format string
}

type RootOpts struct {
	LogOpts *LogOpts
}

func newRootCmd(args []string) (*cobra.Command, *RootOpts) {
	cobraCmd := &cobra.Command{}
	rootOpts := &RootOpts{
		LogOpts: &LogOpts{},
	}

	options := &barrelman.CmdOptions{
		DataDir:      Default().DataDir,
		ManifestFile: Default().ManifestFile,
	}
	config := &barrelman.Config{}

	flags := cobraCmd.PersistentFlags()
	flags.StringVarP(
		&options.ConfigFile,
		"config",
		"c",
		Default().ConfigFile,
		"specify manifest (YAML) file or a URL")

	flags.StringVar(
		&rootOpts.LogOpts.Level,
		"log-level",
		Default().LogLevel,
		"Set the log level. [ debug | info | warn | error ]")

	flags.StringVar(
		&rootOpts.LogOpts.Format,
		"log-format",
		Default().LogFormat,
		"Set the log format. [ text | json ]")

	cobraCmd.AddCommand(newDeleteCmd(&barrelman.DeleteCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newApplyCmd(&barrelman.ApplyCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newListCmd(&barrelman.ListCmd{
		Options: options,
		Config:  config,
	}))
	cobraCmd.AddCommand(newTemplateCmd(&barrelman.TemplateCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newVersionCmd(&barrelman.VersionCmd{}))

	flags.Parse(args)
	return cobraCmd, rootOpts
}

func Execute() {
	rootCmd, rootOpts := newRootCmd(os.Args[1:])
	log.Configure(rootOpts.logSettings()...)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func (rootOpts *RootOpts) logSettings() []func(*log.Logger) error {
	settings := []func(*log.Logger) error{}
	// Set log level
	settings = append(settings, log.OptSetLevel(rootOpts.LogOpts.Level))

	// If log-format has JSON set, configure for JSON
	switch rootOpts.LogOpts.Format {
	case "json":
		settings = append(settings, log.OptSetJSON())
	case "text":
	default:
		log.Error(errors.New("--log-format must be [text | json]"))
		os.Exit(1)
	}
	return settings
}
