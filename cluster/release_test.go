package cluster

import (
	"testing"

	mockCluster "github.com/charter-se/barrelman/cluster/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListReleases(t *testing.T) {
	session := mockCluster.Sessioner{}
	Convey("Given some integer with a starting value", t, func() {
		x := 1
		Convey("When the integer is incremented", func() {
			x++

			Convey("The value should be greater by one", func() {
				So(x, ShouldEqual, 2)
			})
		})
	})
}
