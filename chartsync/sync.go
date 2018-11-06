package chartsync

import (
	"fmt"

	"github.com/charter-se/barrelman/sourcetype"
	git "gopkg.in/src-d/go-git.v4"
)

type ChartSync struct {
	Charts []*ChartMeta
}

type ChartMeta struct {
	Name       string
	Depends    []string
	Groups     []string
	SourceType sourcetype.SourceType
}

func New() *ChartSync {
	return &ChartSync{}
}

func (cs *ChartSync) Sync() error {
	for _, v := range cs.Charts {
		switch v.SourceType {
		case sourcetype.Missing:
			fmt.Printf("Got sourcetype.Missing\n")
		case sourcetype.Git:
			fmt.Printf("Got sourcetype.Git\n")
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

func (cs *ChartSync) Add(c *ChartMeta) error {
	cs.Charts = append(cs.Charts, c)
	return nil
}

func (cs *ChartSync) gitDownload(c *ChartMeta) error {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "https://github.com/src-d/go-siva",
	})
	if err != nil {
		return err
	}

}
