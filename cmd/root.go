package cmd

import (
	"os"
	"runtime"

	"github.com/charter-se/structured/log"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
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

	cobraCmd.AddCommand(newDeleteCmd(&deleteCmd{
		Options: options,
		Config:  config,
	}))

	cobraCmd.AddCommand(newApplyCmd(&applyCmd{
		Options: options,
		Config:  config,
	}))

	flags.Parse(args)
	return cobraCmd
}

func Execute() {
	rootCmd := newRootCmd(os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
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
