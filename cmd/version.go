package cmd

import (
	"github.com/spf13/cobra"

	"github.com/charter-se/barrelman/pkg/barrelman"
)

func newVersionCmd(cmd *barrelman.VersionCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "version",
		Short: "version something",
		Long:  `Something something else...`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true

			return cmd.Run()
		},
	}
	return cobraCmd
}
