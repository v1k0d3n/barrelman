package chartsync

import (
	"os"

	"github.com/charter-se/structured"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
)

type SyncDir struct {
	ChartMeta *ChartMeta
	Repo      *dirRepo
	DataDir   string
	Log       structured.Logger
}

type dirControl struct {
}

type dirRepo string

func init() {
	r := &dirControl{}
	Register(&Registration{
		Name: "dir",
		New: func(logger structured.Logger, dataDir string, cm *ChartMeta, acc AccountTable) (Archiver, error) {
			return &SyncDir{
				ChartMeta: cm,
				DataDir:   dataDir,
				Log:       logger,
			}, nil
		},
		Control: r,
	})
}

func (r *dirControl) Sync(cs *ChartSync, acc AccountTable) error {
	return nil
}

func (g *SyncDir) ArchiveRun(ac *ArchiveConfig) (string, error) {
	g.Log.WithFields(log.Fields{
		"DataDir":     ac.DataDir,
		"AcrhivePath": ac.Path,
	}).Debug("Dir handler running archiveFunc")
	return ac.ArchiveFunc(ac.DataDir, ac.Path, ac.DependCharts)
}

func (g *SyncDir) GetChartMeta() *ChartMeta {
	return g.ChartMeta
}

func (g *SyncDir) GetPath() (string, error) {
	target := g.ChartMeta.Source.Location
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	}
	return target, nil
}
