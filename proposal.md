## Weave GitOps API Proposal

### Goal

Demonstrate an end-to-end proof-of-concept for an entire HTTP API workflow

Optimizing for:

- Ease of development (Weaveworks)
- Ease of adoption (Customers)

4. Generated API docs
5. Middlware
6. Auth

Protobuf Benefits

- The discussion around HTTP verbs/params/query goes away. Everything is a POST
- Generated types and clients ensure both server and client are in sync
- Compile-time warnings

#### 1. Protocol Definition

```protobuf
syntax = "proto3";
package gitops;

option go_package = "pkg/rpc/gitops";
option java_outer_classname = "GitOpsProtos";

service GitOps {
    // Lists the applications that have been GitOps-ified
    rpc AddApplication(AddApplicationReq) returns (AddApplicationRes);
}

enum DeploymentType {
    kustomize = 0;
    helm = 1;
}


// Represents a Weave GitOps application
message Application {
    string name = 1;
    DeploymentType type = 2;
}

message AddApplicationReq {
    string         owner           = 1;   // The owner or org of the source git repo
    string         name            = 2;   // The Name of the application
    string         url             = 3;   // The URL of the source git repository
    string         path            = 4;   // The path within the git repository that holds the files
    string         branch          = 5;   // The git branch to use
    DeploymentType deployment_type = 6;   // The type of deployment
    string         private_key     = 7;   // The private key for creating this repo
    bool           dry_run         = 8;   // Whether or not to do a dry run
    bool           private         = 9;   // Whether or not the repo is private
    string         namespace       = 10;  // The target namespace for this application
    string         dir             = 11;  // The ...dir?
}

message AddApplicationRes {
    Application application = 1;
}

message AddApplicationRes {
    Application application = 1;
}
```

#### 2. Server Implementation

```golang
// ./pkg/server/server.go
package server

import (
    "context"
	"fmt"
	"net/http"

    "github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"
    pb "github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
    "github.com/weaveworks/weave-gitops/pkg/middleware"
)


type Server struct {}

func NewServer() http.Handler {
	defaultHooks := twirp.ChainHooks(middleware.MetricsHooks())

	gitops := Server{}

	s := pb.NewGitOpsServer(&gitops, defaultHooks)

	return s
}

func (s *Server) AddApplication(ctx context.Context, msg *pb.AddApplicationReq) (*pb.AddApplicationRes, error) {

	params := cmdimpl.AddParamSet{
		Dir:            msg.Dir,
		Name:           msg.Name,
		Owner:          msg.Owner,
		Url:            msg.Url,
		Path:           msg.Path,
		Branch:         msg.Branch,
		PrivateKey:     msg.PrivateKey,
		DeploymentType: msg.DeploymentType.String(),
		Namespace:      msg.Namespace,
	}

	if err := cmdimpl.Add([]string{msg.Dir}, params); err != nil {
		return nil, fmt.Errorf("could not add application: %v", err)
	}

	return &pb.AddApplicationRes{Application: &pb.Application{Name: msg.Name}}, nil
}

```

```golang
// ./cmd/server/main.go
package main

import (
    "github.com/weaveworks/weave-gitops/pkg/server"
)

func main() {
    mux := http.NewServeMux()

    gitopsServer := server.NewServer()
	mux.Handle("/api/gitops/", http.StripPrefix("/api/gitops", gitopsServer))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error(err, "server exited")
		os.Exit(1)
	}
}
```

#### 3. Generated clients and tests

```golang
// pkg/server/server_test.go
package server_test

import (
    pb "github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

func Test_AddApplication(t *testing.T) {
	client := createAuthenticatedClient(t)
	server := http.Server{
        Addr: ":50051",
        Handler: server.NewServer(),
    }
	ctx := context.Background()

	defer server.Shutdown(ctx)

    name := "my-cool-app"

	res, err := client.AddApplication(ctx, &pb.AddApplicationReq{
		Name:           name,
		Owner:          "jpellizzari",
		Url:            "https://github.com/stefanprodan/podinfo.git",
		Path:           "./kustomize",
		Branch:         "main",
		PrivateKey:     "",
		DeploymentType: pb.DeploymentType_kustomize,
		Namespace:      "default",
		DryRun:         true,
		Dir:            "./",
	})

	if err != nil {
		t.Fatal(err)
	}

	if res.Application.Name != "name" {
		t.Fatal(errors.New("expected name to be correct"))
	}
}
```

4. Generated API Docs
   ![](images/docs.png?raw=true)
