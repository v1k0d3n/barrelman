package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/structured/log"
)

func newRootCmd(args []string) *cobra.Command {
	cobraCmd := &cobra.Command{}
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

	cobraCmd.AddCommand(newDeleteCmd(&barrelman.DeleteCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newApplyCmd(&barrelman.ApplyCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newListCmd(&barrelman.ListCmd{
		Options:       options,
		Config:        config,
		ManifestNames: &[]string{},
	}))

	cobraCmd.AddCommand(newTemplateCmd(&barrelman.TemplateCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newVersionCmd(&barrelman.VersionCmd{}))

	flags.Parse(args)

	return cobraCmd
}

func Execute() {
	rootCmd := newRootCmd(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func logSettings(args *[]string) []func(*log.Logger) error {
	ret := []func(*log.Logger) error{}
	for _, v := range *args {
		switch v {
		case "debug", "info", "warn", "error":
			ret = append(ret, log.OptSetLevel(v))
		case "JSON":
			ret = append(ret, log.OptSetJSON())
		}
	}
	return ret
}
