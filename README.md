# Barrelman

[![CodeFactor](https://www.codefactor.io/repository/github/charter-oss/barrelman/badge)](https://www.codefactor.io/repository/github/charter-oss/barrelman)
[![Barrelman Container on Quay](https://quay.io/repository/charter-se/barrelman/status "Docker Repository on Quay")](https://quay.io/repository/charter-se/barrelman)

## Overview

*A project to deploy extremely atomic Helm charts as more complex application release groups.*

Barrelman uses a single manifest to organize complex application deployments that can consist of many 
microservices and independent shared services such as databases and caches.

Barrelman does diff analysis on each release and only executes those changes necessary to achieve 
the desired state.

Additionally, Helm charts can be sourced from different locations like local file, directory, GitHub repos, Helm 
repos, etc. This makes Barrelman manifests very flexible.

## Requirements

- Go >= 1.11 (for module support)
- Access to [charter-oss](https://github.com/charter-oss) GitHub organization

## Build

1. Get the code

    ```sh
    git clone https://github.com/charter-oss/barrelman.git
    cd barrelman
    ```

2. Build the binary

    ```sh
    export GO111MODULE=on
    go build
    ```

## Install as Helm Plugin

As a Helm plugin, Barrelman is largely configured by the Helm environment including Kubernetes server 
settings and authorization keys.

Copy or link the Barrelman binary and plugin.yaml files to `~/.helm/plugins/barrelman/`

```sh
mkdir ~/.helm/plugins/barrelman
cp barrelman ~/.helm/plugins/barrelman
cp plugin.yaml ~/.helm/plugins/barrelman
```

When Barrelman is installed as a plugin you can run commands like `helm barrelman ...`.

## Install as a Standalone Binary

This will install the Barrelman binary to $GOPATH/bin.

```sh
cd barrelman
export GO111MODULE=on
go install
```

If it is in your $PATH you will be able to run commands like `barrelman ...`.

## Quick Start

This quick start will use an example LAMP stack manifest to show some of the basic features of Barrelman.

_We assume Barrelman has been installed as described above. Commands are shown using the standalone 
Barrelman binary. Prepend `helm ` to use Barrelman a Helm plugin._

Requirements:

- Barrelman is installed
- An existing Kubernetes cluster
- Helm and Tiller installed on the Kubernetes cluster, see [Installing Helm](https://helm.sh/docs/using_helm/#installing-helm)

We have a Barrelman manifest defined in `examples/go-web-service/manifest.yaml` that we will use to 
deploy our application. Run the following command to deploy the application:

```sh
cd examples/go-web-service/
barrelman apply manifest.yaml
```

You can test the application by port-forwarding to your Kubernetes cluster and the service created 
by the chart and then browsing to `http://localhost:8080/`.

```sh
kubectl -n barrelman-go-web-service port-forward svc/go-web-service 8080:8080
```

Next we will modify our manifest and scale up the number of pods running our service.

Update the number of replicas set in the `examples/go-web-service/manifest.yaml` file.

```yaml
  values:
    replicas: 3
```

Barrelman can show you a diff of the current state in the cluster and pending changes in your manifest.

```sh
barrelman apply --diff manifest.yaml
```

Now apply the new manifest

```sh
barrelman apply manifest.yaml
```

Verify that Kubernetes has scaled up the pods

```sh
kubectl -n barrelman-go-web-service get pods
```

Finally, let's cleanup by deleting the resources deployed by Barrelman

```sh
barrelman delete manifest.yaml
```

_Note that the namespace is not deleted by this command since we do not want to accidentally delete 
resources that have been created there outside of the Barrelman process._

## Usage

For main usage documentation run

```sh
barrelman help
```

For command help run

```sh
barrelman <command> -h
```

## Authentication

Some Git repositories may require authentication. Barrelman supports authentication using 
[GitHub personal access tokens](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).

The personal access token can be configured in `~/.barrelman/config`

```yaml
---
account:
  - github.com:
      type: token
      user: demond2
      secret: 867530986753098675309
```

## Examples

Example Barrelman manifests can be found in `examples/` and `testdata/`.
