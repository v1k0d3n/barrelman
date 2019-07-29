package e2e

import (
	"os"
	"os/exec"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"regexp"
)

func TestAccBarrelmanVersion(t *testing.T) {
	barrelmanPath, _ := os.Getwd()
        if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "N" || accUserValue == "n" {
		t.Skip("Skipping Version Test")
	}
	Convey("When version is run", t, func() {
		out, err := exec.Command(barrelmanPath+"/../barrelman", "version").CombinedOutput()
		Convey("The output should include Barrelman", func() {
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "msg=Barrelman")
		})
		Convey("The output should include the branch", func() {
			matched, err := regexp.MatchString(`Branch=.*`, string(out))
			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
		})
		Convey("The output should include the commit", func() {
			matched, err := regexp.MatchString(`Commit=.*`, string(out))
			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
	               })
		Convey("The output should include the version", func() {
			matched, err := regexp.MatchString(`Version=.*`, string(out))
			So(err, ShouldBeNil)
	                So(matched, ShouldBeTrue)
		})
	})
}

