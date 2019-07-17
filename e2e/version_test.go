package e2e

import (
	// "errors"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVersionCommand(t *testing.T) {
	// t.Skip("test")
	Convey("Version", t, func() {
		out, err := exec.Command("/Users/p2723614/go/src/github.com/charter-oss/barrelman/barrelman", "version").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
	})
}
