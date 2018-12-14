package chartsync

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	//"github.com/charter-se/barrelman/manifest/sourcetype"
	"github.com/charter-se/structured/errors"
)

type Git struct {
	ChartMeta *ChartMeta
	Repo      *gitRepo
	DataDir   string
}

type repoList map[string]bool
type gitRepo struct {
	sync.RWMutex
	list repoList
}

func init() {
	r := &gitRepo{list: make(repoList)}
	Register(&Registration{
		Name: "git",
		New: func(dataDir string, cm *ChartMeta) (interface{}, error) {
			uri, err := cm.GetURI()
			if err != nil {
				return nil, errors.WithFields(errors.Fields{
					"Name": cm.Name,
				}).Wrap(err, "git module failed to parse chart Location")
			}
			if ok := r.list[uri]; !ok {
				r.list[uri] = true
			}
			return &Git{
				ChartMeta: cm,
				DataDir:   dataDir,
			}, nil
		},
		Control: r,
	})
}

func (r *gitRepo) Sync() error {
	fmt.Printf("git.Sync() called\n")
	return nil
}

func (g *Git) Sync() error {
	fmt.Printf("git.Sync() called\n")
	// if err := g.gitDownload(); err != nil {
	// 	return errors.WithFields(errors.Fields{"Location": g.ChartMeta.Location}).Wrap(err, "error doing git download")
	// }
	return nil
}

func (g *Git) GetChartMeta() *ChartMeta {
	return g.ChartMeta
}

func (g *Git) GetPath() (string, error) {
	var target string

	u, err := url.Parse(g.ChartMeta.Source.Location)
	if err != nil {
		return "", err
	}
	target = fmt.Sprintf("%v/%v%v/%v", g.DataDir, u.Host, u.Path, g.ChartMeta.Source.SubPath)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	}

	return target, nil
}

// func (g *Git) gitDownload() error {
// 	u, err := url.Parse(c.Location)
// 	if err != nil {
// 		return err
// 	}

// 	cloneOptions := &git.CloneOptions{
// 		URL:      c.Location,
// 		Progress: os.Stdout,
// 	}
// 	pullOptions := &git.PullOptions{
// 		RemoteName:   "origin",
// 		SingleBranch: true,
// 		Progress:     os.Stdout,
// 	}

// 	if v, exists := acc[u.Host]; exists {
// 		cloneOptions.Auth = &http.BasicAuth{
// 			Username: v.User,
// 			Password: v.Secret,
// 		}
// 		pullOptions.Auth = &http.BasicAuth{
// 			Username: v.User,
// 			Password: v.Secret,
// 		}
// 	}

// 	target := fmt.Sprintf("%v/%v/%v", cs.Library, u.Host, u.Path)
// 	for _, v := range cs.CompletedURI {
// 		if v == target {
// 			//We already downloaded this one
// 			return nil
// 		}
// 	}

// 	if _, err := os.Stat(target); os.IsNotExist(err) {
// 		_, err = git.PlainClone(target, false, cloneOptions)
// 		if err != nil {
// 			if cloneOptions.Auth != nil {
// 				return errors.WithFields(errors.Fields{
// 					"Repository": cloneOptions.URL,
// 					"AuthName":   cloneOptions.Auth.Name(),
// 				}).Wrap(err, "could not clone via git")
// 			}
// 			return errors.WithFields(errors.Fields{
// 				"Repository": cloneOptions.URL,
// 			}).Wrap(err, "could not clone via git")
// 		}
// 	} else {
// 		d, err := git.PlainOpen(target)
// 		if err != nil {
// 			return errors.WithFields(errors.Fields{
// 				"LocalRepository": target,
// 			}).Wrap(err, "could not open local repository")
// 		}
// 		wt, err := d.Worktree()
// 		if err != nil {
// 			return errors.Wrap(err, "could not create working tree")
// 		}
// 		err = wt.Pull(pullOptions)
// 		if err != nil {
// 			if err != git.NoErrAlreadyUpToDate {
// 				if cloneOptions.Auth != nil {
// 					return errors.WithFields(errors.Fields{
// 						"Repository": cloneOptions.URL,
// 						"AuthName":   cloneOptions.Auth.Name(),
// 					}).Wrap(err, "could not pull from repository")
// 				}
// 				return errors.WithFields(errors.Fields{
// 					"Repository": cloneOptions.URL,
// 				}).Wrap(err, "could not pull from repository")
// 			}
// 		}
// 	}
// 	cs.CompletedURI = append(cs.CompletedURI, target)
// 	return nil
// }
