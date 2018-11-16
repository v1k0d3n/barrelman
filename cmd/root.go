package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var manifestFile string
var debug bool

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug mode enables verbose and stack trace on error")
}

var rootCmd = &cobra.Command{
	Use:   "barrelman",
	Short: "barrelman is an Armada compatible Helm plugin",
	Long:  `Something something else...`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

//Execute as per https://github.com/spf13/cobra
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
