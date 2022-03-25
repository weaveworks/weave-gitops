package kube

import (
	"context"
	"fmt"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource interface {
	client.Object
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type WegoConfig struct {
	FluxNamespace string
	WegoNamespace string
	ConfigRepo    string
}

//counterfeiter:generate . Kube
type Kube interface {
	Apply(ctx context.Context, manifest []byte, namespace string) error
	GetClusterName(ctx context.Context) (string, error)
	GetWegoConfig(ctx context.Context, namespace string) (*WegoConfig, error)
	Raw() client.Client
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
