package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Account struct {
	Secret string `yaml:"Secret"`
	Type   string `yaml:"Type"`
	User   string `yaml:"User"`
}
type Accounts map[string]Account

type Config struct {
	Accounts
}

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Display Barrelman config",
		Run: func(cmd *cobra.Command, args []string) {

			defaultConfigPath := Default().ConfigFile
			c := Config{}
			yamlFile, err := ioutil.ReadFile(defaultConfigPath)
			err = yaml.Unmarshal(yamlFile, &c)
			if err != nil {
				log.Fatalf("Unmarshal: %v", err)
			}
			b, _ := yaml.Marshal(c)
			fmt.Print(string(b))
		},
	}
}
