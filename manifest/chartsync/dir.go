package chartsync

import (
	"os"

	"github.com/charter-se/structured/errors"
)

type SyncDir struct {
	ChartMeta *ChartMeta
	Repo      *dirRepo
	DataDir   string
}

type dirControl struct {
}

type dirRepo string

func init() {
	r := &dirControl{}
	Register(&Registration{
		Name: "dir",
		New: func(dataDir string, cm *ChartMeta, acc AccountTable) (Archiver, error) {
			return &SyncDir{
				ChartMeta: cm,
				DataDir:   dataDir,
			}, nil
		},
		Control: r,
	})
}

func (r *dirControl) Sync(cs *ChartSync, acc AccountTable) error {
	return nil
}

func (g *SyncDir) ArchiveRun(ac *ArchiveConfig) (string, error) {
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
