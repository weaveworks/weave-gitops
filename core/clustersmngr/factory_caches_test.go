package clustersmngr_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

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
		g.Expect(nss).To(Equal([]v1.Namespace{ns}))
	})

	t.Run("all namespaces from all", func(t *testing.T) {
		cl, err := cluster.NewSingleCluster(clusterName, &rest.Config{}, nil, kube.UserPrefixes{})
		g.Expect(err).NotTo(HaveOccurred())
		nsMap := un.GetAll(user, []cluster.Cluster{cl})
		g.Expect(nsMap).To(Equal(map[string][]v1.Namespace{clusterName: {ns}}))
	})
}

func TestClusters(t *testing.T) {
	g := NewGomegaWithT(t)

	cs := clustersmngr.Clusters{}

	c1 := "cluster-1"
	c2 := "cluster-2"

	cluster1, err := cluster.NewSingleCluster(c1, &rest.Config{}, nil, kube.UserPrefixes{})
	g.Expect(err).NotTo(HaveOccurred())

	cluster2, err := cluster.NewSingleCluster(c2, &rest.Config{}, nil, kube.UserPrefixes{})
	g.Expect(err).NotTo(HaveOccurred())

	testClusters := []cluster.Cluster{cluster1, cluster2}

	// simulating concurrent access
	go cs.Set(testClusters)
	go cs.Set(testClusters)

	cs.Set(testClusters)

	g.Expect(cs.Get()).To(Equal([]cluster.Cluster{cluster1, cluster2}))

	g.Expect(cs.Hash()).To(Equal(fmt.Sprintf("%s%s", c1, c2)))
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

	cs.Clear()

	g.Expect(cs.Get(clusterName)).To(HaveLen(0))
}

var ClusterComparer = cmp.Comparer(func(a, b cluster.Cluster) bool {
	return a.GetName() == b.GetName() && a.GetHost() == b.GetHost()
})

func TestClusterSet_Set(t *testing.T) {
	cs := clustersmngr.Clusters{}
	cluster1 := newTestCluster(t, "cluster1", "server1")
	cluster2 := newTestCluster(t, "cluster2", "server2")
	cluster3 := newTestCluster(t, "cluster2", "server3")

	testClusters := []cluster.Cluster{cluster1, cluster2, cluster3}

	added, removed := cs.Set(testClusters)
	if diff := cmp.Diff([]cluster.Cluster{cluster1, cluster2, cluster3}, added, ClusterComparer); diff != "" {
		t.Fatalf("failed to calculate added:\n%s", diff)
	}

	if diff := cmp.Diff([]cluster.Cluster{}, removed, ClusterComparer); diff != "" {
		t.Fatalf("failed to calculate removed:\n%s", diff)
	}

	testClusters = []cluster.Cluster{cluster1}

	added, removed = cs.Set(testClusters)
	if diff := cmp.Diff([]cluster.Cluster{}, added, ClusterComparer); diff != "" {
		t.Fatalf("failed to calculate added:\n%s", diff)
	}

	if diff := cmp.Diff([]cluster.Cluster{cluster2, cluster3}, removed, ClusterComparer); diff != "" {
		t.Fatalf("failed to calculate removed:\n%s", diff)
	}
}

func newTestCluster(t *testing.T, name, server string) cluster.Cluster {
	t.Helper()
	c, err := cluster.NewSingleCluster(name, &rest.Config{Host: server}, nil, kube.UserPrefixes{})
	if err != nil {
		t.Error("Expected error to be nil, got", err)
	}

	return c
}
