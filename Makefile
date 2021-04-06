VERSION=$(shell git describe --always --match "v*")
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

BUILD_TIME=$(shell date +'%Y-%m-%d_%T')
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT=$(shell git log -n1 --pretty='%h')
CURRENT_DIR := $(shell pwd)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: wego

# Run tests
unit-tests:
	CGO_ENABLED=0 go test -v ./cmd/...

# Build wego binary
wego: flux fmt vet unit-tests
	go build -ldflags "-X github.com/weaveworks/weave-gitops/cmd/wego/version.BuildTime=$(BUILD_TIME) -X github.com/weaveworks/weave-gitops/cmd/wego/version.Branch=$(BRANCH) -X github.com/weaveworks/weave-gitops/cmd/wego/version.GitCommit=$(GIT_COMMIT)" -o bin/wego cmd/wego/*.go
	rm -f cmd/wego/flux/flux
# Clean up images and binaries
clean:
	rm -f bin/wego
# Download flux
flux:
	curl -s https://toolkit.fluxcd.io/install.sh | bash
	cp /usr/local/bin/flux ./cmd/wego/flux
# Run go fmt against code
fmt:
	go fmt ./...
# Run go vet against code
vet:
	go vet ./...

dependencies:
	$(CURRENT_DIR)/tools/download-deps.sh $(CURRENT_DIR)/tools/dependencies.toml
