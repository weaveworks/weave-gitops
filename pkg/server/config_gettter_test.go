package server_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/rest"
)

var _ server.ConfigGetter = (*server.ImpersonatingConfigGetter)(nil)

func TestImpersonatingConfigGetterPrincipalInContext(t *testing.T) {
	g := server.NewImpersonatingConfigGetter(&rest.Config{}, false)
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
	g := server.NewImpersonatingConfigGetter(&rest.Config{}, false)
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
	g := server.NewImpersonatingConfigGetter(&rest.Config{}, true)
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
	g := server.NewImpersonatingConfigGetter(&rest.Config{}, true)

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
