package barrelman

import (
	"bytes"
	"os"
	"testing"

	"github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	"github.com/cirrocloud/yamlpack"
	"github.com/lithammer/dedent"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMFest(t *testing.T) {

	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)
	Convey("process manifest", t, func() {
		config := &manifest.Config{
			DataDir:      tmpDir,
			AccountTable: make(chartsync.AccountTable),
		}
		Convey("can create archives from yaml sections", func() {
			sections, err := sectionsFromBytes(twoGood())
			So(err, ShouldBeNil)
			archives, err := processManifestSections(config, sections, true)
			So(err, ShouldBeNil)
			So(archives.List, ShouldHaveLength, 1)
		})
	})
}

func sectionsFromBytes(name string, bytesIn []byte) ([]*yamlpack.YamlSection, error) {
	yp := yamlpack.New()
	err := yp.Import(name, bytes.NewReader(bytesIn))
	if err != nil {
		return nil, err
	}
	return yp.AllSections(), nil
}

func twoGood() (string, []byte) {
	name := "file-test"
	b := []byte(dedent.Dedent(`
	---
    schema: armada/Chart/v1
    metadata:
        schema: metadata/Document/v1
        name: kubernetes-common
    data:
        chart_name: kubernetes-common
        release: kubernetes-common
        namespace: scratch
        install:
            no_hooks: false
        upgrade:
            no_hooks: false
        values: {}
        source:
            type: file
            location: ./testdata/kubernetes-common.tgz
        dependencies: []
    ---
    schema: armada/Chart/v1
    metadata:
        schema: metadata/Document/v1
        name: storage-minio
    data:
        chart_name: storage-minio
        release: storage-minio
        namespace: scratch
        timeout: 3600
        wait:
        timeout: 3600
        labels:
            release_group: flagship-storage-minio
        install:
            no_hooks: false
        upgrade:
            no_hooks: false
        values: 
            elasticsearch: openstack-minus
        source:
            type: file
            location: ./testdata/test-minio.tgz
        dependencies: 
            - kubernetes-common
    ---
    schema: armada/ChartGroup/v1
    metadata:
        schema: metadata/Document/v1
        name: scratch-test
    data:
        description: "Keystone Infra Services"
        sequenced: True
        chart_group:
        - storage-minio
    ---
    schema: armada/Manifest/v1
    metadata:
        schema: metadata/Document/v1
        name: scratch-manifest
    data:
        release_prefix: armada
        chart_groups:
        - scratch-test
	`))
	return name, b
}
