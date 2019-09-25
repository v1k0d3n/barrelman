//go:generate mockery -output "mockHelm" -dir ../../vendor/k8s.io/helm/pkg/helm -name=Interface
//go:generate mockery -output "mockClientset" -dir ../../vendor/k8s.io/client-go/kubernetes -name=Interface
//go:generate mockery -output "mockCoreV1" -dir ../../vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name=CoreV1Interface
//go:generate mockery -output "mockConfigMap" -dir ../../vendor/k8s.io/client-go/kubernetes/typed/core/v1 -name=ConfigMapInterface

package cluster

import (
	"os"
	"testing"

	mockClientset "github.com/charter-oss/barrelman/pkg/cluster/mockClientset"
	mockConfigMap "github.com/charter-oss/barrelman/pkg/cluster/mockConfigMap"
	mockCoreV1 "github.com/charter-oss/barrelman/pkg/cluster/mockCoreV1"
	mockHelm "github.com/charter-oss/barrelman/pkg/cluster/mockHelm"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/helm/pkg/kube"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

var TestHelm = &mockHelm.Interface{}
var TestClientSet = &mockClientset.Interface{}
var TestCoreV1 = &mockCoreV1.CoreV1Interface{}
var TestConfigMap = &mockConfigMap.ConfigMapInterface{}

func TestNewSession(t *testing.T) {
	Convey("MockNewSession() works", t, func() {
		s := NewMockSession()
		TestHelm.On("GetVersion").Return(&rls.GetVersionResponse{}, nil)
		res, err := s.Helm.GetVersion()
		So(err, ShouldBeNil)
		Print(res)
	})
	Convey("TLS", t, func() {
		os.Setenv("HELM_TLS_CA_CERT", "")
		os.Setenv("HELM_TLS_CERT", "")
		os.Setenv("HELM_TLS_KEY", "")
		os.Setenv("HELM_TLS_ENABLE", "")
		os.Setenv("HELM_TLS_VERIFY", "")
		os.Setenv("HELM_HOME", "")

		SkipConvey("Can init", func() {
			c := NewSession("", "")
			err := c.Init()
			So(err, ShouldBeNil)
		})
		SkipConvey("Can load TLS files", func() {
			os.Setenv("HELM_TLS_ENABLE", "true")
			os.Setenv("HELM_TLS_CA_CERT", "./testdata/ca.pem")
			os.Setenv("HELM_TLS_CERT", "./testdata/server1.pem")
			os.Setenv("HELM_TLS_KEY", "./testdata/server1.key")
			c := NewSession("", "")
			err := c.Init()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "context deadline exceeded")
		})

		SkipConvey("Can fail to load key results in os.Exit(2)", func() {
			os.Setenv("HELM_TLS_ENABLE", "true")
			os.Setenv("HELM_TLS_CA_CERT", "./testdata/ca.pem")
			os.Setenv("HELM_TLS_CERT", "./testdata/server1.pem")
			os.Setenv("HELM_TLS_KEY", "./testdata/missing.key")
			c := NewSession("", "")
			err := c.Init()
			So(err, ShouldBeNil)
		})

		Reset(func() {
			os.Setenv("HELM_TLS_CA_CERT", "")
			os.Setenv("HELM_TLS_CERT", "")
			os.Setenv("HELM_TLS_KEY", "")
			os.Setenv("HELM_TLS_ENABLE", "")
			os.Setenv("HELM_TLS_VERIFY", "")
			os.Setenv("HELM_TLS_HOME", "")
		})
	})

}

func NewMockSession() *Session {
	//NewSession returns a *Session with kubernetes connections established
	s := &Session{}
	s.Helm = TestHelm
	s.Tunnel = &kube.Tunnel{}
	s.Clientset = TestClientSet
	return s
}
