package cmd

import (
	config2 "github.com/charter-oss/barrelman/cmd/config"
	"os"

	"github.com/spf13/cobra"

	"github.com/charter-oss/barrelman/pkg/barrelman"
	"github.com/charter-oss/structured/log"
)

func newRootCmd(args []string) *cobra.Command {
	cobraCmd := &cobra.Command{}
	options := &barrelman.CmdOptions{
		DataDir:      Default().DataDir,
		ManifestFile: Default().ManifestFile,
	}
	config := &barrelman.Config{}

	flags := cobraCmd.PersistentFlags()
	flags.StringVarP(
		&options.ConfigFile,
		"config",
		"c",
		Default().ConfigFile,
		"specify manifest (YAML) file or a URL")

	logOptions := &[]string{}
	tmpLogOptions := flags.StringSliceP(
		"log",
		"l",
		nil,
		"log options (e.g. --log=debug,JSON")

	cobraCmd.AddCommand(newDeleteCmd(&barrelman.DeleteCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))

	cobraCmd.AddCommand(config2.NewCmdConfig(out))

	cobraCmd.AddCommand(newApplyCmd(&barrelman.ApplyCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))

	cobraCmd.AddCommand(newListCmd(&barrelman.ListCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))
	cobraCmd.AddCommand(newTemplateCmd(&barrelman.TemplateCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))

	cobraCmd.AddCommand(newVersionCmd(&barrelman.VersionCmd{}))

	flags.Parse(args)
	//We are triggering Cobra to set that value twice somewhere
	//This snapshots the values before we pass them to the command
	*logOptions = *tmpLogOptions
	return cobraCmd
}

func Execute() {
	rootCmd := newRootCmd(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func LogSettings(args *[]string) []func(*log.Logger) error {
	ret := []func(*log.Logger) error{}
	for _, v := range *args {
		switch v {
		case "debug", "info", "warn", "error":
			ret = append(ret, log.OptSetLevel(v))
		case "JSON":
			ret = append(ret, log.OptSetJSON())
		}
	}
	return ret
}
