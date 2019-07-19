package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBarrelmanCommand(t *testing.T) {
	barrelmanPath := fullPath()
	Convey("Barrelman Binary Exists", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "version").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
		fmt.Fprintln(os.Stdout, string(out))
	})

}
