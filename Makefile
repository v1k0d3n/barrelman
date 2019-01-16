# Build Environment:
REGISTRY	?=quay.io
NAMESPACE	?=charter-se
VERSION		?=latest
COMMIT		?=$(shell git rev-parse --short HEAD)
BRANCH      ?=$(shell git symbolic-ref -q --short HEAD)

# GoLang Environment:
GOCMD		?=go
GOOS            ?=linux
GOARCH          ?=amd64
BINARY_NAME	?=barrelman
BINARY_ARCH	?=amd64
BINARY_LINUX	?=$(BINARY_NAME)-$(VERSION)-linux-$(BINARY_ARCH)
BINARY_DARWIN	?=$(BINARY_NAME)-$(VERSION)-darwin-$(BINARY_ARCH)
GOBUILD		=$(GOCMD) build
GOCLEAN		=$(GOCMD) clean
GOTEST		=$(GOCMD) test
GOGET		=$(GOCMD) get
SET_VERSION =github.com/charter-se/barrelman/version.version=$(VERSION)
SET_COMMIT  =github.com/charter-se/barrelman/version.commit=$(COMMIT)
SET_BRANCH  =github.com/charter-se/barrelman/version.branch=$(BRANCH)

LDFLAGS         =-w -s -X $(SET_VERSION) -X $(SET_COMMIT) -X $(SET_BRANCH)

all: test build build-linux docker-push

build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)
	rm -f $(BINARY_DARWIN)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:
	$(GOGET) github.com/charter-se/structured


# Go Cross-Compilation Tasks:
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -a -installsuffix cgo -o $(BINARY_LINUX) -v
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -a -installsuffix cgo -o $(BINARY_DARWIN) -v


# Docker Tasks:
## Use: make docker-build BINARY_NAME=barrelman REGISTRY=quay.io NAMESPACE=charter-se VERSION=v0.2.5 COMMIT=$(git rev-parse --short HEAD) GOOS=darwin GOARCH=amd64
docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t $(REGISTRY)/$(NAMESPACE)/$(BINARY_NAME):$(VERSION) .

## Use: make docker-push BINARY_NAME=barrelman REGISTRY=quay.io NAMESPACE=charter-se VERSION=v0.2.5
docker-push:
	docker push $(REGISTRY)/$(NAMESPACE)/$(BINARY_NAME):$(VERSION)

## Use: make docker-push BINARY_NAME=barrelman REGISTRY=quay.io NAMESPACE=charter-se VERSION=v0.2.5 COMMIT=$(git rev-parse --short HEAD)
docker-push-commit:
	docker build $(REGISTRY)/$(NAMESPACE)/$(BINARY_NAME):$(VERSION)-$(COMMIT)