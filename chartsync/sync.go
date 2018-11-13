package chartsync

import (
	"fmt"
	"net/url"
	"os"

	"github.com/charter-se/barrelman/sourcetype"
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
	Library      string
	CompletedURI []string
}

type ChartMeta struct {
	Name       string
	Location   string
	SubPath    string
	Depends    []string
	Groups     []string
	SourceType sourcetype.SourceType
}

func New(l string) *ChartSync {
	return &ChartSync{Library: l}
}

func (cs *ChartSync) Sync(acc AccountTable) error {
	for _, v := range cs.Charts {
		switch v.SourceType {
		case sourcetype.Missing:
		case sourcetype.Git:
			if err := cs.gitDownload(v, acc); err != nil {
				return fmt.Errorf("[%v] %v", v.Location, err)
			}
		case sourcetype.Local:
			fmt.Printf("Got sourcetype.Local\n")
		case sourcetype.Unknown:
			return fmt.Errorf("sourceType: Unknown")
		default:
			return fmt.Errorf("Unhandled sourceType: %v", v.SourceType)
		}
	}
	return nil
}

func (cs *ChartSync) GetPath(c *ChartMeta) (string, error) {
	u, err := url.Parse(c.Location)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("%v/%v%v/%v", cs.Library, u.Host, u.Path, c.SubPath)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", fmt.Errorf("%v does not exist: %v", target, err)
	}
	return target, nil
}

func (cs *ChartSync) Add(c *ChartMeta) error {
	cs.Charts = append(cs.Charts, c)
	return nil
}

func (cs *ChartSync) gitDownload(c *ChartMeta, acc AccountTable) error {
	u, err := url.Parse(c.Location)
	if err != nil {
		return err
	}

	cloneOptions := &git.CloneOptions{
		URL:      c.Location,
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

	target := fmt.Sprintf("%v/%v/%v", cs.Library, u.Host, u.Path)
	for _, v := range cs.CompletedURI {
		if v == target {
			//We already downloaded this one
			return nil
		}
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		_, err = git.PlainClone(target, false, cloneOptions)
		if err != nil {
			return fmt.Errorf("Could not clone: %v", err)
		}
	} else {
		d, err := git.PlainOpen(target)
		if err != nil {
			return fmt.Errorf("Could not open local repository: %v", err)
		}
		wt, err := d.Worktree()
		if err != nil {
			return fmt.Errorf("Could not create working tree: %v", err)
		}
		err = wt.Pull(pullOptions)
		if err != nil {
			if err != git.NoErrAlreadyUpToDate {
				return fmt.Errorf("Could pull from repository: %v", err)
			}
		}
	}
	cs.CompletedURI = append(cs.CompletedURI, target)
	return nil
}
