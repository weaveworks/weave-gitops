package kube

import (
	"context"

	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// UserPrefixes contains the prefixes for the user and groups
// that will be used when impersonating a user.
type UserPrefixes struct {
	UsernamePrefix string
	GroupsPrefix   string
}

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
	insecure     bool
	cfg          *rest.Config
	userPrefixes UserPrefixes
}

// NewImpersonatingConfigGetter creates and returns a ConfigGetter with a known
// config.
func NewImpersonatingConfigGetter(cfg *rest.Config, insecure bool, userPrefixes UserPrefixes) *ImpersonatingConfigGetter {
	return &ImpersonatingConfigGetter{cfg: cfg, insecure: insecure, userPrefixes: userPrefixes}
}

// Config returns a *rest.Config configured to impersonate a user or
// use the default service account credentials.
func (r *ImpersonatingConfigGetter) Config(ctx context.Context) *rest.Config {
	cfg := rest.CopyConfig(r.cfg)

	if p := auth.Principal(ctx); p != nil {
		cfg = ConfigWithPrincipal(p, cfg, r.userPrefixes)
	}

	if r.insecure {
		cfg.TLSClientConfig = rest.TLSClientConfig{
			Insecure: r.insecure,
		}
	}

	return cfg
}

// ConfigWithPrincipal returns a new config with the principal set as the
// impersonated user or bearer token.
func ConfigWithPrincipal(user *auth.UserPrincipal, config *rest.Config, userPrefixes UserPrefixes) *rest.Config {
	cfg := rest.CopyConfig(config)

	if tok := user.Token(); tok != "" {
		cfg.BearerToken = tok
		// Clear the token file as it takes precedence over the token.
		cfg.BearerTokenFile = ""
	} else {
		prefixedGroups := user.Groups
		if prefixedGroups != nil {
			prefixedGroups = addGroupsPrefix(user.Groups, userPrefixes.GroupsPrefix)
		}

		cfg.Impersonate = rest.ImpersonationConfig{
			UserName: userPrefixes.UsernamePrefix + user.ID,
			Groups:   prefixedGroups,
		}
	}

	return cfg
}

func addGroupsPrefix(groups []string, prefix string) []string {
	prefixedGroups := make([]string, len(groups))

	for i, group := range groups {
		prefixedGroups[i] = prefix + group
	}

	return prefixedGroups
}
