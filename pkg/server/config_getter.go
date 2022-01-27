package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigGetter implementations should extract the details from a context and
// create a *rest.Config for use in clients.
type ConfigGetter interface {
	Config(ctx context.Context) *rest.Config
}

// ImpersonatingConfigGetter is an implementation of the ConfigGetter interface
// that returns configs based on a base one. It inspects the context for a
// principal and if it finds one, it configures the *rest.Config to impersonate
// that principal. Otherwise it returns a copy of the base config.
type ImpersonatingConfigGetter struct {
	insecure bool
	cfg      *rest.Config
}

// NewImpersonatingConfigGetter creates and returns a ConfigGetter with a known
// config.
func NewImpersonatingConfigGetter(cfg *rest.Config, insecure bool) *ImpersonatingConfigGetter {
	return &ImpersonatingConfigGetter{cfg: cfg, insecure: insecure}
}

// Config returns a *rest.Config configured to impersonate a user or
// use the default service account credentials.
func (r *ImpersonatingConfigGetter) Config(ctx context.Context) *rest.Config {
	shallowCopy := *r.cfg

	if p := auth.Principal(ctx); p != nil {
		shallowCopy.Impersonate = rest.ImpersonationConfig{
			UserName: p.ID,
			Groups:   p.Groups,
		}
	}

	if r.insecure {
		shallowCopy.TLSClientConfig = rest.TLSClientConfig{
			Insecure: r.insecure,
		}
	}

	return &shallowCopy
}

// ClientGetter implementations should create a Kubernetes client from a context.
type ClientGetter interface {
	Client(ctx context.Context) (client.Client, error)
}

// DefaultClientGetter implements the ClientGetter interface and uses a ConfigGetter
// to get a *rest.Config and create a Kubernetes client.
type DefaultClientGetter struct {
	configGetter ConfigGetter
	clusterName  string
}

// Client creates a new Kubernetes client using the *rest.Config returned from its
// ConfigGetter.
func (g *DefaultClientGetter) Client(ctx context.Context) (client.Client, error) {
	config := g.configGetter.Config(ctx)

	_, rawClient, err := kube.NewKubeHTTPClientWithConfig(config, g.clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return rawClient, nil
}

// KubeGetter implementations should create a Kube client from a context.
type KubeGetter interface {
	Kube(ctx context.Context) (kube.Kube, error)
}

// DefaultKubeGetter implements the KubeGetter interface and uses a ConfigGetter
// to get a *rest.Config and create a Kube client.
type DefaultKubeGetter struct {
	configGetter ConfigGetter
	clusterName  string
}

// Kube creates a new Kube client using the *rest.Config returned from its
// ConfigGetter.
func (g *DefaultKubeGetter) Kube(ctx context.Context) (kube.Kube, error) {
	config := g.configGetter.Config(ctx)

	kube, _, err := kube.NewKubeHTTPClientWithConfig(config, g.clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not create kube http client: %w", err)
	}

	return kube, nil
}
