package version

var (
	version = "UNSET"
	branch  = "UNSET"
	commit  = "UNSET"
)

type Version struct {
	Version string
	Branch  string
	Commit  string
}

func Get() *Version {
	return &Version{
		Version: version,
		Branch:  branch,
		Commit:  commit,
	}
}

func (ver *Version) ShortReport() map[string]interface{} {
	return map[string]interface{}{
		"Version": ver.Version,
		"Branch":  ver.Branch,
		"Commit":  ver.Commit,
	}
}

func (ver *Version) DetailedReport() map[string]interface{} {
	return ver.ShortReport()
}
