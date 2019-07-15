package config

import (
	"fmt"
	"github.com/charter-oss/barrelman/cmd/util"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
)

func newConfigUpdateCmd() *cobra.Command {
	var barrelmanConfigFile, configValue string
	const secret = "secret"
	cobraCmd := &cobra.Command{
		Use:   "update [config]",
		Short: "Update barrelman config, When no config is provided barrelman will consider from ~/.barrelman/config",
		Run: func(cobraCmd *cobra.Command, args []string) {

			/*If user doesn't provide the config file path then,
			the default location will be used ($USER/.barrelman/config)*/
			if len(args) == 0 {
				barrelmanConfigFile = util.Default().ConfigFile
				fmt.Println("Using barrelman config: ", barrelmanConfigFile)
			} else {
				barrelmanConfigFile = args[0]
			}

			//Add command line args for secret (--secret)
			secretValue, _ := cobraCmd.Flags().GetString(secret)

			configValue = secretValue

			//Update config secret
			if _, err := barrelman.UpdateConfig(barrelmanConfigFile, configValue); err != nil {
				fmt.Println("Update Failed! ",err)
			}
			fmt.Println("Config updated!")
		},
	}
	cobraCmd.Flags().StringP("secret", "s", "", "--secret")
	return cobraCmd
}
