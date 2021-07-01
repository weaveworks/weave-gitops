// +build tools

package tools

import (
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	_ "github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts"
	_ "github.com/jandelgado/gcov2lcov"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
