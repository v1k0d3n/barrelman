package cmd

import (
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/barrelman/pkg/cluster"
)

func newListCmd(cmd *barrelman.ListCmd) *cobra.Command {

	example := strings.TrimSpace(dedent.Dedent(`
		barrelman list 
		barrelman list lamp-stack`))

	longDesc := strings.TrimSpace(dedent.Dedent(`
		Display Barrelman manifests deployed in the kuerbenetes cluster.
		Manifest names can be used with the rollback command.`))

	shortDesc := `list Barrelman manifests and releases`

	cobraCmd := &cobra.Command{
		Use:           "list [manifest name]",
		Short:         shortDesc,
		Long:          longDesc,
		Example:       example,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.ManifestName = args[0]
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

	f := cobraCmd.Flags()
	cmd.LogOptions = f.StringSliceP(
		"log",
		"l",
		nil,
		"log options (e.g. --log=debug,JSON")
	return cobraCmd
}
