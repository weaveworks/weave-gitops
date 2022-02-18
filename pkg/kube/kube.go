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
	ConfigRepo    string
}

//counterfeiter:generate . Kube
type Kube interface {
	Apply(ctx context.Context, manifest []byte, namespace string) error
	Delete(ctx context.Context, manifest []byte) error
	DeleteByName(ctx context.Context, name string, gvr schema.GroupVersionResource, namespace string) error
	SecretPresent(ctx context.Context, string, namespace string) (bool, error)
	FluxPresent(ctx context.Context) (bool, error)
	NamespacePresent(ctx context.Context, namespace string) (bool, error)
	GetClusterName(ctx context.Context) (string, error)
	GetClusterStatus(ctx context.Context) ClusterStatus
	GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error)
	GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error
	SetResource(ctx context.Context, resource Resource) error
	GetSecret(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error)
	FetchNamespaceWithLabel(ctx context.Context, key string, value string) (*corev1.Namespace, error)
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

// KubeGetter implementations should create a Kube client from a context.
type KubeGetter interface {
	Kube(ctx context.Context) (Kube, error)
}

var _ KubeGetter = &DefaultKubeGetter{}

// DefaultKubeGetter implements the KubeGetter interface and uses a ConfigGetter
// to get a *rest.Config and create a Kube client.
type DefaultKubeGetter struct {
	configGetter ConfigGetter
	clusterName  string
}

// NewDefaultKubeGetter creates a new DefaultKubeGetter
func NewDefaultKubeGetter(configGetter ConfigGetter, clusterName string) KubeGetter {
	return &DefaultKubeGetter{
		configGetter: configGetter,
		clusterName:  clusterName,
	}
}

// Kube creates a new Kube client using the *rest.Config returned from its
// ConfigGetter.
func (g *DefaultKubeGetter) Kube(ctx context.Context) (Kube, error) {
	config := g.configGetter.Config(ctx)

	kube, _, err := NewKubeHTTPClientWithConfig(config, g.clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return kube, nil
}
