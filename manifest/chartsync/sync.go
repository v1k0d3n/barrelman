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
	Control Controller
}

type Controller interface {
	Syncer
}
type Syncer interface {
	Sync() error
}

type Charter interface {
	ChartMeta() *ChartMeta
}

func New(d string, acc AccountTable) *ChartSync {
	return &ChartSync{
		DataDir:      d,
		AccountTable: acc,
	}
}

// func (cs *ChartSync) Sync() error {
// 	for _, v := range cs.Charts {
// 		if err := v.Control.Sync(); err != nil {
// 			return errors.WithFields(errors.Fields{"Location": v.Source.Location}).Wrap(err, "error doing git download")
// 		}
// 	}
// 	return nil
// }

func (cs *ChartSync) Sync() error {
	for _, control := range registry.AllControllers() {
		if err := control.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ChartMeta) GetURI() (string, error) {
	u, err := url.Parse(c.Source.Location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v/%v", u.Host, u.Path), nil
}

func (cs *ChartSync) GetPath(c *ChartMeta) (string, error) {
	var target string

	switch c.Type {
	case "git":
		u, err := url.Parse(c.Source.Location)
		if err != nil {
			return "", err
		}
		target = fmt.Sprintf("%v/%v%v/%v", cs.DataDir, u.Host, u.Path, c.Source.SubPath)
	case "file":
		target = c.Source.Location
	case "dir":
		target = c.Source.Location
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	}
	return target, nil
}

func (cs *ChartSync) Add(c *ChartMeta) error {
	fmt.Printf("Adding %v\n", c.Name)
	if registration, ok := registry.Lookup(c.Type); ok {
		fmt.Printf("Registering %v as %v\n", c.Name, registration.Name)
		c.Control = registration.Control
	} else {
		return errors.WithFields(errors.Fields{
			"Name":       c.Name,
			"SourceType": c.Type,
		}).New("failed to find handler for source")
	}
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
