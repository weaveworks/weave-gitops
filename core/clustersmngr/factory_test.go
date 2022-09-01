package clustersmngr_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
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

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)
	err = clientsFactory.UpdateClusters(ctx)
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

func TestGetImpersonatedDiscoveryClient(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()

	ns1 := createNamespace(g)

	nsChecker := &nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesReturns([]v1.Namespace{*ns1}, nil)

	clustersFetcher := fetcher.NewSingleClusterFetcher(k8sEnv.Rest)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	err = clientsFactory.UpdateNamespaces(ctx)
	g.Expect(err).To(BeNil())

	dc, err := clientsFactory.GetImpersonatedDiscoveryClient(ctx, &auth.UserPrincipal{ID: "user-id"}, clustersmngr.DefaultCluster)
	g.Expect(err).To(BeNil())

	_, err = dc.ServerVersion()
	g.Expect(err).To(BeNil())
}

func TestUpdateNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()
	nsChecker := &nsaccessfakes.FakeChecker{}
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)

	clusterName1 := "foo"
	clusterName2 := "bar"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	t.Run("UpdateNamespaces will return a map based on the clusters returned by Fetch", func(t *testing.T) {
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clientsFactory.UpdateNamespaces(ctx)).To(Succeed())

		contents := clientsFactory.GetClustersNamespaces()

		g.Expect(contents).To(HaveLen(2))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).To(HaveKey(clusterName2))
	})

	t.Run("When a cluster is no longer in the clusters cache, the clustersNamespaces cache updates to reflect this", func(t *testing.T) {
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clientsFactory.UpdateNamespaces(ctx)).To(Succeed())

		contents := clientsFactory.GetClustersNamespaces()

		g.Expect(contents).To(HaveLen(1))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).ToNot(HaveKey(clusterName2))
	})

	t.Run("UpdateNamespaces will return partial results if a single cluster fails to connect", func(t *testing.T) {
		clusterName3 := "foobar"
		c3 := makeUnreachableLeafCluster(t, clusterName3)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2, c3}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clientsFactory.UpdateNamespaces(ctx)).To(MatchError(MatchRegexp("failed adding cluster client to pool.*cluster: %s.*", clusterName3)))

		contents := clientsFactory.GetClustersNamespaces()

		g.Expect(contents).To(HaveLen(2))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).To(HaveKey(clusterName2))
	})
}

func TestUpdateUsers(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()
	nsChecker := &nsaccessfakes.FakeChecker{}
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)

	clusterName1 := "foo"
	clusterName2 := "bar"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	u1 := &auth.UserPrincipal{ID: "drstrange"}

	t.Run("UpdateUsers will return a map based on the clusters returned by Fetch", func(t *testing.T) {
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		clientsFactory.UpdateUserNamespaces(ctx, u1)

		contents := clientsFactory.GetUserNamespaces(u1)

		g.Expect(contents).To(HaveLen(2))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).To(HaveKey(clusterName2))
	})

	t.Run("GetUsersNamespaces will only return cached items matched to the current clusters list", func(t *testing.T) {
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())

		contents := clientsFactory.GetUserNamespaces(u1)

		g.Expect(contents).To(HaveLen(1))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).NotTo(HaveKey(clusterName2))
	})
}

func TestUpdateUsersFailsToConnect(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()
	nsChecker := nsaccess.NewChecker(nil)
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)

	clusterName1 := "foo"

	c1 := makeLeafCluster(t, clusterName1)

	u1 := &auth.UserPrincipal{ID: "drstrange"}

	t.Run("UpdateUserNamespaces remains unchanged if a connection failure occurs", func(t *testing.T) {
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clientsFactory.UpdateNamespaces(ctx)).To(Succeed())
		g.Expect(clientsFactory.GetClustersNamespaces()).To(HaveLen(1))

		c1 = makeUnreachableLeafCluster(t, clusterName1)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1}, nil)
		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		clientsFactory.UpdateUserNamespaces(ctx, u1)

		// Get clusters namespace hasn't been changed.
		g.Expect(clientsFactory.GetClustersNamespaces()).To(HaveLen(1))
	})
}
