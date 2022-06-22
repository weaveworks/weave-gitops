package server

import (
	"context"
	"fmt"

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

	key := client.ObjectKey{
		Name: "flux-system",
	}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	c, err2 := clustersClient.Scoped(clustersmngr.DefaultCluster)
	if err2 != nil {
		return nil, fmt.Errorf("error getting scoped client: %w", err2)
	}

	err3 := c.Get(ctx, key, u)
	if err3 != nil {
		return nil, fmt.Errorf("error getting object: %w", err3)
	}

	fmt.Println("unstructured object:")
	fmt.Println(u)
	fmt.Println("labels:")
	fmt.Println(u.GetLabels())

	// dc := discovery.NewDiscoveryClient(c)
	// dc, err4 := discovery.NewDiscoveryClientForConfig(cfg(clustersmngr.Cluster{
	// 	Name:      clustersmngr.DefaultCluster,
	// 	Server:    k8sEnv.Rest.Host,
	// 	TLSConfig: k8sEnv.Rest.TLSClientConfig,
	// }))

	// fmt.Println("error:")
	// fmt.Println(err4)
	// fmt.Println("discovery client:")
	// fmt.Println(dc)
	// fmt.Println("server version:")

	// v, _ := dc.ServerVersion()

	// fmt.Println(v)

	return &pb.GetVersionResponse{
		Semver:      Version,
		Commit:      GitCommit,
		Branch:      Branch,
		BuildTime:   Buildtime,
		FluxVersion: "",
		KubeVersion: "",
	}, nil
}
