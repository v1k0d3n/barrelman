package barrelman

import (
	"bytes"
	"fmt"
	"path"
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
)

func TestConfig(t *testing.T) {
	Convey("config", t, func() {
		Convey("Can succeed from file", func() {
			c, err := GetConfigFromFile(getTestDataDir() + "/config")
			So(err, ShouldBeNil)
			So(c, ShouldNotBeNil)
			So(c.Account, ShouldContainKey, "github.com")
		})
		Convey("Can fail from file", func() {
			_, err := GetConfigFromFile(getTestDataDir() + "/unit-test-manifest.yaml")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unit-test-manifest.yaml")
		})
		Convey("Can fail to parse", func() {
			config := &Config{}
			config.Account = make(map[string]*chartsync.Account)
			r := bytes.NewBufferString(badConfig)
			bc, err := toBarrelmanConfig("/pretend/path", r)
			So(err, ShouldBeNil)
			_, err = config.LoadAcc(bc)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to parse")
		})
	})
}

var badConfig = `
account:
  github.com:
     type: token
     user: username
     secret: 12345678901011112113114115
`

//getTestDataDir returns a string representing the location of the testdata directory as derived from THIS source file
//our tests are run in temporary directories, so finding the testdata can be a little troublesome
func getTestDataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return fmt.Sprintf("%v/../../testdata", path.Dir(filename))
}
