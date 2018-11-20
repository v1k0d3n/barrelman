package cmd

import (
	"os"

	"github.com/charter-se/structured/log"

	"github.com/spf13/cobra"
)

type cmdOptions struct {
	ManifestFile string
	ConfigFile   string
	DataDir      string
	Debug        bool
}

func newRootCmd(args []string) *cobra.Command {
	cobraCmd := &cobra.Command{}
	options := &cmdOptions{
		DataDir:      Default().DataDir,
		ManifestFile: Default().ManifestFile,
	}
	config := &Config{}

	flags := cobraCmd.PersistentFlags()
	flags.StringVarP(
		&options.ConfigFile,
		"config",
		"c",
		Default().ConfigFile,
		"specify manifest (YAML) file or a URL")

	cobraCmd.AddCommand(newDeleteCmd(&deleteCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newInstallCmd(&installCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newUpgradeCmd(&upgradeCmd{
		Options: options,
		Config:  config,
	}))

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
