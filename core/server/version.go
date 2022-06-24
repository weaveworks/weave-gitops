package server

import (
	"context"

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

	var ns string

	if msg.Namespace != "" {
		ns = msg.Namespace
	} else {
		ns = wego.DefaultNamespace
	}

	key := client.ObjectKey{
		Name: ns,
	}

	fluxVersion := ""

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		cs.logger.Error(err, "error getting impersonating client")
	} else {
		c, err := clustersClient.Scoped(clustersmngr.DefaultCluster)
		if err != nil {
			cs.logger.Error(err, "error getting scoped client")
		} else {
			err = c.Get(ctx, key, u)
			if err != nil {
				cs.logger.Error(err, "error getting object")
			} else {
				fluxVersion = u.GetLabels()["app.kubernetes.io/version"]
			}
		}
	}

	kubeVersion := ""

	dc, err := cs.clientsFactory.GetImpersonatedDiscoveryClient(ctx, auth.Principal(ctx), clustersmngr.DefaultCluster)
	if err != nil {
		cs.logger.Error(err, "error creating discovery client")
	} else {
		sVersion, err := dc.ServerVersion()

		if err != nil {
			cs.logger.Error(err, "error getting server version")
		} else {
			kubeVersion = sVersion.GitVersion
		}
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
