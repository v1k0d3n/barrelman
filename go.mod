module github.com/charter-oss/barrelman

require (
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Masterminds/sprig v2.18.0+incompatible
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/charter-oss/structured v0.0.2
	github.com/cirrocloud/yamlpack v0.0.1
	github.com/docker/distribution v2.7.0+incompatible // indirect
	github.com/docker/docker v0.0.0-20181201151923-ad1354ffb423 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/golang/groupcache v0.0.0-20181024230925-c65c006176ff // indirect
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181103185306-d547d1d9531e // indirect
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/jtolds/gls v0.0.0-20181110203027-b4936e06046b // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/petar/GoLLRB v0.0.0-20130427215148-53be0d36a84c // indirect
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190208162519-de1b801bf34b // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/smartystreets/assertions v0.0.0-20180301161246-7678a5452ebe // indirect
	github.com/smartystreets/goconvey v0.0.0-20170602164621-9e8dc3f972df
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2-0.20190315063904-3954e415200e
	github.com/stretchr/testify v1.2.2
	golang.org/x/oauth2 v0.0.0-20181128211412-28207608b838 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	google.golang.org/appengine v1.3.0 // indirect
	google.golang.org/genproto v0.0.0-20181202183823-bd91e49a0898 // indirect
	google.golang.org/grpc v1.18.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.2.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.8.1
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31
	k8s.io/helm v2.14.0+incompatible
	k8s.io/klog v0.1.0 // indirect
	k8s.io/kube-openapi v0.0.0-20181114233023-0317810137be // indirect
	k8s.io/kubernetes v1.13.1
	k8s.io/utils v0.0.0-20180907001310-011bbbe3b287 // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

replace github.com/spf13/viper => ../../demond2/viper

replace k8s.io/helm => ../../demond2/helm
