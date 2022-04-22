package clustersmngr_test

import (
	"testing"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/go-logr/logr"
	"github.com/onsi/gomega"
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

	clustersFetcher, err := fetcher.NewSingleClusterFetcher(k8sEnv.Rest)
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(k8sEnv.Client, clustersFetcher, nsChecker, logger)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	err = clientsFactory.UpdateNamespaces(ctx)
	g.Expect(err).To(BeNil())

	_, err = clientsFactory.GetImpersonatedClient(ctx, &auth.UserPrincipal{ID: "user-id"})
	g.Expect(err).To(BeNil())

	t.Run("checks all namespaces in the cluster when through the filtering", func(t *testing.T) {
		g.Expect(nsChecker.FilterAccessibleNamespacesCallCount()).To(gomega.Equal(1))

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

func TestUsersNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	un := clustersmngr.UsersNamespaces{Cache: ttlcache.New(1 * time.Second)}

	user := &auth.UserPrincipal{ID: "user-id"}

	ns := v1.Namespace{}
	ns.Name = "ns1"

	clusterName := "cluster-1"

	un.Set(user, clusterName, []v1.Namespace{ns})

	t.Run("namespaces of a single cluster", func(t *testing.T) {
		nss, found := un.Get(user, clusterName)
		g.Expect(found).To(BeTrue())
		g.Expect(nss).To(gomega.Equal([]v1.Namespace{ns}))
	})

	t.Run("all namespaces from all", func(t *testing.T) {
		nsMap := un.GetAll(user, []clustersmngr.Cluster{{Name: clusterName}})
		g.Expect(nsMap).To(gomega.Equal(map[string][]v1.Namespace{clusterName: {ns}}))
	})
}

func TestClusters(t *testing.T) {
	g := NewGomegaWithT(t)

	cs := clustersmngr.Clusters{}

	clusterName := "cluster-1"
	clusters := []clustersmngr.Cluster{{Name: clusterName}}

	// simulating concurrent access
	go cs.Set(clusters)
	go cs.Set(clusters)

	cs.Set(clusters)

	g.Expect(cs.Get()).To(Equal([]clustersmngr.Cluster{{Name: clusterName}}))
}

func TestClustersNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	cs := clustersmngr.ClustersNamespaces{}

	clusterName := "cluster-1"

	ns := v1.Namespace{}
	ns.Name = "ns1"

	// simulating concurrent access
	go cs.Set(clusterName, []v1.Namespace{ns})
	go cs.Set(clusterName, []v1.Namespace{ns})

	cs.Set(clusterName, []v1.Namespace{ns})

	g.Expect(cs.Get(clusterName)).To(Equal([]v1.Namespace{ns}))
}
