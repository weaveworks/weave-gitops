package clustersmngr_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
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
		nsMap := un.GetAll(user, []clustersmngr.Cluster{{Name: clusterName}})
		g.Expect(nsMap).To(Equal(map[string][]v1.Namespace{clusterName: {ns}}))
	})
}

func TestClusters(t *testing.T) {
	g := NewGomegaWithT(t)

	cs := clustersmngr.Clusters{}

	c1 := "cluster-1"
	c2 := "cluster-2"
	clusters := []clustersmngr.Cluster{{Name: c1}, {Name: c2}}

	// simulating concurrent access
	go cs.Set(clusters)
	go cs.Set(clusters)

	cs.Set(clusters)

	g.Expect(cs.Get()).To(Equal([]clustersmngr.Cluster{{Name: c1}, {Name: c2}}))

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

	got, found := cs.Get(clusterName)
	g.Expect(found).To(BeTrue())
	g.Expect(got).To(Equal([]v1.Namespace{ns}))

	cs.Clear()

	_, found = cs.Get(clusterName)
	g.Expect(found).To(BeFalse())
}

func TestClusterSet_Set(t *testing.T) {
	cs := clustersmngr.Clusters{}
	cluster1 := newTestCluster("cluster1", "server1")
	cluster2 := newTestCluster("cluster2", "server2")
	cluster3 := newTestCluster("cluster2", "server3")

	clusters := []clustersmngr.Cluster{cluster1, cluster2, cluster3}

	added, removed := cs.Set(clusters)
	if diff := cmp.Diff([]clustersmngr.Cluster{cluster1, cluster2, cluster3}, added); diff != "" {
		t.Fatalf("failed to calculate added:\n%s", diff)
	}

	if diff := cmp.Diff([]clustersmngr.Cluster{}, removed); diff != "" {
		t.Fatalf("failed to calculate removed:\n%s", diff)
	}

	clusters = []clustersmngr.Cluster{cluster1}

	added, removed = cs.Set(clusters)
	if diff := cmp.Diff([]clustersmngr.Cluster{}, added); diff != "" {
		t.Fatalf("failed to calculate added:\n%s", diff)
	}

	if diff := cmp.Diff([]clustersmngr.Cluster{cluster2, cluster3}, removed); diff != "" {
		t.Fatalf("failed to calculate removed:\n%s", diff)
	}
}

func newTestCluster(name, server string) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name:   name,
		Server: server,
	}
}
