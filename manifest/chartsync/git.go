package chartsync

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/charter-se/structured"
	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type SyncGit struct {
	ChartMeta *ChartMeta
	Repo      *gitRepo
	DataDir   string
	Log       structured.Logger
}

type repoList map[string]gitRepo
type gitRepoList struct {
	sync.RWMutex
	list repoList
}

type gitRepo string

func init() {
	r := &gitRepoList{list: make(repoList)}
	Register(&Registration{
		Name: "git",
		New: func(logger structured.Logger, dataDir string, cm *ChartMeta, acc AccountTable) (Archiver, error) {
			uri, err := cm.GetURI()
			if err != nil {
				return nil, errors.WithFields(errors.Fields{
					"Name": cm.Name,
				}).Wrap(err, "git module failed to parse chart Location")
			}
			if _, ok := r.list[uri]; !ok {
				r.list[uri] = gitRepo(uri)
			}
			return &SyncGit{
				ChartMeta: cm,
				DataDir:   dataDir,
				Log:       logger,
			}, nil
		},
		Control: r,
	})
}

func (r *gitRepoList) Sync(cs *ChartSync, acc AccountTable) error {
	r.Lock()
	defer func() {
		r.Unlock()
	}()
	for k := range r.list {
		if err := r.Download(cs, acc, k); err != nil {
			return err
		}
	}
	return nil
}

func (g *SyncGit) ArchiveRun(ac *ArchiveConfig) (string, error) {
	g.Log.WithFields(log.Fields{
		"DataDir":     ac.DataDir,
		"AcrhivePath": ac.Path,
	}).Debug("Git handler running archiveFunc")
	return ac.ArchiveFunc(ac.DataDir, ac.Path, ac.DependCharts)
}

func (g *SyncGit) GetChartMeta() *ChartMeta {
	return g.ChartMeta
}

func (g *SyncGit) GetPath() (string, error) {
	u, err := url.Parse(g.ChartMeta.Source.Location)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("%v/%v%v/%v", g.DataDir, u.Host, u.Path, g.ChartMeta.Source.SubPath)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	}
	return target, nil
}

func (r *gitRepoList) Download(cs *ChartSync, acc AccountTable, location string) error {
	u, err := url.Parse(location)
	if err != nil {
		return err
	}

	cloneOptions := &git.CloneOptions{
		URL:      location,
		Progress: os.Stdout,
	}
	pullOptions := &git.PullOptions{
		RemoteName:   "origin",
		SingleBranch: true,
		Progress:     os.Stdout,
	}

	if v, exists := acc[u.Host]; exists {
		cloneOptions.Auth = &http.BasicAuth{
			Username: v.User,
			Password: v.Secret,
		}
		pullOptions.Auth = &http.BasicAuth{
			Username: v.User,
			Password: v.Secret,
		}
	}

	target := fmt.Sprintf("%v/%v/%v", cs.DataDir, u.Host, u.Path)
	for _, v := range cs.CompletedURI {
		if v == target {
			//We already downloaded this one
			return nil
		}
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		_, err = git.PlainClone(target, false, cloneOptions)
		if err != nil {
			if cloneOptions.Auth != nil {
				return errors.WithFields(errors.Fields{
					"Repository": cloneOptions.URL,
					"AuthName":   cloneOptions.Auth.Name(),
				}).Wrap(err, "could not clone via git")
			}
			return errors.WithFields(errors.Fields{
				"Repository": cloneOptions.URL,
			}).Wrap(err, "could not clone via git")
		}
	} else {
		d, err := git.PlainOpen(target)
		if err != nil {
			return errors.WithFields(errors.Fields{
				"LocalRepository": target,
			}).Wrap(err, "could not open local repository")
		}
		wt, err := d.Worktree()
		if err != nil {
			return errors.Wrap(err, "could not create working tree")
		}
		err = wt.Pull(pullOptions)
		if err != nil {
			if err != git.NoErrAlreadyUpToDate {
				if cloneOptions.Auth != nil {
					return errors.WithFields(errors.Fields{
						"Repository": cloneOptions.URL,
						"AuthName":   cloneOptions.Auth.Name(),
					}).Wrap(err, "could not pull from repository")
				}
				return errors.WithFields(errors.Fields{
					"Repository": cloneOptions.URL,
				}).Wrap(err, "could not pull from repository")
			}
		}
	}
	cs.CompletedURI = append(cs.CompletedURI, target)
	return nil
}
