package cmd

import (
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/spf13/cobra"
)

func newConfigCmd(cmd *barrelman.ConfigCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "config view [config]",
		Short: "config view/ config update",
		Long:  `View or Update barrelman config, When no config is provided barrelman will consider from ~/.barrelman/config`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true

			session := cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)
			if err := cmd.Run(session); err != nil {
				return err
			}
			return nil
		},
	}

	cobraCmd.AddCommand(newConfigInitCmd())

	return cobraCmd
}
