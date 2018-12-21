package cmd

import (
	"bytes"
	"testing"

	"github.com/charter-se/barrelman/manifest/chartsync"
	. "github.com/smartystreets/goconvey/convey"
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
