package e2e

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

func TestAccBarrelmanVersion(t *testing.T) {
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "" {
		t.Log("To run Acceptance tests, run 'make testacc")
		t.Skip("Skipping Version Test")
	}
	bmBin := os.Getenv("BM_BIN")
	if bmBin == "" {
		t.Fatal("Barrelman binary environment variable BM_BIN not set")
	}

	Convey("When version is run", t, func() {
		out, err := exec.Command(bmBin, "version").CombinedOutput()
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
