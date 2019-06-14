package cluster

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestVersioning(t *testing.T) {
	manifestName := "testManifest"
	s := NewMockSession()
	s.Tunnel.Namespace = "system"
	Convey("GetVersions", t, func() {
		TestClientSet.On("CoreV1").Return(TestCoreV1)
		TestCoreV1.On("ConfigMaps", mock.Anything).Return(TestConfigMap)

		// Items will need to be populated with pkg/cluster/driver/util.go:decodeRelease()
		TestConfigMap.On("List", mock.Anything).Return(&v1.ConfigMapList{
			Items: []v1.ConfigMap{},
		}, nil)

		versions, err := s.GetVersions(manifestName)
		So(err, ShouldBeNil)
		Convey("this", func() {
			So(versions.Data, ShouldHaveLength, 0)
		})
	})
}
