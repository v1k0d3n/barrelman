package cmd

import (
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
)

func newConfigCmd(cmd *barrelman.ConfigCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "config view",
		Short: "View the active Barrelman config file.",
		Long:  `View the active Barrelman config file.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			defaultConfig := Default().ConfigFile
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true

			cmd.Run(defaultConfig)
			return nil
		},
	}

	cobraCmd.AddCommand(newConfigInitCmd())

	return cobraCmd
}
