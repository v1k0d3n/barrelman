package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/cirrocloud/structured/log"
)

type LogOpts struct {
	Level  string
	Format string
}

type RootOpts struct {
	LogOpts *LogOpts
}

func newRootCmd(args []string) (*cobra.Command, *RootOpts) {

	longDesc := strings.TrimSpace(dedent.Dedent(`
		Barrelman uses a single manifest to organize complex application deployments that can consist 
		of many microservices and independent shared services such as databases and caches.

		Barrelman does diff analysis on each release and only executes those changes necessary to achieve 
		the desired state.
		
		Additionally, Helm charts can be sourced from different locations like local file, directory, 
		GitHub repos, Helm repos, etc. This makes Barrelman manifests very flexible. 
	`))

	shortDesc := `Deploys groups of kubernetes releases from a manifest.`

	examples := `barrelman help`

	cobraCmd := &cobra.Command{
		Use:           "barrelman",
		Short:         shortDesc,
		Long:          longDesc,
		Example:       examples,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	options := &barrelman.CmdOptions{
		DataDir:      Default().DataDir,
		ManifestFile: Default().ManifestFile,
	}
	config := &barrelman.Config{}

	rootOpts := &RootOpts{
		LogOpts: &LogOpts{},
	}

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

	flags.StringVar(
		&options.KubeConfigFile,
		"kubeconfig",
		Default().KubeConfigFile,
		"use alternate kube config file")

	flags.StringVar(
		&options.KubeContext,
		"kubecontext",
		Default().KubeContext,
		"use alternate kube context")

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

	cobraCmd.AddCommand(newRollbackCmd(&barrelman.RollbackCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newHistoryCmd(&barrelman.HistoryCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newDescribeCmd(&barrelman.DescribeCmd{
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
