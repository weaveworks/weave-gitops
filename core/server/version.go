package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
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

func (cs *coreServer) getKubeVersion(ctx context.Context) (string, error) {
	dc, err := cs.clustersManager.GetImpersonatedDiscoveryClient(ctx, auth.Principal(ctx), cluster.DefaultCluster)
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
		KubeVersion: kubeVersion,
	}, nil
}
