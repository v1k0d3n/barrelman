package cmd

import (
	"fmt"
	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"sigs.k8s.io/yaml"
)

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{

		Use:   "init",
		Short: "Initializes barrelman config, creates default config under $USER/.barrelman/config",
		Run: func(cmd *cobra.Command, args []string) {

			defaultConfig := Default().ConfigFile
			if initConfig(defaultConfig) != nil {
				fmt.Println("Error initializing config!")
			}
			fmt.Println("Config Initialized!")
		},
	}
}

//Initializes barrelman config to a default location specified in defaults.go
func initConfig(configFilePath string) error {
	c := barrelman.Config{}
	const permissions = 0644
	var initData = `
account:
  default:
    typ: type
    user: test
    secret: test
`
	err := yaml.Unmarshal([]byte(initData), &c)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(configFilePath, d, permissions)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
