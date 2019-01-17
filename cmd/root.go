package cmd

import (
	"os"
	"runtime"

	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type cmdOptions struct {
	ManifestFile   string
	ConfigFile     string
	KubeConfigFile string
	KubeContext    string
	DataDir        string
	DryRun         bool
	Diff           bool
	NoSync         bool
	Debug          bool
	InstallRetry   int
	Force          *[]string
}

func newRootCmd(args []string) *cobra.Command {
	cobraCmd := &cobra.Command{}
	options := &cmdOptions{
		DataDir:      Default().DataDir,
		ManifestFile: Default().ManifestFile,
	}
	config := &Config{}

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

	cobraCmd.AddCommand(newDeleteCmd(&deleteCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))

	cobraCmd.AddCommand(newApplyCmd(&applyCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))

	cobraCmd.AddCommand(newListCmd(&listCmd{
		Options:    options,
		Config:     config,
		LogOptions: logOptions,
	}))

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

func logSettings(args *[]string) []func(*log.Logger) error {
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

func ensureWorkDir(datadir string) error {
	return os.MkdirAll(datadir, os.ModePerm)
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
