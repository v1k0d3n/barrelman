package chartsync

import (
	"os"

	"github.com/charter-se/structured"
	"github.com/charter-se/structured/errors"
)

type SyncFile struct {
	ChartMeta *ChartMeta
	DataDir   string
	Log       structured.Logger
}

type fileHandler struct {
}

func init() {
	r := &fileHandler{}
	Register(&Registration{
		Name: "file",
		New: func(logger structured.Logger, dataDir string, cm *ChartMeta, acc AccountTable) (Archiver, error) {
			return &SyncFile{
				ChartMeta: cm,
				DataDir:   dataDir,
				Log:       logger,
			}, nil
		},
		Control: r,
	})
}

func (fh *fileHandler) Sync(cs *ChartSync, acc AccountTable) error {
	return nil
}

func (sf *SyncFile) ArchiveRun(ac *ArchiveConfig) (string, error) {
	target := sf.ChartMeta.Source.Location
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	}
	return target, nil
}

func (sf *SyncFile) GetChartMeta() *ChartMeta {
	return sf.ChartMeta
}

func (sf *SyncFile) GetPath() (string, error) {
	target := sf.ChartMeta.Source.Location
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	}
	return target, nil
}
