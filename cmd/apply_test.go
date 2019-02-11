package cmd

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/cluster/mocks"
)

func TestNewApplyCmd(t *testing.T) {
	Convey("newApplyCmd", t, func() {
		logOpts := []string{}
		Convey("Can succeed", func() {
			cmd := newApplyCmd(&applyCmd{
				Options:    &cmdOptions{},
				Config:     &Config{},
				LogOptions: &logOpts,
			})
			So(cmd.Name(), ShouldEqual, "apply")
		})
		Convey("Can fail Run", func() {
			cmd := newApplyCmd(&applyCmd{
				Options:    &cmdOptions{},
				Config:     &Config{},
				LogOptions: &logOpts,
			})

			err := cmd.RunE(cmd, []string{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "config file does not exist")
		})
	})
}

func TestDiff(t *testing.T) {
	Convey("Should handle DiffRelease failure", t, func() {

		session := &mocks.Sessioner{}
		rt := releaseTargets{
			&releaseTarget{
				ReleaseMeta: &cluster.ReleaseMeta{
					MetaName:  "storage-minio",
					Namespace: "scratch",
				},
				State: Upgradable,
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
}
func TestApplyRun(t *testing.T) {
	Convey("Run", t, func() {
		Convey("Can fail to find config file", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile: "",
				},
			}
			session := &mocks.Sessioner{}
			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "config file does not exist")
		})

		Convey("Can handle Init failure", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(errors.New("simulated Init failure")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "simulated Init failure")
			session.AssertExpectations(t)
		})

		SkipConvey("Can handle Releases failure", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(nil, errors.New("small error")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "small error")
			session.AssertExpectations(t)
		})

		SkipConvey("Can succeed with one install failure (retry)", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{}, nil).Once()

			//This run of InstallRelease will be a dry run to check for errors
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil).Once()

			//This run will be the real deal, but we error first round
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install failed", "some_release", errors.New("Error injection")).Once()

			//Will delete release in attempt to clear condition
			session.On("DeleteRelease", mock.Anything).Return(nil).Once()

			//This run will be the real deal
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil).Once()

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})

		SkipConvey("Should fail after retry count exceeded", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{}, nil).Once()

			//This run of InstallRelease will be a dry run to check for errors
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil).Once()

			//This run will be the real deal, but we error first round
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install failed", "some_release", errors.New("Error injection")).Times(3)

			session.On("DeleteRelease", mock.Anything).Return(nil).Times(3)
			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Error injection")
			session.AssertExpectations(t)
		})

		SkipConvey("Should succeed in replacing release (Force)", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
					Force:          &[]string{"storage-minio"},
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ChartName: "storage-minio",
				},
			}, nil).Once()

			session.On("DeleteRelease", mock.Anything).Return(nil).Once()
			//This run of InstallRelease will be a dry run to check for errors
			session.On("InstallRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
				mock.Anything,
			).Return("Simulated install succeeded", "some_release", nil)

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})

		SkipConvey("Should fail in replacing release (Force)", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
					Force:          &[]string{"storage-minio"},
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ChartName: "storage-minio",
				},
			}, nil).Once()

			session.On("DeleteRelease", mock.Anything).Return(
				errors.New("you can't delete here"),
			).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "you can't")
			session.AssertExpectations(t)
		})

		SkipConvey("Should succeed in upgrading release", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
					Force:          &[]string{},
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ChartName: "storage-minio",
				},
			}, nil).Once()

			session.On("DiffRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}),
			).Return(true, []byte{}, nil)

			//One with dry-run, one without
			session.On("UpgradeRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
			).Return("Simulated upgrade complete", nil).Twice()

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
		SkipConvey("Should fail in upgrading release", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
					Force:          &[]string{},
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ChartName: "storage-minio",
				},
			}, nil).Once()
			session.On("UpgradeRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
			).Return("Simulated upgrade fail", nil).Once()

			session.On("DiffRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}),
			).Return(true, []byte{}, nil).Once()

			//One with dry-run, one without
			session.On("UpgradeRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
			).Return("Simulated upgrade fail", errors.New("sim fail")).Once()

			err := c.Run(session)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "sim fail")
			session.AssertExpectations(t)
		})

		SkipConvey("Can skip on no change", func() {
			c := &applyCmd{
				Options: &cmdOptions{
					ConfigFile:     getTestDataDir() + "/config",
					ManifestFile:   getTestDataDir() + "/unit-test-manifest.yaml",
					DataDir:        getTestDataDir() + "/",
					KubeConfigFile: getTestDataDir() + "/kube/config",
					DryRun:         false,
					InstallRetry:   3,
					NoSync:         true,
					Force:          &[]string{},
				},
			}
			session := &mocks.Sessioner{}
			session.On("Init").Return(nil).Once()
			session.On("GetKubeConfig").Return(c.Options.KubeConfigFile).Maybe()
			session.On("GetKubeContext").Return("").Once()
			session.On("Releases").Return(map[string]*cluster.ReleaseMeta{
				"storage-minio": &cluster.ReleaseMeta{
					ChartName: "storage-minio",
				},
			}, nil).Once()

			session.On("DiffRelease", mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
				return true
			}),
			).Return(false, []byte{}, nil)

			//One with dry-run, one without
			session.On("UpgradeRelease",
				mock.MatchedBy(func(crm *cluster.ReleaseMeta) bool {
					return true
				}),
			).Return("Simulated upgrade complete", nil).Once()

			err := c.Run(session)
			So(err, ShouldBeNil)
			session.AssertExpectations(t)
		})
	})
}

//getTestDataDir returns a string representing the location of the testdata directory as derived from THIS source file
//our tests are run in temporary directories, so finding the testdata can be a little troublesome
func getTestDataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return fmt.Sprintf("%v/../testdata", path.Dir(filename))
}
