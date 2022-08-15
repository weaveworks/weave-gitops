package clustersmngr

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/cheshir/ttlcache"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Clusters struct {
	sync.RWMutex
	clusters []Cluster
	names    map[string]Cluster
}

// Set updates Clusters.clusters, and returns the newly added clusters and removed clusters.
func (c *Clusters) Set(clusters []Cluster) (added, removed []Cluster) {
	c.Lock()
	defer c.Unlock()

	currentClustersSet := sets.NewString()
	for _, cluster := range c.clusters {
		currentClustersSet.Insert(cluster.Name)
	}

	newClustersSet := sets.NewString()
	clusterNames := map[string]Cluster{}

	for _, cluster := range clusters {
		newClustersSet.Insert(cluster.Name)
		clusterNames[cluster.Name] = cluster
	}

	addedNames := newClustersSet.Difference(currentClustersSet)
	added = appendClusters(clusterNames, addedNames.List())

	removedNames := currentClustersSet.Difference(newClustersSet)
	removed = appendClusters(c.names, removedNames.List())

	c.clusters = clusters
	c.names = clusterNames

	return added, removed
}

func appendClusters(clustersMap map[string]Cluster, names []string) []Cluster {
	clusters := []Cluster{}

	for _, name := range names {
		clusters = append(clusters, clustersMap[name])
	}

	return clusters
}

func (c *Clusters) Get() []Cluster {
	c.Lock()
	defer c.Unlock()

	return c.clusters
}

func (c *Clusters) Hash() string {
	names := []string{}

	for _, cluster := range c.clusters {
		names = append(names, cluster.Name)
	}

	sort.Strings(names)

	return strings.Join(names, "")
}

type ClustersNamespaces struct {
	sync.RWMutex
	namespaces map[string][]v1.Namespace
}

func (cn *ClustersNamespaces) Set(cluster string, namespaces []v1.Namespace) {
	cn.Lock()
	defer cn.Unlock()

	if cn.namespaces == nil {
		cn.namespaces = make(map[string][]v1.Namespace)
	}

	cn.namespaces[cluster] = namespaces
}

func (cn *ClustersNamespaces) Clear() {
	cn.Lock()
	defer cn.Unlock()

	cn.namespaces = make(map[string][]v1.Namespace)
}

func (cn *ClustersNamespaces) Get(cluster string) []v1.Namespace {
	cn.Lock()
	defer cn.Unlock()

	return cn.namespaces[cluster]
}

type UsersNamespaces struct {
	Cache *ttlcache.Cache
}

func (un *UsersNamespaces) Get(user *auth.UserPrincipal, cluster string) ([]v1.Namespace, bool) {
	if val, found := un.Cache.Get(un.cacheKey(user, cluster)); found {
		return val.([]v1.Namespace), true
	}

	return []v1.Namespace{}, false
}

func (un *UsersNamespaces) Set(user *auth.UserPrincipal, cluster string, nsList []v1.Namespace) {
	un.Cache.Set(un.cacheKey(user, cluster), nsList, userNamespaceTTL)
}

// GetAll will return all namespace mappings based on the list of clusters provided.
// The cache very well may contain more, but this List is targeted.
func (un *UsersNamespaces) GetAll(user *auth.UserPrincipal, clusters []Cluster) map[string][]v1.Namespace {
	namespaces := map[string][]v1.Namespace{}

	for _, cluster := range clusters {
		if nsList, found := un.Get(user, cluster.Name); found {
			namespaces[cluster.Name] = nsList
		}
	}

	return namespaces
}

func (un *UsersNamespaces) Clear() {
	un.Cache.Clear()
}

func (u UsersNamespaces) cacheKey(user *auth.UserPrincipal, cluster string) uint64 {
	return ttlcache.StringKey(fmt.Sprintf("%s:%s", user.ID, cluster))
}
