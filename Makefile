VERSION=$(shell git describe --always --match "v*")
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

BUILD_TIME=$(shell date +'%Y-%m-%d_%T')
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT=$(shell git log -n1 --pretty='%h')
CURRENT_DIR=$(shell pwd)
FLUX_VERSION=$(shell $(CURRENT_DIR)/tools/bin/stoml $(CURRENT_DIR)/tools/dependencies.toml flux.version)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ifeq ($(BINARY_NAME),)
BINARY_NAME := wego
endif

.PHONY: bin

all: wego

# Run tests
unit-tests:
	CGO_ENABLED=0 go test -v -tags unittest ./...

bin:
	go build -ldflags "-X github.com/weaveworks/weave-gitops/cmd/wego/version.BuildTime=$(BUILD_TIME) -X github.com/weaveworks/weave-gitops/cmd/wego/version.Branch=$(BRANCH) -X github.com/weaveworks/weave-gitops/cmd/wego/version.GitCommit=$(GIT_COMMIT) -X github.com/weaveworks/weave-gitops/pkg/version.FluxVersion=$(FLUX_VERSION)" -o bin/$(BINARY_NAME) cmd/wego/*.go

# Build wego binary
wego: dependencies bin

# Clean up images and binaries
clean:
	rm -f bin/wego
# Run go fmt against code
fmt:
	go fmt ./...
# Run go vet against code
vet:
	go vet ./...

dependencies:
	$(CURRENT_DIR)/tools/download-deps.sh $(CURRENT_DIR)/tools/dependencies.toml
