.PHONY: ui-dev
VERSION=$(shell git describe --always --match "v*")
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

BUILD_TIME=$(shell date +'%Y-%m-%d_%T')
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT=$(shell git log -n1 --pretty='%h')
CURRENT_DIR=$(shell pwd)
FLUX_VERSION=$(shell $(CURRENT_DIR)/tools/bin/stoml $(CURRENT_DIR)/tools/dependencies.toml flux.version)
LDFLAGS = "-X github.com/weaveworks/weave-gitops/cmd/wego/version.BuildTime=$(BUILD_TIME) -X github.com/weaveworks/weave-gitops/cmd/wego/version.Branch=$(BRANCH) -X github.com/weaveworks/weave-gitops/cmd/wego/version.GitCommit=$(GIT_COMMIT) -X github.com/weaveworks/weave-gitops/pkg/version.FluxVersion=$(FLUX_VERSION) -X github.com/weaveworks/weave-gitops/cmd/wego/version.Version=$(VERSION)"

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

debug: 
	go build -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME) -gcflags='all=-N -l' cmd/wego/*.go 

bin:
	go build -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/wego/*.go

# Build wego binary
wego: dependencies bin

# Install binaries to GOPATH
install: bin bin/$(BINARY_NAME)_ui
	cp bin/$(BINARY_NAME) bin/$(BINARY_NAME)_ui ${GOPATH}/bin/

# Clean up images and binaries
clean:
	rm -f bin/wego pkg/flux/bin/flux
# Run go fmt against code
fmt:
	go fmt ./...
# Run go vet against code
vet:
	go vet ./...

dependencies:
	test -e pkg/flux/bin/flux || $(CURRENT_DIR)/tools/download-deps.sh $(CURRENT_DIR)/tools/dependencies.toml

package-lock.json:
	npm install

cmd/ui/dist:
	mkdir -p cmd/ui/dist

cmd/ui/dist/index.html: cmd/ui/dist
	touch cmd/ui/dist/index.html

cmd/ui/dist/main.js: package-lock.json
	npm run build

bin/$(BINARY_NAME)_ui: cmd/ui/main.go
	go build -ldflags $(LDFLAGS) -o bin/$(BINARY_NAME)_ui cmd/ui/main.go

lint:
	golangci-lint run --out-format=github-actions --build-tags acceptance

ui-lint:
	npm run lint

ui-test:
	npm run test

ui-audit:
	npm audit

# JS coverage info
coverage/lcov.info:
	npm run test -- --coverage

# Golang gocov data. Not compatible with coveralls at this point.
coverage.out:
	go get -u github.com/ory/go-acc
	go-acc --ignore fakes,acceptance,pkg/api -o coverage.out ./... -- -v --timeout=496s -tags test
	@go mod tidy

# Convert gocov to lcov for coveralls
coverage/golang.info: coverage.out
	@mkdir -p coverage
	@go get -u github.com/jandelgado/gcov2lcov
	gcov2lcov -infile=coverage.out -outfile=coverage/golang.info

# Concat the JS and Go coverage files for the coveralls report/
# Note: you need to install `lcov` to run this locally.
coverage/merged.lcov: coverage/lcov.info coverage/golang.info
	lcov --add-tracefile coverage/golang.info -a coverage/lcov.info -o merged.lcov

proto-deps:
	@go get \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts \
		github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	buf beta mod update

proto:
	buf generate
	oapi-codegen -config oapi-codegen.config.yaml api/applications/applications.swagger.json

api-dev:
	reflex -r '.go' -s -- sh -c 'go run cmd/wego-server/main.go'

ui-dev: cmd/ui/dist/main.js
	reflex -r '.go' -s -- sh -c 'go run cmd/ui/main.go'
