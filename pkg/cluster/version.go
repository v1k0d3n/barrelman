package cluster

import (
	"fmt"

	"github.com/charter-oss/barrelman/pkg/cluster/driver"
	"github.com/charter-oss/structured/errors"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
)

type Versions []*Version

type Versioner interface {
	GetVersions() (Versions, error)
}

type Version struct {
	Name      string
	Namespace string
	Revision  int32
	Chart     string
}

func (s *Session) GetVersions() (Versions, error) {
	versions := Versions{}
	cmap := driver.NewConfigMaps(s.Clientset.Core().ConfigMaps(s.Tiller.Namespace))
	releases, err := cmap.List(releaseFilter)
	if err != nil {
		return nil, errors.Wrap(err, "GetVersion failed to get release list")
	}
	for _, v := range releases {
		fmt.Printf("%v: %v\n", v.Name, v.Version)
		versions = append(versions, &Version{
			Name:      v.Name,
			Namespace: v.Namespace,
			Revision:  v.Version,
			Chart:     v.Info.GetDescription(),
		})
	}
	return versions, nil
}

func releaseFilter(rls *rspb.Release) bool {
	return true
}
