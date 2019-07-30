package e2e

import (
	"os"
	"os/exec"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"regexp"
)

func TestAccBarrelmanVersion(t *testing.T) {
	if accUserValue := os.Getenv("BM_TEST_E2E"); accUserValue == "n" || accUserValue != "Y" {
		t.Log("To run Acceptance tests, run 'BM_TEST_E2E=y BM_BIN=[PathOfBarrelman] RETRYCOUNTACC=20 go test ./e2e -v'")
		t.Skip("Skipping Version Test")
	}
	bmBin := os.Getenv("BM_BIN")
	if bmBin == "" {
		t.Fatal("Barrelman path is not set in BM_BIN parameter")
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

