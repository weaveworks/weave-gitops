.PHONY: all test install clean fmt vet gitops gitops-server _docker docker-gitops docker-gitops-server lint ui ui-audit ui-lint ui-test unit-tests proto proto-deps fakes

CURRENT_DIR=$(shell pwd)

# Metadata for the builds. These can all be over-ridden so we can fix them in docker.
BUILD_TIME?=$(shell date +'%Y-%m-%d_%T')
BRANCH?=$(shell which git > /dev/null && git rev-parse --abbrev-ref HEAD)
GIT_COMMIT?=$(shell which git > /dev/null && git log -n1 --pretty='%h')
VERSION?=$(shell which git > /dev/null && git describe --always --match "v*")
FLUX_VERSION=2.6.4
CHART_VERSION=$(shell which yq > /dev/null && yq e '.version' charts/gitops-server/Chart.yaml)
TIER=oss

# Go build args
GOOS=$(shell which go > /dev/null && go env GOOS)
GOARCH=$(shell which go > /dev/null && go env GOARCH)
LDFLAGS?=-X github.com/weaveworks/weave-gitops/cmd/gitops/version.Branch=$(BRANCH) \
				 -X github.com/weaveworks/weave-gitops/cmd/gitops/version.BuildTime=$(BUILD_TIME) \
				 -X github.com/weaveworks/weave-gitops/cmd/gitops/version.GitCommit=$(GIT_COMMIT) \
				 -X github.com/weaveworks/weave-gitops/cmd/gitops/version.Version=$(VERSION) \
				 -X github.com/weaveworks/weave-gitops/pkg/version.FluxVersion=$(FLUX_VERSION) \
				 -X github.com/weaveworks/weave-gitops/pkg/analytics.Tier=$(TIER) \
				 -X github.com/weaveworks/weave-gitops/core/server.Branch=$(BRANCH) \
				 -X github.com/weaveworks/weave-gitops/core/server.Buildtime=$(BUILD_TIME) \
				 -X github.com/weaveworks/weave-gitops/core/server.GitCommit=$(GIT_COMMIT) \
				 -X github.com/weaveworks/weave-gitops/core/server.Version=$(VERSION) \
				 -X github.com/weaveworks/weave-gitops/cmd/gitops/beta/run.HelmChartVersion=$(CHART_VERSION)

# Docker args
# LDFLAGS is passed so we don't have to copy the entire .git directory into the image
# just to get, e.g. the commit hash
DOCKERARGS+=--build-arg FLUX_VERSION=$(FLUX_VERSION) --build-arg LDFLAGS="$(LDFLAGS)" --build-arg GIT_COMMIT=$(GIT_COMMIT)
# We want to be able to reference this in builds & pushes
DEFAULT_DOCKER_REPO=localhost:5001
DOCKER_IMAGE_TAG?=latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell which go > /dev/null && go env GOBIN))
GOBIN=$(shell which go > /dev/null && go env GOPATH)/bin
else
GOBIN=$(shell which go > /dev/null && go env GOBIN)
endif

# Make sure GOBIN is in PATH, so go install-ed binaries work
export PATH := $(PATH):$(GOBIN)

ifeq ($(BINARY_NAME),)
BINARY_NAME := gitops
endif

##@ Default target
all: gitops gitops-server ## Build Gitops binary. targets: gitops gitops-server

TEST_TO_RUN?=./...
TEST_V?=-v
##@ Test
unit-tests: ## Run unit tests
	# As long as we have a dependency to Ginko in the main module,
	# the following command should pick up the correct version from go.mod.
	@go install github.com/onsi/ginkgo/v2/ginkgo
	# This tool doesn't have releases - it also is only a shim
	@go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	KUBEBUILDER_ASSETS=$$(setup-envtest use -p path 1.32.0) CGO_ENABLED=1 ginkgo $(TEST_V) -race -tags unittest $(TEST_TO_RUN)

local-kind-cluster-with-registry:
	./tools/kind-with-registry.sh

local-registry:
	./tools/deploy-local-registry.sh

local-docker-image: docker-gitops-server

test: TEST_TO_RUN=./core/...
test: unit-tests

fakes: ## Generate testing fakes
	go generate ./...

install: bin ## Install binaries to GOPATH
	cp bin/$(BINARY_NAME) ${GOPATH}/bin/

cluster-dev: ## Start tilt to do development with wego-app running on the cluster
	./tools/bin/tilt up

clean-dev-cluster:
	kind delete cluster --name kind && docker rm -f kind-registry

##@ Build
# In addition to the main file depend on all go files
bin/%: cmd/%/main.go $(shell find . -name "*.go")
ifdef DEBUG
		CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $@ $(GO_BUILD_OPTS) $<
else
		CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -gcflags='all=-N -l' -o $@ $(GO_BUILD_OPTS) $<
endif

gitops: bin/gitops ## Build the Gitops CLI, accepts a 'DEBUG' flag

gitops-server: bin/gitops-server ## Build the Gitops UI server, accepts a 'DEBUG' flag

# Clean up images and binaries
clean: ## Clean up images and binaries
	# Clean up everything. This includes files git has been told to ignore (-x) and directories (-d)
	git clean -xfd -e .idea -e *.iml

fmt: ## Run go fmt against code
	go fmt ./...

vet: ## Run go vet against code
	go vet ./...

lint: golangci-lint protolint ## Run linters against code
	$(GOLANGCI_LINT) run
	$(PROTOLINT) lint -config_path=.protolint.yaml ./api

lint-fix: golangci-lint protolint ## Fix auto-fixable lint issues in code
	$(GOLANGCI_LINT) run --fix
	$(PROTOLINT) lint -fix -config_path=.protolint.yaml ./api

check-format:FORMAT_LIST=$(shell which gofmt > /dev/null && gofmt -l .)
check-format: ## Check go format
# The trailing `\` are important here as this is embedded bash and technically 1 line
	@if [ ! -z "$(FORMAT_LIST)" ] ; then \
		echo invalid format at: ${FORMAT_LIST} && exit 1; \
	fi

proto-deps: ## Update protobuf dependencies
	buf dep update

proto: ## Generate protobuf files
	@# The ones with no version use the library inside the code already
	@# so always use same version
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
	  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
	  google.golang.org/protobuf/cmd/protoc-gen-go
	@go install github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	@go install github.com/bufbuild/buf/cmd/buf@v1.48.0
	buf generate
#	This job is complaining about a missing plugin and error-ing out
#	oapi-codegen -config oapi-codegen.config.yaml api/applications/applications.swagger.json

# Sometimes we get whitespace differences when running this on linux vs mac
# So here's how you can do it under linux, on mac
proto-linux:
	docker run --rm -v "$(CURRENT_DIR):/app" -w /app golang:1.24.2 make proto

##@ Docker
_docker:
	DOCKER_BUILDKIT=1 docker build $(DOCKERARGS)\
										-f $(DOCKERFILE) \
										-t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
										.

docker-gitops: DOCKERFILE:=gitops.dockerfile
docker-gitops: DOCKER_IMAGE_NAME?=$(DEFAULT_DOCKER_REPO)/gitops
docker-gitops: _docker ## Build a Docker image of the gitops CLI

docker-gitops-server: DOCKERFILE:=gitops-server.dockerfile
docker-gitops-server: DOCKER_IMAGE_NAME?=$(DEFAULT_DOCKER_REPO)/gitops-server
docker-gitops-server: _docker ## Build a Docker image of the Gitops UI Server

##@ UI
# Build the UI for embedding
ui: node_modules $(shell find ui -type f) ## Build the UI
	yarn build

node_modules: ## Install node modules
	rm -rf .parcel-cache
	yarn --immutable --network-timeout 600000

ui-lint: ## Run linter against the UI
	yarn lint
	yarn typecheck

ui-lint-fix:
	yarn lint --fix

ui-prettify-check: ## Check format of the UI code with Prettier
	yarn prettify:check

ui-prettify-format: ## Format the UI code with Prettier
	yarn prettify:format

ui-test: ## Run UI tests
	yarn test

ui-audit: ## Run audit against the UI
	yarn npm audit --environment production

ui-audit-fix: ## Fix UI audit errors
	yarn yarn-audit-fix

# Build the UI as an NPM package (hosted on github)
ui-lib: node_modules dist/index.js dist/index.d.ts ## Build UI libraries
# Remove font files from the npm module.
	@find dist -type f -iname \*.otf -delete
	@find dist -type f -iname \*.woff -delete

dist/index.js: ui/index.ts
	yarn build:lib && cp package.json dist

dist/index.d.ts: ui/index.ts
	yarn typedefs

##@ Utilities
tls-files:
	@go install filippo.io/mkcert@v1.4.3
	mkcert localhost


# These echo commands exist to make it easier to pass stuff around github actions
echo-ldflags:
	@echo $(LDFLAGS)

echo-flux-version:
	@echo $(FLUX_VERSION)

download-test-crds:
	curl -sSL https://github.com/fluxcd/flux2/releases/download/v$(FLUX_VERSION)/install.yaml \
	| yq e '. | select(.kind == "CustomResourceDefinition")' \
	> tools/testcrds/flux.yaml

.PHONY: help
# Thanks to https://www.thapaliya.com/en/writings/well-documented-makefiles/
help:  ## Display this help.
ifeq ($(OS),Windows_NT)
				@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n make <target>\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-40s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
else
				@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
endif

## Tool Binaries

GOLANGCI_LINT_VERSION ?= v2.1.6
PROTOLINT_VERSION ?= v0.52.0

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT):
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

PROTOLINT = $(LOCALBIN)/protolint
protolint: $(PROTOLINT) ## Download protolint locally if necessary.
$(PROTOLINT):
	GOBIN=$(LOCALBIN) go install github.com/yoheimuta/protolint/cmd/protolint@${PROTOLINT_VERSION}
