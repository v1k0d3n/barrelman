package cluster

import (
	"fmt"
	"time"

	"k8s.io/helm/pkg/helm"
)

type ReleaseMeta struct {
	Path             string
	NameSpace        string
	ValueOverrides   []byte
	InstallDryRun    bool
	InstallReuseName bool
	InstallWait      bool
	InstallTimeout   time.Duration
}

func (s *Session) InstallRelease(m *ReleaseMeta, chart []byte) error {
	res, err := s.Helm.InstallRelease(
		m.Path,
		m.NameSpace,
		helm.ValueOverrides(m.ValueOverrides),
		helm.InstallDryRun(m.InstallDryRun),
		helm.InstallReuseName(m.InstallReuseName),
		helm.InstallWait(m.InstallWait),
		helm.InstallTimeout(int64(m.InstallTimeout.Seconds())),
		helm.InstallOption(helm.UpgradeForce(true)),
	)
	fmt.Printf("RESPONSE: %v\n", res)
	return err

	// _, err = e.helmClient.InstallRelease(
	// 	chartPath,
	// 	cmp.Namespace,
	// 	helm.ValueOverrides([]byte(rawValues)),
	// 	helm.ReleaseName(cmp.Name),
	// 	helm.InstallDryRun(e.dryRun),
	// 	helm.InstallReuseName(true),
	// 	helm.InstallWait(e.wait),
	// 	helm.InstallTimeout(e.waitTimeout),
	// )
	// if err != nil {
	// 	return errors.New(grpc.ErrorDesc(err))
	// }
}
