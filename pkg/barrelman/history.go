package barrelman

import (
	"strconv"

	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
	"k8s.io/helm/pkg/timeconv"
)

type HistoryCmd struct {
	Options         *CmdOptions
	Config          *Config
	ManifestName    string
	ManifestVersion int32
	LogOptions      *[]string
}

type HistoryTarget struct {
	ReleaseMeta *cluster.ReleaseMeta
	Chart       *cluster.Chart
	State       int
	Diff        []byte
	Changed     bool
}

type HistoryTargets struct {
	ManifestName string
	Data         []*ReleaseTarget
}

func (cmd *HistoryCmd) Run(session cluster.Sessioner) error {
	var err error
	log.Rep(version.Get()).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	log.Rep(session).Debug("connecting to cluster")
	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	versions, err := session.GetVersions(cmd.ManifestName)
	if err != nil {
		return errors.Wrap(err, "Failed to get versions")
	}
	for _, v := range versions.Data {
		log.WithFields(log.Fields{
			"manifestName": v.Name,
			"namespace":    v.Namespace,
			"Revision":     v.Revision,
		}).Info("Rollback manifest")
	}

	cluster.By(revisionNumber).Sort(versions.Data)
	for _, v := range versions.Data {
		log.WithFields(log.Fields{
			"ReleaseName":  v.Name,
			"Revision":     strconv.Itoa(int(v.Revision)),
			"LastDeployed": timeconv.String(v.Info.GetLastDeployed()),
		}).Info("history")
	}
	return nil
}

// revisionNumber is a sorting algorythm for sorting []*cluster.Version
func revisionNumber(r1, r2 *cluster.Version) bool {
	return r1.Revision < r2.Revision
}
