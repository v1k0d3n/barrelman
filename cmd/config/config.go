package config

import (
	"github.com/charter-oss/barrelman/cmd/util"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/structured/log"
	"github.com/spf13/cobra"
)

func NewConfigCmd(cmd *barrelman.ConfigCmd) *cobra.Command {
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
			log.Configure(util.LogSettings(cmd.LogOptions)...)
			session := cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)
			if err := cmd.Run(session); err != nil {
				return err
			}
			return nil
		},
	}

	cobraCmd.AddCommand(newConfigViewCmd())
	cobraCmd.AddCommand(newConfigUpdateCmd())

	return cobraCmd
}
