package server

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
)

var _ kube.ConfigGetter = &ImpersonatingConfigGetter{}

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
