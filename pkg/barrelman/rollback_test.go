package barrelman

import (
	"testing"

	"github.com/charter-oss/barrelman/pkg/cluster"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/charter-oss/barrelman/pkg/cluster/mocks"
	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	"github.com/cirrocloud/structured/errors"
)

func TestRollbackCmd(t *testing.T) {
	newRollbackCmd := func() *RollbackCmd {
		return &RollbackCmd{
			ManifestName: "testManifest",
			Options: &CmdOptions{
				Force:          &[]string{},
				DataDir:        "testdata/datadir",
				ConfigFile:     "testdata/config",
				KubeConfigFile: "testdata/kubeconfig",
				KubeContext:    "default",
				ManifestFile:   "testdata/manifest.yaml",
			},
			Config: &Config{
				Account: make(chartsync.AccountTable),
			},
		}
	}

	Convey("Rollback Command", t, func() {

		rollbackCmd := newRollbackCmd()
		session := &mocks.Sessioner{}
		mockTransaction := &mocks.Transactioner{}

		Convey("Should error on session.Init()", func() {
			session.On("Init").Return(errors.New("simulated Init error"))
			err := rollbackCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("this", func() {
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(rollbackCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(rollbackCmd.Options.KubeContext)
			session.On("GetVersions", rollbackCmd.ManifestName).Return(&cluster.Versions{}, nil)
			session.On("NewTransaction", rollbackCmd.ManifestName).Return(mockTransaction, nil)
			mockTransaction.On("Cancel").Return(nil)
			mockTransaction.On("Complete").Return(nil)
			err := rollbackCmd.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}
