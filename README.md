[![CodeFactor](https://www.codefactor.io/repository/github/charter-se/barrelman/badge)](https://www.codefactor.io/repository/github/charter-se/barrelman)
[![Barrelman Container on Quay](https://quay.io/repository/charter-se/barrelman/status "Docker Repository on Quay")](https://quay.io/repository/charter-se/barrelman)
# barrelman
*A project to deploy extremely atomic Helm charts as more complex application release groups.*

Barrelman is a [Helm plugin](https://github.com/helm/helm/blob/master/docs/plugins.md) that strives for document compatability with [Armada](https://github.com/att-comdev/armada) and follows Aramada YAML conventions.

## Build

### MacOS
install xcode and xcode tools

depending on the state of your install you may need to run:
```sh
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
```

### Get the code
```sh
go get github.com/CirroCloud/yamlpack
go get github.com/charter-se/structured
git clone https://github.com/charter-se/barrelman.git
cd barrelman
```
### Build
```sh
go build ./...
```
## Install

Copy or link the barrelman binary and plugin.yaml files to ~/.helm/plugins/barrelman/
```sh
mkdir ~/.helm/plugins/barrelman
cp barrelman ~/.helm/plugins/barrelman
cp plugins.yaml ~/.helm/plugins/barrelman
``` 


## Overview
The two main concepts are the ability to process a single YAML file that consists of multiple charts and target state commanding.

The YAML configuration document may contain multiple sub-documents or charts denoted by the YAML directive seperator "---". Each section within the YAML file will be sent as a chart to kubernetes, routed to a kubernetes namespace specified in the section.

Barrelman does diff analysis on each release and only executes those changes necassary to acheive the configured state.

~~Barrelman can be configured to rollback all changes within the current or last transaction on a detected failure, or when commanded by the command line interface. A failure as indicated by kubernetes when commiting one chart will result in the rolling back to the previously commited state on all configured charts.~~  *(rollback not yet implimented)*

As a Helm plugin, Barrelman is largely configured by the Helm environment including Kubernetes server settings and authorization keys.

## Usage

### apply

```sh
helm barrelman apply testdata/flagship-manifest.yaml
```

#### --nosync
Disable automatic repository syncing

#### --dry-run
Run a "dry-run" on each release in the manifest, does not commit any changes

#### --diff
Renders the local charts on k8s and compares them against the running releases and displays any differences, does not commit any changes

## Authenitcation
Some git repositories may require authenitcation in order to syncronize. Barrelman currently supports github [personal access tokens](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).

The personal access token can be configured in ~/.barrelman/config

#### Example
```yaml
---
account:
  - github.com:
      type: token
      user: demond2
      secret: 867530986753098675309
```

## Example Manifest
```yaml
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
    type: git
    location: https://github.com/v1k0d3n/flagship
    subpath: charts/kubernetes-common
    reference: master
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
  values: {}
  source:
    type: git
    location: https://github.com/v1k0d3n/flagship
    subpath: charts/storage-minio
    reference: master
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
```
