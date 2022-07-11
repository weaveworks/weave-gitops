package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/flux"
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

const (
	defaultVersion = ""
)

func (cs *coreServer) getScopedClient(ctx context.Context) (client.Client, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	scopedClient, err := clustersClient.Scoped(clustersmngr.DefaultCluster)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client: %w", err)
	}

	return scopedClient, nil
}

func (cs *coreServer) getFluxVersion(ctx context.Context, k8sClient client.Client) (string, error) {
	listResult := unstructured.UnstructuredList{}

	listResult.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Namespace",
	})

	opts := client.MatchingLabels{
		coretypes.PartOfLabel: FluxNamespacePartOf,
	}

	u := unstructured.Unstructured{}

	err := k8sClient.List(ctx, &listResult, opts)
	if err != nil {
		return "", fmt.Errorf("error getting list of objects")
	} else {
		for _, item := range listResult.Items {
			if item.GetLabels()[flux.VersionLabelKey] != "" {
				u = item
				break
			}
		}
	}

	labels := u.GetLabels()
	if labels == nil {
		return "", fmt.Errorf("error getting labels")
	}

	fluxVersion := labels[flux.VersionLabelKey]
	if fluxVersion == "" {
		return "", fmt.Errorf("no flux version found")
	}

	return fluxVersion, nil
}

func (cs *coreServer) getKubeVersion(ctx context.Context) (string, error) {
	dc, err := cs.clientsFactory.GetImpersonatedDiscoveryClient(ctx, auth.Principal(ctx), clustersmngr.DefaultCluster)
	if err != nil {
		return "", fmt.Errorf("error creating discovery client: %w", err)
	}

	serverVersion, err := dc.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("error getting server version: %w", err)
	} else {
		return serverVersion.GitVersion, nil
	}
}

func (cs *coreServer) GetVersion(ctx context.Context, msg *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	k8sClient, err := cs.getScopedClient(ctx)
	if err != nil {
		cs.logger.Error(err, "error creating scoped client")
	}

	fluxVersion, err := cs.getFluxVersion(ctx, k8sClient)
	if err != nil {
		cs.logger.Error(err, "error getting flux version")

		fluxVersion = defaultVersion
	}

	kubeVersion, err := cs.getKubeVersion(ctx)
	if err != nil {
		cs.logger.Error(err, "error getting k8s version")

		kubeVersion = defaultVersion
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
