package chartsync

import (
	"fmt"
	"github.com/charter-se/structured/log"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"net/url"
	"os"
	"sync"

	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	"gopkg.in/src-d/go-git.v4"

	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type SyncGit struct {
	ChartMeta *ChartMeta
	Repo      *gitRepo
	DataDir   string
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
		New: func(dataDir string, cm *ChartMeta, acc AccountTable) (Archiver, error) {
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
		log.Info("syncing git repo ", k)
		// Ensure that the repo is on master before attempting to sync
		err := ReturnToMaster(k)
		if err != nil {
			return err
		}
		if err := r.Download(cs, acc, k); err != nil {
			return err
		}
	}
	return nil
}

func (g *SyncGit) ArchiveRun(ac *ArchiveConfig) (string, error) {
	log.WithFields(log.Fields{
		"DataDir":     ac.DataDir,
		"AcrhivePath": ac.Path,
	}).Debug("Git handler running archiveFunc")
	return ac.ArchiveFunc(ac.DataDir, ac.Path, ac.DependCharts, g.ChartMeta)
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
	//if _, err := os.Stat(target); os.IsNotExist(err) {
	//	return "", errors.WithFields(errors.Fields{"Path": target}).Wrap(err, "target path missing")
	//}
	return target, nil
}

func (r *gitRepoList) Download(cs *ChartSync, acc AccountTable, location string) error {
	u, err := url.Parse(location)
	if err != nil {
		return err
	}

	//ref := plumbing.NewBranchReferenceName("helm_chart_for_mongo")

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
			log.Info("repo is already up to date")
		}
	}
	cs.CompletedURI = append(cs.CompletedURI, target)
	return nil
}

func NewRef(path string, source *Source) error {
	branch := plumbing.NewRemoteReferenceName("origin", source.Reference)
	tag := plumbing.NewTagReferenceName(source.Reference)
	hash := plumbing.NewHash(source.Reference)

	gitOpt := git.PlainOpenOptions{
		DetectDotGit: true,
	}

	r, err := git.PlainOpenWithOptions(path, &gitOpt)
	// searches for the .git file to open the git options
	repo, err := git.PlainOpenWithOptions(path, &gitOpt)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}
	if err = w.Checkout(chk); err != nil {
	// retrieve all the references to search through
	allRef, err := repo.References()
	if err != nil {
		return err
	}

	// iterates through all references. If the reference matches the branch or tag, then the checkout options is created.
	var opt *git.CheckoutOptions
	_ = allRef.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name() == branch || ref.Name() == tag {
			opt = &git.CheckoutOptions{
				Branch: ref.Name(),
			}
		}
		return nil
	})

	// If the branch is not set and the hash is not zero, checkout the hash instead.
	if !hash.IsZero() && opt == nil {
		opt = &git.CheckoutOptions{
			Hash: hash,
		}
	} else {
		return errors.New("reference " + source.Reference + " does not exist")
	}

	wkTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if opt.Branch.IsBranch() || opt.Branch.IsTag() {
		log.Info("checking out " + opt.Branch.String() + " on " + path)
	} else {
		log.Info("checking out " + opt.Hash.String() + " on " + path)
	}
	if err = wkTree.Checkout(opt); err != nil {
		return err
	}

	return nil
}

func ReturnToMaster(path string) error {
	log.Debug("returning ", path, " to master reference")
	branch := plumbing.NewBranchReferenceName("master")

	chk := &git.CheckoutOptions{
		Branch: branch,
	}
	gitOpt := git.PlainOpenOptions{
		DetectDotGit: true,
	}

	r, err := git.PlainOpenWithOptions(path, &gitOpt)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}
	if err = w.Checkout(chk); err != nil {
		return err
	}

	return nil
}
