# 2. api-definitions-grpc

Date: 2021-06-24

## Status

Proposed

## Context

**THIS ADR DOES NOT COVER AUTHN or AUTHZ.**

Our goal is to standardise how APIs are built.

First of all, let's establish what an "API" is, for our purposes an API is an
interface to the functionality provided by a service, our UX is composed of a
web or cli-based user-interface that talks to a shared backend service.

This backend provides an API to make it easy for both developers, and users to
interact with the functionality provided by the service.

As we build more and more services, having a consistent API scheme to talk to
these becomes more important.

Currently, we hand-roll APIs, but Hand-rolling APIs is pretty easy for
single-endpoints, but harder for multiple-endpoints as you have to maintain
consistency across them, clients have to be built for each endpoint, possibly
across multiple languages, and documentation generally has to be written
manually, this has a cost, and leads to out-of-date documentation, less
certainty for customers ("Is this a bug in the client we implemented or the
server?") and overall provides a suboptimal experience.

The goals for APIs are:

 * Minimise cost for customers who want to interact with the API in their language of choice
 * Minimise cost of maintaining documentation for us
 * Be declarative and uniform across components

## Decision

TLDR; Use protobufs for everything!

This proposal seeks to standardise our request/response APIs on Google's
["protocol buffers"](https://developers.google.com/protocol-buffers) (protobuf) syntax.

This means declaratively describing the request and response structures, and the
"service" API in a `.proto` file.

## Tooling

### Buf

This proposal recommends the use of a tool [buf](https://buf.build/) that can be used to
automate the generation of stubs and files from protobuf formats.

`buf.yaml`
```yaml
version: v1beta1
name: buf.build/weaveworks/api-proposal
deps:
  - buf.build/beta/googleapis
  - buf.build/grpc-ecosystem/grpc-gateway
build:
  roots:
    - ./api
lint:
  except:
    - PACKAGE_DIRECTORY_MATCH
    - PACKAGE_SAME_DIRECTORY
```

The two lint exceptions mean that the `.proto` can be in `/api` rather than in a
deep subtree of the directory.

The dependencies will be used in a later example.

The real power of the buf generation mechanism comes from the file
`buf.gen.yaml`.

```yaml
version: v1beta1
plugins:
  - name: go
    out: pkg/protos
    opt: paths=source_relative
  - name: go-grpc
    out: pkg/protos
    opt: paths=source_relative,require_unimplemented_servers=false
  - name: doc
    out: docs
```

This outputs `docs/index.html`, a full set of protobuf structs, and an example
server that can can be implemented to actually provide the business logic.

### Service implementation

The generated stubs have a _not implemented_ stub that returns an error for each
unimplemented service method, but implementing the described service description
is fairly easy.

```go
func (s grpcServer) RenderTemplate(ctx context.Context, r *capiv1.RenderTemplateRequest) (*capiv1.RenderTemplateResponse, error) {
    rendered, err := s.renderer.Render(r.Flavor, r.Params)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error rendering template: %s", err)
	}

	return &capiv1.RenderTemplateResponse{
		Content: &capiv1.Content{
			Encoding: "base64",
			Body:     base64.StdEncoding.EncodeToString(rendered),
		},
	}, nil
}
```

### grpcurl

This endpoint can be called via grpcurl as long as reflection is turned on:
```shell
$ grpcurl -d '{"flavor": "testing-template", "params": {"CLUSTER_NAME": "my-test-cluster"}}' -plaintext localhost:50051 weave.works.wego.capi.v1.CAPIService/RenderTemplate
{
  "content": {
    "encoding": "base64",
    "body": "dmFsdWU6IG15LXRlc3QtY2x1c3Rlcg=="
  }
}
```

With reflection, it's easy to see the exposed API from the service.
```shell
$ grpcurl -plaintext localhost:50051 list weave.works.wego.capi.v1.CAPIService
weave.works.wego.capi.v1.CAPIService.RenderTemplate
```

And describe the operations...
```shell
$ grpcurl -plaintext localhost:50051 describe weave.works.wego.capi.v1.CAPIService.RenderTemplate
weave.works.wego.capi.v1.CAPIService.RenderTemplate is a method:
rpc RenderTemplate ( .weave.works.wego.capi.v1.RenderTemplateRequest ) returns ( .weave.works.wego.capi.v1.RenderTemplateResponse );
$ grpcurl -plaintext localhost:50051 describe weave.works.wego.capi.v1.RenderTemplateRequest
weave.works.wego.capi.v1.RenderTemplateRequest is a message:
message RenderTemplateRequest {
  string flavor = 1;
  map<string, string> params = 2;
}
```

### Standardising gRPC services

Any exposed gRPC services should have reflection turned on, and use the standard
prometheus metrics mechanism.

```go
package main

import (
	"log"
	"net"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bigkevmcd/api-proposal/pkg/server"
)

const port = ":50051"

func main() {
	// this is just a fake for demo purposes.
	testConfig := &server.Config{
		FlavorTemplates: map[string]string{
			"testing-template": "value: ${CLUSTER_NAME}\n",
		},
	}

	srv, err := server.NewGRPCServer(testConfig,
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	if err != nil {
		log.Fatal(err)
	}
	reflection.Register(srv)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Omitted for brevity - exposing the prometheus metrics via HTTP.
	log.Println("Listening at localhost:50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

### grpc-gateway

One downside of gRPC is that it relies on HTTP/2 and so isn't always trivial to
deliver requests to a running container, also support from some languages
(including from the browers) isn't always easy.

This along with some annotations to the server exposes the same service via
HTTP.
```yaml
version: v1beta1
plugins:
  - name: go
    out: pkg/protos
    opt: paths=source_relative
  - name: go-grpc
    out: pkg/protos
    opt: paths=source_relative,require_unimplemented_servers=false
  - name: doc
    out: docs
  - name: grpc-gateway
    out: pkg/protos
    opt: paths=source_relative
```

```protobuf
import "google/api/annotations.proto";

// CAPIService provides functionality for the templating service.
service CAPIService {
  /**
  * RenderTemplate combines the flavour and params in the request and renders
  * the template.
  */
  rpc RenderTemplate(RenderTemplateRequest) returns (RenderTemplateResponse) {
    option (google.api.http) = {
      post: "/v1/flavors/{flavor=*}/render"
      body: "*"
    };
  }
}
```

The `"post: /v1/flavors/{flavor=*}/render"`annotation indicates that it accepts
POST requests, to `/v1/flavors/my-template/render` which is in keeping with
operations like `/v1/flavors/my-template` to get the unrendered body.
```shell
$ curl -H "Content-Type: application/json" -d '{"params": {"CLUSTER_NAME": "my-test-cluster"}}' localhost:9090/v1/flavors/testing-template/render
{
  "content": {
    "encoding": "base64",
    "body": "dmFsdWU6IG15LXRlc3QtY2x1c3Rlcgo="
  }
}
```

This is easy to consume from most languages.

Exposing this from code is fairly simple, it's pretty much all boilerplate,
including applying prometheus metrics around the service.

```go
	mux := runtime.NewServeMux()
	httpmux := http.NewServeMux()
	httpmux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	httpmux.Handle("/", promhttp.InstrumentMetricHandler(reg, mux))

	err := capiv1.RegisterCAPIServiceHandlerServer(ctx, mux, srv)
	if err != nil {
		return err
	}

	log.Println("Listening at :9091")
	return http.ListenAndServe(":9091", httpmux)
```

### OpenAPI/swagger

OpenAPI is exposed by Kubernetes, and is a popular format for describing APIs.

There is tooling for automatically generating clients for a lot of languages,
and also for calling an API interactively.

Again, by simply layering configuration onto the original protobuf declaration,
and adding to the buf.gen.yaml, we can generate a compliant spec.

```yaml
version: v1beta1
plugins:
  - name: go
    out: pkg/protos
    opt: paths=source_relative
  - name: go-grpc
    out: pkg/protos
    opt: paths=source_relative,require_unimplemented_servers=false
  - name: doc
    out: docs
  - name: grpc-gateway
    out: pkg/protos
    opt: paths=source_relative
  - name: openapiv2
    out: api
```

```protobuf
import "protoc-gen-openapiv2/options/annotations.proto";

option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "CAPI Templating API",
    version: "0.1";
    description: "CAPI Templating as a Service";
  };
  consumes: "application/json";
  produces: "application/json";
};
```

## Consequences

This simplifies the process of exposing functionality via an HTTP API, to the
process of writing the `.proto` file, add the annotations and run buf to
generate the Go protobuf files and other library files and documentation.

The `main` for the API server command becomes fairly boilerplate, creating the
gRPC server, providing it with dependencies, and wrapping it via the
grpc-gateway and logging & metrics middlewares.

Implementing the actual functionality is reduced to implementing a _Service_
implementation, which provides the actual functionality, talking to the
underlying code, and translating incoming requests, and populating outgoing
response values.

All APIs should at least initially be exposed over HTTP (and TLS), rather than
the default gRPC, but it's fairly trivial to expose the same functionality over
gRPC.
