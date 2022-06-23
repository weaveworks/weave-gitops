package server

import (
	"context"
	"fmt"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Variables that we'll set @ build time
var (
	Version   = "v0.0.0"
	GitCommit = ""
	Branch    = ""
	Buildtime = ""
)

func (cs *coreServer) GetVersion(ctx context.Context, msg *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	u := &unstructured.Unstructured{}

	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Namespace",
	})

	fmt.Println("msg.Namespace")
	fmt.Println(msg.Namespace)

	var ns string

	if msg.Namespace != "" {
		ns = msg.Namespace
	} else {
		ns = wego.DefaultNamespace
	}

	key := client.ObjectKey{
		Name: ns,
	}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	c, err := clustersClient.Scoped(clustersmngr.DefaultCluster)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client: %w", err)
	}

	err = c.Get(ctx, key, u)
	if err != nil {
		return nil, fmt.Errorf("error getting object: %w", err)
	}

	fmt.Println("unstructured object:")
	fmt.Println(u)
	fmt.Println("labels:")
	fmt.Println(u.GetLabels())

	dc, err := cs.clientsFactory.GetImpersonatedDiscoveryClient(ctx, auth.Principal(ctx), clustersmngr.DefaultCluster)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %w", err)
	}

	fmt.Println("error:")
	fmt.Println(err)
	fmt.Println("discovery client:")
	fmt.Println(dc)
	fmt.Println("server version:")

	v, _ := dc.ServerVersion()

	fmt.Println(v)

	return &pb.GetVersionResponse{
		Semver:      Version,
		Commit:      GitCommit,
		Branch:      Branch,
		BuildTime:   Buildtime,
		FluxVersion: "",
		KubeVersion: "",
	}, nil
}
