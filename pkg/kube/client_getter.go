package kube

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClientGetter implementations should create a Kubernetes client from a context.
type ClientGetter interface {
	Client(ctx context.Context) (client.Client, error)
}

var _ ClientGetter = &DefaultClientGetter{}

// DefaultClientGetter implements the ClientGetter interface and uses a ConfigGetter
// to get a *rest.Config and create a Kubernetes client.
type DefaultClientGetter struct {
	configGetter ConfigGetter
	clusterName  string
}

// NewDefaultClientGetter creates a new DefaultClientGetter
func NewDefaultClientGetter(configGetter ConfigGetter, clusterName string) ClientGetter {
	return &DefaultClientGetter{
		configGetter: configGetter,
		clusterName:  clusterName,
	}
}

// Client creates a new Kubernetes client using the *rest.Config returned from its
// ConfigGetter.
func (g *DefaultClientGetter) Client(ctx context.Context) (client.Client, error) {
	config := g.configGetter.Config(ctx)

	_, rawClient, err := NewKubeHTTPClientWithConfig(config, g.clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return rawClient, nil
}
