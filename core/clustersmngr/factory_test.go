package clustersmngr_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestGetImpersonatedClient(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ns1 := createNamespace(g)
	ns2 := createNamespace(g)

	nsChecker := &nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesReturns([]v1.Namespace{*ns2}, nil)

	cluster, err := cluster.NewSingleCluster("test", k8sEnv.Rest, nil, cluster.DefaultKubeConfigOptions...)
	g.Expect(err).To(BeNil())

	clustersFetcher := fetcher.NewSingleClusterFetcher(cluster)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)
	err = clustersManager.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	err = clustersManager.UpdateNamespaces(ctx)
	g.Expect(err).To(BeNil())

	_, err = clustersManager.GetImpersonatedClient(ctx, &auth.UserPrincipal{ID: "user-id"})
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

func TestUseUserClientForNamespaces(t *testing.T) {
	// enabled the feature flag for this test
	useUserClientForNamespaces := featureflags.Get("WEAVE_GITOPS_FEATURE_USE_USER_CLIENT_FOR_NAMESPACES")

	featureflags.Set("WEAVE_GITOPS_FEATURE_USE_USER_CLIENT_FOR_NAMESPACES", "true")
	defer featureflags.Set("WEAVE_GITOPS_FEATURE_USE_USER_CLIENT_FOR_NAMESPACES", useUserClientForNamespaces)

	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ns1 := createNamespace(g)
	ns2 := createNamespace(g)
	// User has be able to list namespaces
	createClusterRoleBinding(g, "user-id")

	nsChecker := &nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesReturns([]v1.Namespace{*ns2}, nil)

	cluster, err := cluster.NewSingleCluster("test", k8sEnv.Rest, nil, cluster.DefaultKubeConfigOptions...)
	g.Expect(err).To(BeNil())

	clustersFetcher := fetcher.NewSingleClusterFetcher(cluster)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)
	err = clustersManager.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	// We would usually call UpdateNamespaces here, but we are testing the
	// user client for namespaces, so we don't need to call it.
	// E.g. can skip some call like this:
	// err = clustersManager.UpdateNamespaces(ctx)

	userClient, err := clustersManager.GetImpersonatedClient(ctx, &auth.UserPrincipal{ID: "user-id"})
	g.Expect(err).To(BeNil())

	t.Run("checks all namespaces in the cluster when through the filtering", func(t *testing.T) {
		g = NewGomegaWithT(t)
		g.Expect(nsChecker.FilterAccessibleNamespacesCallCount()).To(Equal(1))

		g.Expect(userClient.Namespaces()).To(HaveLen(1))
		g.Expect(userClient.Namespaces()["test"]).To(HaveLen(1))
		g.Expect(userClient.Namespaces()["test"][0].GetName()).To(Equal(ns2.Name))

		a, b, nss := nsChecker.FilterAccessibleNamespacesArgsForCall(0)
		fmt.Println(a, b, nss)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ns1 := createNamespace(g)

	nsChecker := &nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesReturns([]v1.Namespace{*ns1}, nil)

	cl, err := cluster.NewSingleCluster(cluster.DefaultCluster, k8sEnv.Rest, nil, cluster.DefaultKubeConfigOptions...)
	g.Expect(err).To(BeNil())

	clustersFetcher := fetcher.NewSingleClusterFetcher(cl)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)
	err = clustersManager.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	err = clustersManager.UpdateNamespaces(ctx)
	g.Expect(err).To(BeNil())

	dc, err := clustersManager.GetImpersonatedDiscoveryClient(ctx, &auth.UserPrincipal{ID: "user-id"}, cluster.DefaultCluster)
	g.Expect(err).To(BeNil())

	_, err = dc.ServerVersion()
	g.Expect(err).To(BeNil())
}

func TestUpdateNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := &nsaccessfakes.FakeChecker{}
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)

	clusterName1 := "foo"
	clusterName2 := "bar"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	t.Run("UpdateNamespaces will return a map based on the clusters returned by Fetch", func(t *testing.T) {
		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.UpdateNamespaces(ctx)).To(Succeed())

		contents := clustersManager.GetClustersNamespaces()

		g.Expect(contents).To(HaveLen(2))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).To(HaveKey(clusterName2))
	})

	t.Run("When a cluster is no longer in the clusters cache, the clustersNamespaces cache updates to reflect this", func(t *testing.T) {
		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.UpdateNamespaces(ctx)).To(Succeed())

		contents := clustersManager.GetClustersNamespaces()

		g.Expect(contents).To(HaveLen(1))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).ToNot(HaveKey(clusterName2))
	})

	t.Run("UpdateNamespaces will return partial results if a single cluster fails to connect", func(t *testing.T) {
		clusterName3 := "foobar"
		c3 := makeUnreachableLeafCluster(t, clusterName3)
		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2, c3}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.UpdateNamespaces(ctx)).To(MatchError(MatchRegexp("failed creating server client to pool.*cluster: %s.*", clusterName3)))
		contents := clustersManager.GetClustersNamespaces()

		g.Expect(contents).To(HaveLen(2))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).To(HaveKey(clusterName2))
	})
}

func TestUpdateUserNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := &nsaccessfakes.FakeChecker{}
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)

	clusterName1 := "foo"
	clusterName2 := "bar"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	u1 := &auth.UserPrincipal{ID: "drstrange"}

	t.Run("UpdateUserNamespaces will return a map based on the clusters returned by Fetch", func(t *testing.T) {
		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		clustersManager.UpdateUserNamespaces(ctx, u1)

		contents := clustersManager.GetUserNamespaces(u1)

		g.Expect(contents).To(HaveLen(2))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).To(HaveKey(clusterName2))
	})

	t.Run("GetUsersNamespaces will only return cached items matched to the current clusters list", func(t *testing.T) {
		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		contents := clustersManager.GetUserNamespaces(u1)

		g.Expect(contents).To(HaveLen(1))
		g.Expect(contents).To(HaveKey(clusterName1))
		g.Expect(contents).NotTo(HaveKey(clusterName2))
	})
}

func TestUpdateUserNamespacesFailsToConnect(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := nsaccess.NewChecker(nil)
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)

	clusterName1 := "foo"

	c1 := makeLeafCluster(t, clusterName1)

	u1 := &auth.UserPrincipal{ID: "drstrange"}

	t.Run("UpdateUserNamespaces remains unchanged if a connection failure occurs", func(t *testing.T) {
		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.UpdateNamespaces(ctx)).To(Succeed())
		g.Expect(clustersManager.GetClustersNamespaces()).To(HaveLen(1))

		c1 = makeUnreachableLeafCluster(t, clusterName1)
		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		clustersManager.UpdateUserNamespaces(ctx, u1)

		// Get clusters namespace hasn't been changed.
		g.Expect(clustersManager.GetClustersNamespaces()).To(HaveLen(1))
	})
}

func TestGetClusters(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := nsaccess.NewChecker(nil)
	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)

	c1 := makeLeafCluster(t, "foo")
	c2 := makeLeafCluster(t, "foo")

	t.Run("GetClusters returns clusters that were fetched", func(t *testing.T) {
		g.Expect(clustersManager.GetClusters()).To(BeEmpty())

		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.GetClusters()).To(Equal([]cluster.Cluster{c1}))

		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.GetClusters()).To(Equal([]cluster.Cluster{c1, c2}))

		clustersFetcher.FetchReturns([]cluster.Cluster{c2}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.GetClusters()).To(Equal([]cluster.Cluster{c2}))

		clustersFetcher.FetchReturns([]cluster.Cluster{}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())
		g.Expect(clustersManager.GetClusters()).To(BeEmpty())
	})
}

func TestUpdateClusters(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsChecker := &nsaccessfakes.FakeChecker{}

	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)
	err := clustersManager.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	clusterName1 := "bar"
	clusterName2 := "foo"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	watcher := clustersManager.Subscribe()
	g.Expect(watcher).ToNot(BeNil())

	clusterNames := func(c []cluster.Cluster) []string {
		names := []string{}
		for _, v := range c {
			names = append(names, v.GetName())
		}

		return names
	}

	t.Run("watcher should be notified with two clusters added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		updates := <-watcher.Updates

		g.Expect(clusterNames(updates.Added)).To(Equal(clusterNames([]cluster.Cluster{c1, c2})))
		g.Expect(clusterNames(updates.Removed)).To(BeEmpty())
	})

	t.Run("watcher should be notified with one cluster removed", func(t *testing.T) {
		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		updates := <-watcher.Updates

		g.Expect(clusterNames(updates.Added)).To(BeEmpty())
		g.Expect(clusterNames(updates.Removed)).To(Equal(clusterNames([]cluster.Cluster{c2})))
	})

	t.Run("watcher shouldn't be notified when there are no updates", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// Call 1 with no updates
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		// Call 2 with updates
		clustersFetcher.FetchReturns([]cluster.Cluster{c2}, nil)
		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		updates := <-watcher.Updates

		// Assert watcher received a notification from the second UpdateClusters call
		g.Expect(clusterNames(updates.Added)).To(Equal(clusterNames([]cluster.Cluster{c2})))
		g.Expect(clusterNames(updates.Removed)).To(Equal(clusterNames([]cluster.Cluster{c1})))
	})

	t.Run("Updates channel should be closed when calling Unsubscribe", func(t *testing.T) {
		g := NewGomegaWithT(t)
		watcher.Unsubscribe()

		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		updates, ok := <-watcher.Updates

		g.Expect(ok).To(BeFalse())
		g.Expect(updates.Added).To(BeNil())
	})

	t.Run("Unsubscribe should close the correct channel", func(t *testing.T) {
		g := NewGomegaWithT(t)

		watcher1 := clustersManager.Subscribe()
		g.Expect(watcher1).ToNot(BeNil())
		watcher2 := clustersManager.Subscribe()
		g.Expect(watcher2).ToNot(BeNil())

		clustersFetcher.FetchReturns([]cluster.Cluster{c1, c2}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		watcher2.Unsubscribe()

		clustersFetcher.FetchReturns([]cluster.Cluster{c1}, nil)

		g.Expect(clustersManager.UpdateClusters(ctx)).To(Succeed())

		updates1, ok1 := <-watcher1.Updates
		g.Expect(ok1).To(BeTrue())
		g.Expect(clusterNames(updates1.Added)).To(BeEmpty())
		g.Expect(clusterNames(updates1.Removed)).To(Equal(clusterNames([]cluster.Cluster{c2})))

		updates2, ok2 := <-watcher2.Updates
		g.Expect(ok2).To(BeFalse())
		g.Expect(updates2.Added).To(BeNil())
	})
}

func TestClientCaching(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ns1 := createNamespace(g)

	nsChecker := &nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesReturns([]v1.Namespace{*ns1}, nil)

	cluster := new(clusterfakes.FakeCluster)
	cluster.GetNameReturns("Default")
	cluster.GetServerClientReturns(k8sEnv.Client, nil)
	cluster.GetUserClientReturns(k8sEnv.Client, nil)
	cs, err := kubernetes.NewForConfig(k8sEnv.Rest)
	g.Expect(err).To(BeNil())
	cluster.GetUserClientsetReturns(cs, nil)
	cluster.GetServerClientsetReturns(cs, nil)

	clustersFetcher := fetcher.NewSingleClusterFetcher(cluster)

	userID := "user-id"

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{clustersFetcher}, nsChecker, logger)

	err = clustersManager.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	err = clustersManager.UpdateNamespaces(ctx)
	g.Expect(err).To(BeNil())

	_, err = clustersManager.GetImpersonatedClient(ctx, &auth.UserPrincipal{ID: userID})
	g.Expect(err).To(BeNil())

	_, err = clustersManager.GetImpersonatedClient(ctx, &auth.UserPrincipal{ID: userID})
	g.Expect(err).To(BeNil())

	g.Expect(cluster.GetUserClientCallCount()).To(Equal(1))
	g.Expect(cluster.GetUserClientArgsForCall(0).ID).To(Equal(userID))
}

func TestClusterNamespacesLogValue(t *testing.T) {
	g := NewGomegaWithT(t)

	nss := func(names ...string) []v1.Namespace {
		var namespaces []v1.Namespace
		for _, name := range names {
			namespaces = append(namespaces, v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			})
		}
		return namespaces
	}

	clusterNamespaces := map[string][]v1.Namespace{
		"cluster1": nss("ns1", "ns2"),
		"cluster2": nss("ns3", "ns4"),
	}

	expected := map[string]clustersmngr.NamespacesSlice{
		"cluster1": {
			Namespaces: []string{"ns1", "ns2"},
			TotalCount: 2,
		},
		"cluster2": {
			Namespaces: []string{"ns3", "ns4"},
			TotalCount: 2,
		},
	}

	result := clustersmngr.ClusterNamespacesLogValue(clusterNamespaces)

	g.Expect(result).To(Equal(expected))

}
