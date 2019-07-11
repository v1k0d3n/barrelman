package config

import (
	"fmt"
	"github.com/charter-oss/barrelman/cmd/util"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
	"os"
)

func newConfigUpdateCmd() *cobra.Command {
	var barrelmanConfigFile string
	var configKey string
	var configValue string
	const secret = "secret"
	const user = "user"
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
			configKey = secret
			configValue = secretValue

			//Update config as per provided arguments (secret/user)
			if isUpdateConfig, err := barrelman.UpdateConfig(barrelmanConfigFile, configKey, configValue); err != nil {
				fmt.Println(isUpdateConfig, " Update Failed!")
				os.Exit(1)
			}
			fmt.Printf("Update Success!")
		},
	}
	cobraCmd.Flags().StringP("secret", "s", "", "--secret")
	cobraCmd.Flags().StringP("user", "u", "", "--user")
	return cobraCmd
}
