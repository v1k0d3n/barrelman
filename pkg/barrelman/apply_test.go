package barrelman

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/cluster/mocks"
	"github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
)

func TestApplyRun(t *testing.T) {

	newTestApplyCmd := func() *ApplyCmd {
		return &ApplyCmd{
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

	Convey("Apply", t, func() {
		applyCmd := newTestApplyCmd()

		session := &mocks.Sessioner{}
		transaction := &mocks.Transactioner{}

		Convey("Should error on session.Init()", func() {
			session.On("Init").Return(errors.New("simulated Init error"))
			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should error on manifest file not found", func() {
			applyCmd.Options.ManifestFile = "testdata/nofile"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "nofile")
			session.AssertExpectations(t)
		})
		Convey("Should successfuly error during sync", func() {
			applyCmd.Options.ManifestFile = "testdata/repo-not-exist.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository does not exist")
			session.AssertExpectations(t)
		})

		Convey("Should successfuly handle error in ListReleases()", func() {
			applyCmd.Options.ManifestFile = "testdata/dir-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			session.On("ReleasesByManifest", mock.AnythingOfType("string")).Return(nil, errors.New("simulated"))
			session.On("NewTransaction", mock.AnythingOfType("string")).Return(transaction, nil)
			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		Convey("Should return nil on diff", func() {
			applyCmd.Options.ManifestFile = "testdata/dir-test-manifest.yaml"
			applyCmd.Options.Diff = true
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			session.On("ReleasesByManifest", mock.AnythingOfType("string")).Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, nil)
			session.On("ChartFromArchive", mock.MatchedBy(func(crm *bytes.Buffer) bool {
				return true
			})).Return(
				&chart.Chart{
					Metadata: &chart.Metadata{
						Name: "storage-minio",
					},
				}, nil)
			session.On("NewTransaction", mock.AnythingOfType("string")).Return(transaction, nil)
			transaction.On("Versions").Return(&cluster.Versions{
				Name: "someVersionsName",
				Data: []*cluster.Version{
					&cluster.Version{
						Name:             "someRelease",
						Namespace:        "someNamespace",
						Revision:         int32(5),
						PreviousRevision: int32(2),
						Modified:         true,
					},
				},
			})

			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string")).Return(&cluster.InstallReleaseResponse{}, nil)

			err := applyCmd.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})

		Convey("Should handle failure on Releases()", func() {
			applyCmd.Options.ManifestFile = "testdata/dir-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			session.On("ReleasesByManifest", mock.AnythingOfType("string")).Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, errors.New("simulated"))
			session.On("NewTransaction", mock.AnythingOfType("string")).Return(transaction, nil)
			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		Convey("Should handle failure on ChartFromArchive()", func() {
			applyCmd.Options.ManifestFile = "testdata/dir-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			session.On("ReleasesByManifest", mock.AnythingOfType("string")).Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, nil)
			session.On("ChartFromArchive", mock.MatchedBy(func(crm *bytes.Buffer) bool {
				return true
			})).Return(
				&chart.Chart{
					Metadata: &chart.Metadata{
						Name: "storage-minio",
					},
				}, errors.New("simulated"))
			session.On("NewTransaction", mock.AnythingOfType("string")).Return(transaction, nil)
			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		Convey("Should handle failure to install release", func() {
			applyCmd.Options.ManifestFile = "testdata/dir-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			session.On("ReleasesByManifest", mock.AnythingOfType("string")).Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, nil)
			session.On("NewTransaction", mock.AnythingOfType("string")).Return(transaction, nil)
			session.On("ChartFromArchive", mock.MatchedBy(func(crm *bytes.Buffer) bool {
				return true
			})).Return(
				&chart.Chart{
					Metadata: &chart.Metadata{
						Name: "storage-minio",
					},
				}, nil)
			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, errors.New("simulated"))

			transaction.On("Versions").Return(&cluster.Versions{
				Name: "someVersionsName",
				Data: []*cluster.Version{
					&cluster.Version{
						Name:             "someRelease",
						Namespace:        "someNamespace",
						Revision:         int32(5),
						PreviousRevision: int32(2),
						Modified:         true,
					},
				},
			})

			err := applyCmd.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		Convey("Should succeed", func() {
			applyCmd.Options.ManifestFile = "testdata/dir-test-manifest.yaml"
			session.On("Init").Return(nil)
			session.On("GetKubeConfig").Return(applyCmd.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return(applyCmd.Options.KubeContext).Maybe()
			session.On("ReleasesByManifest", mock.AnythingOfType("string")).Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					Chart: &chart.Chart{
						Metadata: &chart.Metadata{
							Name: "storage-minio",
						},
					},
				},
			}, nil)
			session.On("NewTransaction", mock.AnythingOfType("string")).Return(transaction, nil)
			transaction.On("Versions").Return(&cluster.Versions{
				Name: "someVersionsName",
				Data: []*cluster.Version{
					&cluster.Version{
						Name:             "someRelease",
						Namespace:        "someNamespace",
						Revision:         int32(5),
						PreviousRevision: int32(2),
						Modified:         true,
					},
				},
			})
			transaction.On("Complete").Return(nil)

			session.On("ChartFromArchive", mock.MatchedBy(func(crm *bytes.Buffer) bool {
				return true
			})).Return(
				&chart.Chart{
					Metadata: &chart.Metadata{
						Name: "storage-minio",
					},
				}, nil)
			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, nil)

			session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(nil)

			err := applyCmd.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})

		Reset(func() {
			chartsync.Reset()
			session = &mocks.Sessioner{}
			applyCmd = newTestApplyCmd()
			os.RemoveAll(applyCmd.Options.DataDir)
		})
	})
}
func TestComputeReleases(t *testing.T) {
	//ComputeReleases calls cluster.ChartFromArchive() directly,
	//this in turn calls chartutil.LoadArchive(). This is not configurable... :(
	//however it also does not change any state, so supply a valid/invalid chart as needed

	chartReader, err := os.Open("testdata/kubernetes-common-0.1.0.tgz")
	if err != nil {
		panic(err)
	}
	defer chartReader.Close()
	Convey("ComputeReleases", t, func() {
		manifestName := "testManifest"
		releaseMatch := "releaseMatch"
		releaseDontMatch := "releaseDontMatch"
		applyCmd := &ApplyCmd{
			Options: &CmdOptions{
				Force: &[]string{},
			},
		}
		session := &mocks.Sessioner{}
		transaction := &mocks.Transactioner{}
		releases := map[string]*cluster.ReleaseMeta{
			releaseMatch: &cluster.ReleaseMeta{
				ReleaseName: releaseMatch,
			},
		}
		Convey("Should fail to generate chart archive", func() {
			archives := &manifest.ArchiveFiles{
				List: []*manifest.ArchiveSpec{
					&manifest.ArchiveSpec{
						MetaName:  "test",
						ChartName: "testChart",
						Reader:    bytes.NewReader([]byte{}),
						Namespace: "default",
						Overrides: []byte{},
					},
				},
			}

			session.On("ChartFromArchive", mock.MatchedBy(func(r io.Reader) bool {
				return true
			}),
			).Return(&cluster.Chart{}, errors.New("simulated fail in DiffRelease"))

			_, err := applyCmd.ComputeReleases(session, transaction, manifestName, archives, releases)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Failed to generate chart from archive")
			session.AssertExpectations(t)
		})
		Convey("Should result in Replaceable", func() {
			applyCmd.Options.Force = &[]string{
				releaseMatch,
			}
			archives := &manifest.ArchiveFiles{
				List: []*manifest.ArchiveSpec{
					&manifest.ArchiveSpec{
						ReleaseName: releaseMatch,
						MetaName:    "test",
						ChartName:   "testChart",
						Reader:      chartReader,
						Namespace:   "default",
						Overrides:   []byte{},
					},
				},
			}

			session.On("ChartFromArchive", mock.MatchedBy(func(r io.Reader) bool {
				return true
			}),
			).Return(&cluster.Chart{}, nil)

			transaction.On("Versions").Return(&cluster.Versions{
				Name: "someVersionsName",
				Data: []*cluster.Version{
					&cluster.Version{
						Name:             "someRelease",
						Namespace:        "someNamespace",
						Revision:         int32(5),
						PreviousRevision: int32(2),
						Modified:         true,
					},
				},
			})

			// session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
			// 	return true
			// })).Return(nil)

			releaseTargets, err := applyCmd.ComputeReleases(session, transaction, manifestName, archives, releases)
			So(err, ShouldBeNil)
			So(releaseTargets.Data[0].TransitionState, ShouldEqual, Replaceable)
			session.AssertExpectations(t)
		})
		Convey("Should result in Upgradable", func() {
			archives := &manifest.ArchiveFiles{
				List: []*manifest.ArchiveSpec{
					&manifest.ArchiveSpec{
						ReleaseName: releaseMatch,
						MetaName:    "test",
						ChartName:   "testChart",
						Reader:      chartReader,
						Namespace:   "default",
						Overrides:   []byte{},
					},
				},
			}

			session.On("ChartFromArchive", mock.MatchedBy(func(r io.Reader) bool {
				return true
			}),
			).Return(&cluster.Chart{}, nil)

			transaction.On("Versions").Return(&cluster.Versions{
				Name: "someVersionsName",
				Data: []*cluster.Version{
					&cluster.Version{
						Name:             "someRelease",
						Namespace:        "someNamespace",
						Revision:         int32(5),
						PreviousRevision: int32(2),
						Modified:         true,
					},
				},
			})

			releaseTargets, err := applyCmd.ComputeReleases(session, transaction, manifestName, archives, releases)
			So(err, ShouldBeNil)
			So(releaseTargets.Data[0].TransitionState, ShouldEqual, Upgradable)
			session.AssertExpectations(t)
		})
		Convey("Should result in Installable", func() {
			archives := &manifest.ArchiveFiles{
				List: []*manifest.ArchiveSpec{
					&manifest.ArchiveSpec{
						ReleaseName: releaseDontMatch,
						MetaName:    "test",
						ChartName:   "testChart",
						Reader:      chartReader,
						Namespace:   "default",
						Overrides:   []byte{},
					},
				},
			}

			session.On("ChartFromArchive", mock.MatchedBy(func(r io.Reader) bool {
				return true
			}),
			).Return(&cluster.Chart{}, nil)

			transaction.On("Versions").Return(&cluster.Versions{
				Name: "someVersionsName",
				Data: []*cluster.Version{
					&cluster.Version{
						Name:             "someRelease",
						Namespace:        "someNamespace",
						Revision:         int32(5),
						PreviousRevision: int32(2),
						Modified:         true,
					},
				},
			})

			releaseTargets, err := applyCmd.ComputeReleases(session, transaction, manifestName, archives, releases)
			So(err, ShouldBeNil)
			So(releaseTargets.Data[0].TransitionState, ShouldEqual, Installable)
			session.AssertExpectations(t)
		})
		Reset(func() {
			applyCmd.Options.Force = &[]string{}
		})
	})
}

func TestApply(t *testing.T) {
	Convey("Apply", t, func() {
		session := &mocks.Sessioner{}
		rt := ReleaseTargets{
			session: session,
			Data: []*ReleaseTarget{
				&ReleaseTarget{
					ReleaseMeta: &cluster.ReleaseMeta{
						MetaName:  "storage-minio",
						Namespace: "scratch",
					},
					TransitionState: Replaceable,
					ReleaseVersion:  &cluster.Version{},
				},
			},
		}
		opt := &CmdOptions{
			Force:        &[]string{},
			InstallRetry: 3,
		}
		Convey("Should handle DeleteRelease failure with state Replaceable", func() {
			session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(errors.New("simulated fail in DeleteRelease"))
			err := rt.Apply(opt)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should handle InstallRelease failure with state Installable, then fallthrough to UpgradeRelease", func() {
			rt.Data[0].TransitionState = Installable

			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, errors.New("simulated fail in InstallRelease")).Times(3)

			session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(nil)

			err := rt.Apply(opt)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should succeed in InstallRelease", func() {
			rt.Data[0].TransitionState = Installable
			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, nil)

			err := rt.Apply(opt)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Convey("Should handle InstallRelease failure with state Installable, then succeed", func() {
			rt.Data[0].TransitionState = Installable
			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, errors.New("simulated fail in InstallRelease")).Once()

			session.On("DeleteRelease", mock.MatchedBy(func(crm *cluster.DeleteMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(nil)

			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, nil).Once()

			err := rt.Apply(opt)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Convey("Should fail in UpgradeRelease with state Upgradable", func() {
			rt.Data[0].TransitionState = Upgradable
			rt.Data[0].Changed = true
			session.On("UpgradeRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.UpgradeReleaseResponse{}, errors.New("simulated fail in UpgradeRelease")).Times(1)

			err := rt.Apply(opt)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should succeed in UpgradeRelease with state Upgradable", func() {
			rt.Data[0].TransitionState = Upgradable
			rt.Data[0].Changed = true
			session.On("UpgradeRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.UpgradeReleaseResponse{}, nil).Times(1)

			err := rt.Apply(opt)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Convey("Should skip due to Upgradable and no change", func() {
			rt.Data[0].TransitionState = Upgradable
			rt.Data[0].Changed = false

			err := rt.Apply(opt)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Convey("Should skip unhandled state", func() {
			rt.Data[0].TransitionState = -1

			err := rt.Apply(opt)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Reset(func() {
			rt.Data[0].TransitionState = Replaceable
			rt.Data[0].Changed = false
		})
	})
}

func TestDiff(t *testing.T) {
	Convey("Diff", t, func() {
		Convey("Should handle DiffRelease failure", func() {

			session := &mocks.Sessioner{}
			rt := ReleaseTargets{
				Data: []*ReleaseTarget{
					&ReleaseTarget{
						ReleaseMeta: &cluster.ReleaseMeta{
							MetaName:  "storage-minio",
							Namespace: "scratch",
						},
						TransitionState: Upgradable,
					},
				},
			}

			session.On("DiffRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}),
			).Return(true, []byte{}, errors.New("simulated fail in DiffRelease"))
			_, err := rt.Diff(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
		Convey("Should handle DiffRelease success", func() {

			session := &mocks.Sessioner{}
			rt := ReleaseTargets{
				Data: []*ReleaseTarget{
					&ReleaseTarget{
						ReleaseMeta: &cluster.ReleaseMeta{
							MetaName:  "storage-minio",
							Namespace: "scratch",
						},
						TransitionState: Upgradable,
					},
				},
			}

			session.On("DiffRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}),
			).Return(true, []byte{}, nil)
			_, err := rt.Diff(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}
func TestDryRun(t *testing.T) {
	Convey("dryRun", t, func() {
		session := &mocks.Sessioner{}
		rt := ReleaseTargets{
			Data: []*ReleaseTarget{
				&ReleaseTarget{
					ReleaseMeta: &cluster.ReleaseMeta{
						MetaName:  "storage-minio",
						Namespace: "scratch",
					},
					TransitionState: Upgradable,
				},
			},
		}
		Convey("Should handle Upgradable, no error", func() {
			session.On("UpgradeRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.UpgradeReleaseResponse{}, nil)
			err := rt.dryRun(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Convey("Should handle Upgradable, with error", func() {
			session.On("UpgradeRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.UpgradeReleaseResponse{}, errors.New("simulated error"))
			err := rt.dryRun(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})

		rt.Data[0].TransitionState = Installable
		Convey("Should handle Installable, no error", func() {
			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, nil)
			err := rt.dryRun(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		Convey("Should handle Installable, with error", func() {
			session.On("InstallRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}), mock.AnythingOfType("string"),
			).Return(&cluster.InstallReleaseResponse{}, errors.New("simulated error"))
			err := rt.dryRun(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated")
			session.AssertExpectations(t)
		})
	})

}
