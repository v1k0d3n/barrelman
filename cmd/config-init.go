package cmd

import (
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"sigs.k8s.io/yaml"
)

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{

		Use:   "barrelman config init",
		Short: "Initializes a default Barrelman config file in $USER/.barrelman/config",
		Long: "The config init command will create a new Barrelman config file in the Barrelman home directory with default values. " +
			"After running init, the user can update this file to set global default config for Barrelman.",
		Run: func(cmd *cobra.Command, args []string) {
			defaultConfig := Default().ConfigFile
			if initConfig(defaultConfig) != nil {
				log.Fatal("Error initializing config!")
			}
			log.Print("Config Initialized!")
		},
	}
}

//Initializes barrelman config to a default location specified in defaults.go
func initConfig(configFilePath string) error {

	type Account struct {
		Type   string
		User   string
		Secret string
	}

	type Config struct {
		Accounts map[string]Account
	}

	config := Config{
		Accounts: map[string]Account{
			"default": {
				Type:   "type",
				User:   "user",
				Secret: "secret",
			},
		},
	}

	//c := barrelman.Config{}
	const permissions = 0644

	d, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(configFilePath, d, permissions)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
