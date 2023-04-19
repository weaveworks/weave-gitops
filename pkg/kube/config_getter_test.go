package kube_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
)

var _ kube.ConfigGetter = (*kube.ImpersonatingConfigGetter)(nil)

func TestImpersonatingConfigGetterPrincipalInContext(t *testing.T) {
	g := kube.NewImpersonatingConfigGetter(&rest.Config{}, false, kube.UserPrefixes{})
	ctx := auth.WithPrincipal(context.TODO(), &auth.UserPrincipal{ID: "user@example.com"})

	cfg := g.Config(ctx)

	want := &rest.Config{
		Impersonate: rest.ImpersonationConfig{
			UserName: "user@example.com",
		},
	}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Fatalf("incorrect client config:\n%s", diff)
	}
}

func TestImpersonatingConfigGetterPrincipalInContextWithGroups(t *testing.T) {
	g := kube.NewImpersonatingConfigGetter(&rest.Config{}, false, kube.UserPrefixes{})
	ctx := auth.WithPrincipal(context.TODO(), &auth.UserPrincipal{ID: "user@example.com", Groups: []string{"test-group"}})

	cfg := g.Config(ctx)

	want := &rest.Config{
		Impersonate: rest.ImpersonationConfig{
			UserName: "user@example.com",
			Groups:   []string{"test-group"},
		},
	}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Fatalf("incorrect client config:\n%s", diff)
	}
}

func TestImpersonatingConfigGetterInsecureClient(t *testing.T) {
	g := kube.NewImpersonatingConfigGetter(&rest.Config{}, true, kube.UserPrefixes{})
	ctx := auth.WithPrincipal(context.TODO(), &auth.UserPrincipal{ID: "user@example.com"})

	cfg := g.Config(ctx)

	want := &rest.Config{
		Impersonate: rest.ImpersonationConfig{
			UserName: "user@example.com",
		},
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Fatalf("incorrect client config:\n%s", diff)
	}
}

func TestImpersonatingConfigGetterNoPrincipalInContext(t *testing.T) {
	g := kube.NewImpersonatingConfigGetter(&rest.Config{}, true, kube.UserPrefixes{})

	cfg := g.Config(context.TODO())

	want := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Fatalf("incorrect client config:\n%s", diff)
	}
}

func TestConfigWithPrincipal(t *testing.T) {
	user := &auth.UserPrincipal{
		ID:     "user-id",
		Groups: []string{"group1", "group2"},
	}
	config := &rest.Config{
		Host: "https://example.com",
	}

	userPrefixes := kube.UserPrefixes{
		UsernamePrefix: "prefix-",
		GroupsPrefix:   "prefix-",
	}

	// First call.
	cfg := kube.ConfigWithPrincipal(user, config, userPrefixes)

	expectedUser := "prefix-user-id"
	if cfg.Impersonate.UserName != expectedUser {
		t.Fatalf("cfg username didn't match expected: %s", cfg.Impersonate.UserName)
	}

	expectedGroups := []string{"prefix-group1", "prefix-group2"}
	if diff := cmp.Diff(expectedGroups, cfg.Impersonate.Groups); diff != "" {
		t.Fatalf("cfg groups didn't match expected:\n%s", diff)
	}

	// Second call.
	cfg = kube.ConfigWithPrincipal(user, config, userPrefixes)
	if diff := cmp.Diff(expectedGroups, cfg.Impersonate.Groups); diff != "" {
		t.Fatalf("cfg groups didn't match expected:\n%s", diff)
	}
}
