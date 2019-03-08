package chartsync

import (
	"io"
	"os"

	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
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

func (r *dirControl) Reset() {
	return
}

func (r *dirControl) Sync(cs *ChartSync, acc AccountTable) error {
	return nil
}

func (g *SyncDir) ArchiveRun(ac *ArchiveConfig) (io.Reader, error) {
	log.WithFields(log.Fields{
		"DataDir":     ac.DataDir,
		"AcrhivePath": ac.Path,
	}).Debug("Dir handler running archiveFunc")
	return ac.ArchiveFunc(ac.DataDir, ac.Path, ac.DependCharts, ac.ChartMeta)
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
