package cmd

import (
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

func newHistoryCmd(cmd *barrelman.HistoryCmd) *cobra.Command {

	longDesc := strings.TrimSpace(dedent.Dedent(`
		Display Barrelman manifest revision history.
		A new manifest revision is created upon sucessful change is performed on the cluster.
	`))

	shortDesc := `list manifest revision history`

	examples := `barrelman history lamp-stack`

	cobraCmd := &cobra.Command{
		Use:     "history [manifest name]",
		Short:   shortDesc,
		Long:    longDesc,
		Example: examples,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("history requires 'manifest name'")
			}

			cmd.ManifestName = args[0]

			cobraCmd.SilenceUsage = true
			cobraCmd.SilenceErrors = true
			log.Configure(logSettings(cmd.LogOptions)...)
			session := cluster.NewSession(
				cmd.Options.KubeContext,
				cmd.Options.KubeConfigFile)
			if err := cmd.Run(session); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.LogOptions = cobraCmd.Flags().StringSliceP(
		"log",
		"l",
		nil,
		"log options (e.g. --log=debug,JSON")

	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeConfigFile,
		"kubeconfig",
		Default().KubeConfigFile,
		"use alternate kube config file")
	cobraCmd.Flags().StringVar(
		&cmd.Options.KubeContext,
		"kubecontext",
		Default().KubeContext,
		"use alternate kube context")
	return cobraCmd
}
