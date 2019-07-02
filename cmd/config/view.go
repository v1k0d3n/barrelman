package config

import (
	"fmt"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/structured/log"
	"github.com/spf13/cobra"
)

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "view barrelman config",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := barrelman.GetConfigFromFile("/home/armorking/.barrelman/config")
			if err != nil {
				log.Error(err)
			}
			fmt.Print("Reading config", config)
		},
	}
}
