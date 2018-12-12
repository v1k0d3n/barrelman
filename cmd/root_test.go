package cmd

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRootCmd(t *testing.T) {
	Convey("newRootCmd", t, func() {
		Convey("Can succeed", func() {
			cmd := newRootCmd([]string{})
			So(cmd.Name(), ShouldEqual, "")
		})
	})
}
