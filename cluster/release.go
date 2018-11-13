package cluster

import (
	"errors"
	"time"

	"google.golang.org/grpc"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type ReleaseMeta struct {
	Path             string
	Namespace        string
	ValueOverrides   []byte
	InstallDryRun    bool
	InstallReuseName bool
	InstallWait      bool
	InstallTimeout   time.Duration
}

type DeleteMeta struct {
	Namespace string
	Name      string
}

type Chart = chart.Chart

type Release struct {
	Chart     *Chart
	Name      string
	Namespace string
}

func (s *Session) ListReleases() ([]*Release, error) {
	var res []*Release
	r, err := s.Helm.ListReleases()
	for _, v := range r.GetReleases() {
		rel := &Release{
			Chart:     v.GetChart(),
			Name:      v.Name,
			Namespace: v.Namespace,
		}
		res = append(res, rel)
	}
	return res, err
}

func (s *Session) InstallRelease(m *ReleaseMeta, chart []byte) error {
	_, err := s.Helm.InstallRelease(
		m.Path,
		m.Namespace,
		helm.ValueOverrides(m.ValueOverrides),
		helm.InstallDryRun(m.InstallDryRun),
		helm.InstallReuseName(m.InstallReuseName),
		helm.InstallWait(m.InstallWait),
		helm.InstallTimeout(int64(m.InstallTimeout.Seconds())),
	)
	//fmt.Printf("\t[RESPONSE: %v]\n", res)
	return err
}

func (s *Session) UpdateRelease(m *ReleaseMeta, chart []byte) error {
	_, err := s.Helm.UpdateRelease(
		m.Path,
		m.Namespace,
		// helm.ValueOverrides(m.ValueOverrides),
		// helm.InstallDryRun(m.InstallDryRun),
		// helm.InstallReuseName(m.InstallReuseName),
		// helm.InstallWait(m.InstallWait),
		// helm.InstallTimeout(int64(m.InstallTimeout.Seconds())),
	)
	// fmt.Printf("\t[RESPONSE: %v]\n", res)
	return err
}

func (s *Session) DeleteReleases(dm []*DeleteMeta) error {
	for _, v := range dm {
		if err := s.DeleteRelease(v); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) DeleteRelease(m *DeleteMeta) error {
	_, err := s.Helm.DeleteRelease(m.Name, helm.DeletePurge(true))
	if err != nil {
		return errors.New(grpc.ErrorDesc(err))
	}
	return nil
}
