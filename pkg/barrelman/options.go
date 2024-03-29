package barrelman

type CmdOptions struct {
	ManifestFile   string
	ConfigFile     string
	KubeConfigFile string
	KubeContext    string
	DataDir        string
	LogLevel       string
	DryRun         bool
	Diff           bool
	NoSync         bool
	Debug          bool
	Force          *[]string
	InstallRetry   int
	InstallWait    bool
}
