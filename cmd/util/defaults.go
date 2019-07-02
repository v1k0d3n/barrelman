package util

import (
	"fmt"
	"os"
	"os/user"
)

type Defaults struct {
	ManifestFile   string
	KubeConfigFile string
	KubeContext    string
	DataDir        string
	ConfigFile     string
	InstallRetry   int
	Force          *[]string
}

func Default() *Defaults {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	d := &Defaults{}
	d.ConfigFile = fmt.Sprintf("%v/.barrelman/config", usr.HomeDir)
	d.ManifestFile = fmt.Sprintf("%v/.barrelman/manifest.yaml", usr.HomeDir)
	d.KubeConfigFile = os.Getenv("KUBE_CONFIG")
	if d.KubeConfigFile == "" {
		d.KubeConfigFile = fmt.Sprintf("%v/.kube/config", usr.HomeDir)
	}
	d.KubeContext = os.Getenv("KUBE_CONTEXT")
	d.DataDir = fmt.Sprintf("%v/.barrelman/data", usr.HomeDir)
	d.InstallRetry = int(3)
	d.Force = &[]string{}
	return d
}
