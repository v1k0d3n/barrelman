package config

import (
	"fmt"
	"github.com/charter-oss/barrelman/cmd/util"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
	"os"
)

func newConfigUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update <file path>",
		Short: "updates barrelman config",
		Run: func(cmd *cobra.Command, args []string) {
			var barrelmanConfigFile string

			/*If user doesn't provide the config file path then,
			the default location will be used ($USER/.barrelman/config)*/
			if len(args) == 0 {
				barrelmanConfigFile = util.Default().ConfigFile
				fmt.Println("Using barrelman config: ",barrelmanConfigFile)
			} else {
				barrelmanConfigFile = args[0]
			}

			if isUpdateConfig, err := barrelman.UpdateConfig(barrelmanConfigFile); err != nil {
				fmt.Println(isUpdateConfig," Update Failed!")
				os.Exit(1)
			}
			fmt.Printf("Update Success!")
		},
	}
}
