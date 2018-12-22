//go:generate mockery -name=Archiver
package chartsync

import (
	"fmt"
	"net/url"
	"os"

	"github.com/charter-se/structured/errors"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type AccountTable map[string]*Account

type ArchiveConfig struct {
	ChartMeta    *ChartMeta
	ArchiveFunc  func(string, string, []*ChartSpec) (string, error)
	DataDir      string
	Path         string
	DependCharts []*ChartSpec
}

type Account struct {
	Typ    string
	User   string
	Secret string
}

type ChartSync struct {
	Charts       []*ChartMeta
	DataDir      string
	CompletedURI []string
	AccountTable AccountTable
}

type ChartSpec struct {
	Name string
	Path string
}

type Source struct {
	Location  string
	SubPath   string
	Reference string
}
type ChartMeta struct {
	Name    string
	Source  *Source
	Type    string
	Depends []string
	Groups  []string
}

type Controller interface {
	Syncer
}

type Syncer interface {
	Sync(*ChartSync, AccountTable) error
}

type Archiver interface {
	ArchiveRun(*ArchiveConfig) (string, error)
	GetPath() (string, error)
}

//Charter implements chart functions as per standard naming conventions
//any resemblance to anything else is purely coincidental
type Charter interface {
	ChartMeta() *ChartMeta
}

func New(d string, acc AccountTable) *ChartSync {
	return &ChartSync{
		DataDir:      d,
		AccountTable: acc,
	}
}

func (cs *ChartSync) Sync(acc AccountTable) error {
	for _, control := range registry.AllControllers() {
		if err := control.Sync(cs, acc); err != nil {
			return err
		}
	}
	return nil
}

func (c *ChartMeta) GetURI() (string, error) {
	return c.Source.Location, nil
}

func GetControl(s string) (Controller, error) {
	if registration, ok := registry.Lookup(s); ok {
		return registration.Control, nil
	}
	return nil, errors.WithFields(errors.Fields{
		"SourceType": s,
	}).New("failed to find handler for source")
}

func GetHandler(s string) (*Registration, error) {
	if registration, ok := registry.Lookup(s); ok {
		return registration, nil
	}
	return nil, errors.WithFields(errors.Fields{
		"SourceType": s,
	}).New("failed to find handler for source")
}

func (cs *ChartSync) Add(c *ChartMeta) error {
	cs.Charts = append(cs.Charts, c)
	return nil
}

func (cs *ChartSync) gitDownload(c *ChartMeta, acc AccountTable) error {
	u, err := url.Parse(c.Source.Location)
	if err != nil {
		return err
	}

	cloneOptions := &git.CloneOptions{
		URL:      c.Source.Location,
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
