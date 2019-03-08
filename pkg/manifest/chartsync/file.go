package chartsync

import (
	"io"
	"os"

	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
)

type SyncFile struct {
	ChartMeta *ChartMeta
	DataDir   string
}

type fileHandler struct {
}

func init() {
	r := &fileHandler{}
	Register(&Registration{
		Name: "file",
		New: func(dataDir string, cm *ChartMeta, acc AccountTable) (Archiver, error) {
			return &SyncFile{
				ChartMeta: cm,
				DataDir:   dataDir,
			}, nil
		},
		Control: r,
	})
}

func (fh *fileHandler) Reset() {
	return
}

func (fh *fileHandler) Sync(cs *ChartSync, acc AccountTable) error {
	return nil
}

func (sf *SyncFile) ArchiveRun(ac *ArchiveConfig) (io.Reader, error) {
	target, err := os.Open(sf.ChartMeta.Source.Location)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"MetaName": sf.ChartMeta.Name,
	}).Debug("creating archive")
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
