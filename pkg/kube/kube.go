package kube

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/logger"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource interface {
	client.Object
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ClusterStatus int

const (
	Unknown ClusterStatus = iota
	Unmodified
	FluxInstalled
	GitOpsInstalled
)

// Function to translate ClusterStatus to a string
func (cs ClusterStatus) String() string {
	return toStatusString[cs]
}

var toStatusString = map[ClusterStatus]string{
	Unknown:         "Unable to talk to the cluster",
	Unmodified:      "No flux or gitops installed",
	FluxInstalled:   "Flux installed",
	GitOpsInstalled: "GitOps installed",
}

type WegoConfig struct {
	FluxNamespace string
	WegoNamespace string
}

//counterfeiter:generate . Kube
type Kube interface {
	Apply(ctx context.Context, manifest []byte, namespace string) error
	Delete(ctx context.Context, manifest []byte) error
	DeleteByName(ctx context.Context, name string, gvr schema.GroupVersionResource, namespace string) error
	SecretPresent(ctx context.Context, string, namespace string) (bool, error)
	GetApplications(ctx context.Context, namespace string) ([]wego.Application, error)
	FluxPresent(ctx context.Context) (bool, error)
	NamespacePresent(ctx context.Context, namespace string) (bool, error)
	GetClusterName(ctx context.Context) (string, error)
	GetClusterStatus(ctx context.Context) ClusterStatus
	GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error)
	GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error
	SetResource(ctx context.Context, resource Resource) error
	GetSecret(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error)
	FetchNamespaceWithLabel(ctx context.Context, key string, value string) (string, error)
	SetWegoConfig(ctx context.Context, config WegoConfig, namespace string) (*corev1.ConfigMap, error)
	GetWegoConfig(ctx context.Context, namespace string) (*WegoConfig, error)
	Raw() client.Client
}

func IsClusterReady(l logger.Logger, k Kube) error {
	l.Waitingf("Checking cluster status")

	clusterStatus := k.GetClusterStatus(context.Background())

	switch clusterStatus {
	case Unmodified:
		return fmt.Errorf("gitops not installed... exiting")
	case Unknown:
		return fmt.Errorf("can not determine cluster status... exiting")
	}

	l.Successf(clusterStatus.String())

	return nil
}
