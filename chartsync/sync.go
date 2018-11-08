package chartsync

import (
	"fmt"
	"net/url"
	"os"

	"github.com/charter-se/barrelman/sourcetype"
	git "gopkg.in/src-d/go-git.v4"
)

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

func (cs *ChartSync) Sync() error {
	for _, v := range cs.Charts {
		switch v.SourceType {
		case sourcetype.Missing:
			fmt.Printf("Got sourcetype.Missing\n")
		case sourcetype.Git:
			fmt.Printf("Got sourcetype.Git\n")
			if err := cs.gitDownload(v); err != nil {
				return err
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

func (cs *ChartSync) gitDownload(c *ChartMeta) error {
	u, err := url.Parse(c.Location)
	if err != nil {
		return err
	}
	target := fmt.Sprintf("%v/%v/%v", cs.Library, u.Host, u.Path)
	for _, v := range cs.CompletedURI {
		if v == target {
			//We already downloaded this one
			return nil
		}
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		_, err = git.PlainClone(target, false, &git.CloneOptions{
			URL:      c.Location,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
	} else {
		d, err := git.PlainOpen(target)
		if err != nil {
			return err
		}
		wt, err := d.Worktree()
		if err != nil {
			return err
		}
		err = wt.Pull(&git.PullOptions{
			RemoteName:   "origin",
			SingleBranch: true,
			Progress:     os.Stdout,
		})
		if err != nil {
			if err != git.NoErrAlreadyUpToDate {
				return err
			}
		}
	}
	cs.CompletedURI = append(cs.CompletedURI, target)
	return nil
}
