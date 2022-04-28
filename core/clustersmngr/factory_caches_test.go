package clustersmngr_test

import (
	"testing"
	"time"

	"github.com/cheshir/ttlcache"
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
