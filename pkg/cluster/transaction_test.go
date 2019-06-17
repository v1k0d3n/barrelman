package cluster

import (
	"testing"

	"github.com/charter-oss/structured/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

func TestTransactions(t *testing.T) {
	manifestName := "testManifest"
	s := NewMockSession()

	Convey("NewTransaction can succeed with no change", t, func() {
		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil).Twice()

		TestClientSet.On("CoreV1").Return(TestCoreV1).Twice()
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap).Twice()
		transaction, err := s.NewTransaction(manifestName)
		So(err, ShouldBeNil)
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
		Convey("Transaction can complete with no change", func() {
			err := transaction.Complete()
			So(err, ShouldBeNil)
			Convey("then fail to complete again", func() {
				err := transaction.Complete()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "its already been completed")
			})
		})
	})

	Convey("NewTransaction can succeed with change", t, func() {
		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil).Twice()

		TestClientSet.On("CoreV1").Return(TestCoreV1).Twice()
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap).Twice()
		transaction, err := s.NewTransaction(manifestName)
		So(err, ShouldBeNil)
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
		Convey("Transaction can complete with change", func() {
			TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
				Items: []v1.ConfigMap{},
			}, nil).Once()
			TestClientSet.On("CoreV1").Return(TestCoreV1).Once()
			TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap).Once()
			TestConfigMap.On("Create", mock.Anything).Return(&v1.ConfigMap{}, nil).Once()
			transaction.SetChanged()
			err := transaction.Complete()
			So(err, ShouldBeNil)
			TestClientSet.AssertExpectations(t)
			TestConfigMap.AssertExpectations(t)
			TestCoreV1.AssertExpectations(t)
			Convey("then fail to complete again", func() {
				err := transaction.Complete()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "its already been completed")
			})
		})
	})

	Convey("NewTransaction can fail to get versions", t, func() {
		TestConfigMap.On("List", mock.Anything).Return(nil, errors.New("Simulated List failure")).Once()

		TestClientSet.On("CoreV1").Return(TestCoreV1).Once()
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap).Once()
		_, err := s.NewTransaction(manifestName)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Simulated List failure")
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
	})

	Convey("NewTransaction can fail to create new rollback", t, func() {
		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil).Once()
		TestConfigMap.On("List", mock.Anything).Return(nil, errors.New("Simulated failure in rollback")).Once()
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)
		_, err := s.NewTransaction(manifestName)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Simulated failure in rollback")
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
	})
	Convey("Transaction can fail to complete when not started", t, func() {
		transaction := &Transaction{
			startState: &State{},
			endState:   &State{},
		}
		err := transaction.Complete()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "has not been started")
	})

	Convey("Transaction can be canceled", t, func() {
		newReleaseName := "newRelease"
		newRevision := int32(3)
		newRelease := &Version{
			Name:             newReleaseName,
			Namespace:        "oltherNamespace",
			Revision:         newRevision,
			PreviousRevision: int32(2),
			Chart:            &chart.Chart{},
			Info:             &release.Info{},
			Modified:         true,
		}

		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil).Twice()
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)
		transaction, err := s.NewTransaction(manifestName)
		So(err, ShouldBeNil)
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
		transaction.Versions().AddReleaseVersion(newRelease)

		TestHelm.On("RollbackRelease", newReleaseName, mock.Anything, mock.Anything).Return(
			&rls.RollbackReleaseResponse{
				Release: &release.Release{},
			},
			nil,
		).Once()

		err = transaction.Cancel()
		So(err, ShouldBeNil)
		TestHelm.AssertExpectations(t)
	})

	Convey("Transaction can fail to be canceled", t, func() {
		newReleaseName := "newRelease"
		newRevision := int32(3)
		newRelease := &Version{
			Name:             newReleaseName,
			Namespace:        "oltherNamespace",
			Revision:         newRevision,
			PreviousRevision: int32(2),
			Chart:            &chart.Chart{},
			Info:             &release.Info{},
			Modified:         true,
		}

		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil).Twice()
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)
		transaction, err := s.NewTransaction(manifestName)
		So(err, ShouldBeNil)
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
		transaction.Versions().AddReleaseVersion(newRelease)

		TestHelm.On("RollbackRelease", newReleaseName, mock.Anything, mock.Anything).Return(
			nil,
			grpc.ErrServerStopped,
		).Once()

		err = transaction.Cancel()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "grpc: the server has been stopped")
		TestHelm.AssertExpectations(t)
	})

	Convey("Transaction can fail to be canceled with pruge", t, func() {
		newReleaseName := "newRelease"
		newRevision := int32(3)
		newRelease := &Version{
			Name:             newReleaseName,
			Namespace:        "oltherNamespace",
			Revision:         newRevision,
			PreviousRevision: int32(0),
			Chart:            &chart.Chart{},
			Info:             &release.Info{},
			Modified:         true,
		}

		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil).Twice()
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)
		transaction, err := s.NewTransaction(manifestName)
		So(err, ShouldBeNil)
		TestConfigMap.AssertExpectations(t)
		TestClientSet.AssertExpectations(t)
		TestCoreV1.AssertExpectations(t)
		transaction.Versions().AddReleaseVersion(newRelease)

		TestHelm.On("DeleteRelease", newReleaseName, mock.Anything, mock.Anything).Return(
			&rls.UninstallReleaseResponse{},
			grpc.ErrServerStopped).Once()

		err = transaction.Cancel()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "grpc: the server has been stopped")
		TestHelm.AssertExpectations(t)
	})
}
