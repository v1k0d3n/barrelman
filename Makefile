# Build Environment:
REGISTRY        ?=quay.io
NAMESPACE       ?=charter-se
VERSION         ?=latest
COMMIT          ?=$(shell git rev-parse --short HEAD)
BRANCH          ?=$(shell git symbolic-ref -q --short HEAD)

# GoLang Environment:
GOCMD           ?=go
DEPCMD		    ?=dep
GOOS            ?=linux
GOARCH          ?=amd64
BINARY_NAME     ?=barrelman
BINARY_ARCH     ?=amd64
BINARY_LINUX    ?=$(BINARY_NAME)-$(VERSION)-linux-$(BINARY_ARCH)
BINARY_DARWIN   ?=$(BINARY_NAME)-$(VERSION)-darwin-$(BINARY_ARCH)
GOBUILD         =$(GOCMD) build
GOCLEAN         =$(GOCMD) clean
GOTEST          =$(GOCMD) test
GODEP           =$(DEPCMD) ensure
SET_VERSION     =github.com/charter-oss/barrelman/pkg/version.version=$(VERSION)
SET_COMMIT      =github.com/charter-oss/barrelman/pkg/version.commit=$(COMMIT)
SET_BRANCH      =github.com/charter-oss/barrelman/pkg/version.branch=$(BRANCH)

LDFLAGS         =-w -s -X $(SET_VERSION) -X $(SET_COMMIT) -X $(SET_BRANCH)

all: test build build-linux docker-push

build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) -v

test:
	BM_TEST_E2E=${BM_TEST_E2E:""}
	if [ ${BM_TEST_E2E} == "true" ]; then $(GOTEST) ./e2e/ -v; else $(GOTEST) -v ./...; fi

acc:
	$(GOTEST) ./e2e/ -v
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)
	rm -f $(BINARY_DARWIN)
	rm -f testdata/*.tgz

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:
	$(GODEP)


# Go Cross-Compilation Tasks:
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -a -installsuffix cgo -o $(BINARY_LINUX) -v
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -a -installsuffix cgo -o $(BINARY_DARWIN) -v


# Docker Tasks:
## Use: make docker-build BINARY_NAME=barrelman REGISTRY=quay.io NAMESPACE=charter-se VERSION=v0.2.5 COMMIT=$(git rev-parse --short HEAD) BRANCH=$(git symbolic-ref -q --short HEAD) GOOS=linux GOARCH=amd64
docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg BRANCH=$(BRANCH) --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t $(REGISTRY)/$(NAMESPACE)/$(BINARY_NAME):$(VERSION) .

## Use: make docker-push BINARY_NAME=barrelman REGISTRY=quay.io NAMESPACE=charter-se VERSION=v0.2.5
docker-push:
	docker push $(REGISTRY)/$(NAMESPACE)/$(BINARY_NAME):$(VERSION)

## Use: make docker-push BINARY_NAME=barrelman REGISTRY=quay.io NAMESPACE=charter-se VERSION=v0.2.5 COMMIT=$(git rev-parse --short HEAD)
docker-push-commit:
	docker build $(REGISTRY)/$(NAMESPACE)/$(BINARY_NAME):$(VERSION)-$(COMMIT)

docker-test:
	docker build -f Dockertest .
