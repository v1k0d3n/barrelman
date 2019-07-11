package config

import (
	"fmt"
	"github.com/charter-oss/barrelman/cmd/util"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
)

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view [config]",
		Short: "View barrelman config, When no config is provided barrelman will consider from ~/.barrelman/config",
		Run: func(cmd *cobra.Command, args []string) {
			var barrelmanConfigFile string
			const accountType string = "github.com"

			/*If user doesn't provide the config file path then,
			the default location will be used ($USER/.barrelman/config)*/
			if len(args) == 0 {
				barrelmanConfigFile = util.Default().ConfigFile
			} else {
				barrelmanConfigFile = args[0]
			}
			config, _ := barrelman.GetConfigFromFile(barrelmanConfigFile)
			accountMap := config.Account
			account := accountMap[accountType]
			fmt.Println("User: "+account.User, "\n", "type: ", account.Typ, "\n", "Secret: "+account.Secret)
		},
	}
}
