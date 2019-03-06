//go:generate mockery -output "mockHelm" -dir ../../vendor/k8s.io/helm/pkg/helm -name=Interface

package cluster

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	mockHelm "github.com/charter-se/barrelman/pkg/cluster/mockHelm" // requires mockery run
)

var TestHelm = &mockHelm.Interface{}

func TestNewSession(t *testing.T) {
	Convey("MockNewSession() works", t, func() {
		s := NewMockSession()
		TestHelm.On("GetVersion").Return(&rls.GetVersionResponse{}, nil)
		res, err := s.Helm.GetVersion()
		So(err, ShouldBeNil)
		Print(res)
	})

}

func NewMockSession() *Session {
	//NewSession returns a *Session with kubernetes connections established
	s := &Session{}
	s.Helm = TestHelm
	return s
}
