package cmd

import (
	"os"
	"runtime"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/manifest"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type installCmd struct {
	Options *cmdOptions
	Config  *Config
}

func newInstallCmd(cmd *installCmd) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "install [manifest.yaml]",
		Short: "install something",
		Long:  `Something something else...`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cmd.Options.ManifestFile = args[0]
			}
			if err := runInstallCmd(cmd); err != nil {
				log.Error(err)
				os.Exit(1)
			}
		},
	}
	return cobraCmd
}

func runInstallCmd(cmd *installCmd) error {
	log.Warn("Barrelman Install Engage!")

	var err error
	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}
	log.WithFields(log.Fields{"file": cmd.Options.ConfigFile}).Info("Using config")

	// Open connections to the k8s APIs
	c, err := cluster.NewSession(Default().KubeConfigFile)
	if err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	log.Info("syncronizing with remote chart repositories")
	if err := mfest.Sync(cmd.Config.Account); err != nil {
		errors.Wrap(err, "error while downloading charts")
	}

	archives, err := mfest.CreateArchives()
	if err != nil {
		return errors.Wrap(err, "failed to create archives")
	}

	//Remove archive files after we are done with them
	defer func() {
		if err := archives.Purge(); err != nil {
			log.Error(errors.Wrap(err, "failed to purge local archives"))
		}
	}()

	for _, v := range archives.List {
		//Install the release from the tgz above
		relName, err := c.InstallRelease(&cluster.ReleaseMeta{
			Path:      v.Path,
			Namespace: v.Namespace,
		}, []byte{})
		if err != nil {
			return errors.WithFields(errors.Fields{
				"Name":      v.Name,
				"Namespace": v.Namespace,
			}).Wrap(err, "error while installing release")
		}
		log.WithFields(log.Fields{
			"Name":      v.Name,
			"Namespace": v.Namespace,
			"Release":   relName,
		}).Info("installed release")
	}
	return nil
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
