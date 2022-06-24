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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Variables that we'll set @ build time
var (
	Version   = "v0.0.0"
	GitCommit = ""
	Branch    = ""
	Buildtime = ""
)

const (
	defaultVersion = ""
)

func (cs *coreServer) GetFluxVersion(ctx context.Context, key types.NamespacedName) (string, error) {
	u := &unstructured.Unstructured{}

	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Namespace",
	})

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return defaultVersion, fmt.Errorf("error getting impersonating client: %w", err)
	}

	c, err := clustersClient.Scoped(clustersmngr.DefaultCluster)
	if err != nil {
		return defaultVersion, fmt.Errorf("error getting scoped client: %w", err)
	}

	err = c.Get(ctx, key, u)
	if err != nil {
		return defaultVersion, fmt.Errorf("error getting object: %w", err)
	} else {
		return u.GetLabels()["app.kubernetes.io/version"], nil
	}
}

func (cs *coreServer) GetKubeVersion(ctx context.Context, key types.NamespacedName) (string, error) {
	dc, err := cs.clientsFactory.GetImpersonatedDiscoveryClient(ctx, auth.Principal(ctx), clustersmngr.DefaultCluster)
	if err != nil {
		return defaultVersion, fmt.Errorf("error creating discovery client: %w", err)
	}

	sVersion, err := dc.ServerVersion()
	if err != nil {
		return defaultVersion, fmt.Errorf("error getting server version: %w", err)
	} else {
		return sVersion.GitVersion, nil
	}
}

func (cs *coreServer) GetVersion(ctx context.Context, msg *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	var ns string

	if msg.Namespace != "" {
		ns = msg.Namespace
	} else {
		ns = wego.DefaultNamespace
	}

	key := client.ObjectKey{
		Name: ns,
	}

	fluxVersion, err := cs.GetFluxVersion(ctx, key)
	if err != nil {
		cs.logger.Error(err, "error getting flux version")
	}

	kubeVersion, err := cs.GetKubeVersion(ctx, key)
	if err != nil {
		cs.logger.Error(err, "error getting k8s version")
	}

	return &pb.GetVersionResponse{
		Semver:      Version,
		Commit:      GitCommit,
		Branch:      Branch,
		BuildTime:   Buildtime,
		FluxVersion: fluxVersion,
		KubeVersion: kubeVersion,
	}, nil
}
