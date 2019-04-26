package manifest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

type ArchiveSpec struct {
	MetaName    string
	ChartName   string
	ReleaseName string
	Path        string
	Reader      io.Reader
	DataDir     string
	Namespace   string
	Overrides   []byte
	InstallWait bool
}

type ArchiveFiles struct {
	List []*ArchiveSpec
}

func Archive(
	chart *Chart,
	path string,
	dependCharts []*chartsync.ChartSpec,
	archiver chartsync.Archiver) (*ArchiveSpec, error) {

	as := &ArchiveSpec{
		MetaName:    chart.Metadata.Name,
		ChartName:   chart.Data.ChartName,
		ReleaseName: chart.Data.ReleaseName,
		Namespace:   chart.Data.Namespace,
		Overrides:   chart.Data.Overrides,
		InstallWait: chart.Data.InstallWait,
	}
	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get absolute path")
	}

	as.Reader, err = archiver.ArchiveRun(&chartsync.ArchiveConfig{
		ChartMeta: &chartsync.ChartMeta{
			Name:    chart.Metadata.Name,
			Source:  chart.Data.SyncSource,
			Depends: chart.Data.Dependencies,
		},
		ArchiveFunc:  createArchive,
		Path:         path,
		DependCharts: dependCharts,
	})

	return as, err
}

//Package creates an archive based on dependancies contained in []*ChartSpec
func Package(depends []*chartsync.ChartSpec, src string, chartMeta *chartsync.ChartMeta) (io.Reader, error) {
	// ensure the src actually exists before trying to tar it

	if chartMeta.Type == "git" {
		if err := chartsync.NewRef(src, chartMeta.Source); err != nil {
			return nil, errors.Wrap(err, "error checking out branch")
		}
	}

	if _, err := os.Stat(src); err != nil {
		return nil, errors.Wrap(err, "unable to tar files")
	}

	buf := bytes.NewBuffer([]byte{})

	gzw := gzip.NewWriter(buf)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add main chart repo
	if err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return errors.Wrap(err, "failed while filepath.Walk()")
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return errors.Wrap(err, "failed while tar.FileInfoHeader()")
		}
		// update the name to correctly reflect the desired destination when untaring
		// k8s expects the chart to be in a subdir
		header.Name = fmt.Sprintf("this/%v", strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator)))
		if header.Name == "" {
			return errors.Wrap(err, "failed constructing header.Name")
		}

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return errors.Wrap(err, "failed while tw.WriteHeader()")
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return errors.WithFields(errors.Fields{"file": file}).Wrap(err, "failed while os.Open(file)")
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return errors.WithFields(errors.Fields{"file": file}).Wrap(err, "failed while io.Copy()")
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	}); err != nil {
		//Error is already annotated
		return nil, err
	}

	//Add depends
	for _, v := range depends {
		if err := filepath.Walk(v.Path, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrap(err, "failed while processing dependencies filepath.Walk()")
			}
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return errors.Wrap(err, "failed while processing dependancies tar.FileInfoHeader()")
			}
			header.Name = fmt.Sprintf("this/charts/%v/%v", v.Name, strings.TrimPrefix(strings.Replace(file, v.Path, "", -1), string(filepath.Separator)))
			if header.Name == "" {
				return errors.Wrap(err, "failed while processing dependencies constructing header.Name")
			}

			if err := tw.WriteHeader(header); err != nil {
				return errors.Wrap(err, "failed while processing dependencies tw.WriteHeader()")
			}

			if !fi.Mode().IsRegular() {
				return nil
			}

			f, err := os.Open(file)
			if err != nil {
				return errors.WithFields(errors.Fields{"file": file}).Wrap(err, "failed while processing dependencies os.Open(file)")
			}
			if _, err := io.Copy(tw, f); err != nil {
				return errors.WithFields(errors.Fields{"file": file}).Wrap(err, "failed while processing dependencies io.Copy()")
			}
			f.Close()
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func createArchive(datadir string, path string, dependCharts []*chartsync.ChartSpec, meta *chartsync.ChartMeta) (io.Reader, error) {
	log.WithFields(log.Fields{
		"Chart": meta.Name,
	}).Debug("creating archive")
	reader, err := Package(dependCharts, path, meta)
	return reader, err
}
