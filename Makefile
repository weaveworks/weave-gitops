VERSION=$(shell git describe --always --match "v*")
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
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
wego: fmt vet unit-tests
	CGO_ENABLED=0 go build -o bin/wego cmd/wego/*.go
# Clean up images and binaries
clean:
	rm -f bin/wego
# Run go fmt against code
fmt:
	go fmt ./...
# Run go vet against code
vet:
	go vet ./...