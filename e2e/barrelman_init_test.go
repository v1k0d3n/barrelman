package e2e

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccBarrelmanVersion(t *testing.T) {
	barrelmanPath, _ := os.Getwd()
	Convey("Barrelman Binary Exists", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "version").CombinedOutput()
		So(err, ShouldBeNil)
		So(string(out), ShouldContainSubstring, "Barrelman")
	})

}
