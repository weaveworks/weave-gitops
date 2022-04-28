package clustersmngr_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
)

func TestGetImpersonatedClient(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()

	ns1 := createNamespace(g)
	ns2 := createNamespace(g)

	nsChecker := &nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesReturns([]v1.Namespace{*ns2}, nil)

	clustersFetcher := fetcher.NewSingleClusterFetcher(k8sEnv.Rest)

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger)
	err := clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	err = clientsFactory.UpdateNamespaces(ctx)
	g.Expect(err).To(BeNil())

	_, err = clientsFactory.GetImpersonatedClient(ctx, &auth.UserPrincipal{ID: "user-id"})
	g.Expect(err).To(BeNil())

	t.Run("checks all namespaces in the cluster when through the filtering", func(t *testing.T) {
		g.Expect(nsChecker.FilterAccessibleNamespacesCallCount()).To(Equal(1))

		_, _, nss := nsChecker.FilterAccessibleNamespacesArgsForCall(0)
		nsFound := 0
		for _, n := range nss {
			if n.Name == ns1.Name || n.Name == ns2.Name {
				nsFound++
			}
		}

		g.Expect(nsFound).To(Equal(2))
	})
}
