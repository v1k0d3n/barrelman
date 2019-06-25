package barrelman

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/cluster/mocks"
	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	"github.com/cirrocloud/structured/errors"
)

func TestDeleteRun(t *testing.T) {

	newTestDeleteCmd := func() *DeleteCmd {
		return &DeleteCmd{
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

	Convey("Delete", t, func() {
		delCmd := newTestDeleteCmd()

		session := &mocks.Sessioner{}

		Convey("Should error on session.Init()", func() {
			session.On("Init").Return(errors.New("simulated Init error"))
			err := delCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		Convey("Should error on manifest file not found", func() {
			delCmd.Options.ManifestFile = "testdata/nofile"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(delCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(delCmd.Options.KubeContext)
			err := delCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "nofile")
			session.AssertExpectations(t)
		})

		Convey("Should successfuly error during sync", func() {
			delCmd.Options.ManifestFile = "testdata/repo-not-exist.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(delCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(delCmd.Options.KubeContext)
			err := delCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository does not exist")
			session.AssertExpectations(t)
		})

		Convey("Should successfuly handle error in ListReleases()", func() {
			delCmd.Options.ManifestFile = "testdata/file-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(delCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(delCmd.Options.KubeContext)
			session.On("ListReleases").Return([]*cluster.Release{}, errors.New("simulated"))
			err := delCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		Convey("Should successfuly handle error in DeleteByManifest()", func() {
			delCmd.Options.ManifestFile = "testdata/file-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(delCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(delCmd.Options.KubeContext)
			session.On("ListReleases").Return([]*cluster.Release{&cluster.Release{
				ReleaseName: "storage-minio",
			}}, nil)
			session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
				return true
			})).Return(errors.New("simulated"))
			err := delCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should complete without error", func() {
			delCmd.Options.ManifestFile = "testdata/file-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(delCmd.Options.KubeConfigFile)
			session.On("GetKubeContext").Return(delCmd.Options.KubeContext)
			session.On("ListReleases").Return([]*cluster.Release{&cluster.Release{
				ReleaseName: "storage-minio",
			}}, nil)
			session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
				return true
			})).Return(nil)
			err := delCmd.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})

		Reset(func() {
			chartsync.Reset()
			session = &mocks.Sessioner{}
			delCmd = newTestDeleteCmd()
			os.RemoveAll(delCmd.Options.DataDir)
		})
	})
}
