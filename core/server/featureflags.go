package server

import (
	"context"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	serverauth "github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) GetFeatureFlags(ctx context.Context, msg *pb.GetFeatureFlagsRequest) (*pb.GetFeatureFlagsResponse, error) {
	flags := make(map[string]string)

	flags["WEAVE_GITOPS_AUTH_ENABLED"] = os.Getenv("WEAVE_GITOPS_AUTH_ENABLED")

	cl, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get Kubernetes client from context: %w", err)
	}

	var secret corev1.Secret
	err = cl.Get(ctx, client.ObjectKey{
		Namespace: v1alpha1.DefaultNamespace,
		Name:      serverauth.ClusterUserAuthSecretName,
	}, &secret)

	if err != nil {
		if apierrors.IsNotFound(err) {
			flags["CLUSTER_USER_AUTH"] = "false"
		} else {
			cs.logger.Error(err, "could not get secret for cluster user")
		}
	} else {
		flags["CLUSTER_USER_AUTH"] = "true"
	}

	err = cl.Get(ctx, client.ObjectKey{
		Namespace: v1alpha1.DefaultNamespace,
		Name:      serverauth.OIDCAuthSecretName,
	}, &secret)

	if err != nil {
		if apierrors.IsNotFound(err) {
			flags["OIDC_AUTH"] = "false"
		} else {
			cs.logger.Error(err, "could not get secret for oidc")
		}
	} else {
		flags["OIDC_AUTH"] = "true"
	}

	return &pb.GetFeatureFlagsResponse{
		Flags: flags,
	}, nil
}
