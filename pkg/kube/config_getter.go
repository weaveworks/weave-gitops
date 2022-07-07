package kube

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
)

// ConfigGetter implementations should extract the details from a context and
// create a *rest.Config for use in clients.
type ConfigGetter interface {
	Config(ctx context.Context) *rest.Config
}

var _ ConfigGetter = &ImpersonatingConfigGetter{}

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

	var hasToken bool

	if t := auth.BearerToken(ctx); len(t) != 0 {
		shallowCopy.BearerToken = t
		hasToken = true
	}

	if p := auth.Principal(ctx); p != nil && !hasToken {
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
