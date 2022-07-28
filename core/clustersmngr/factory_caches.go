package clustersmngr

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/cheshir/ttlcache"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Clusters is a list of clusters
type Clusters struct {
	sync.RWMutex
	clusters    []Cluster
	clustersMap map[string]Cluster
}

func (c *Clusters) Set(clusters []Cluster) (added, removed []Cluster) {
	c.Lock()
	defer c.Unlock()

	currentClustersSet := sets.NewString()

	for _, cluster := range c.clusters {
		clusterKey := fmt.Sprintf("%s:%s", cluster.Name, cluster.Server)
		currentClustersSet.Insert(clusterKey)
	}

	newClustersSet := sets.NewString()
	clustersMap := map[string]Cluster{}

	for _, cluster := range clusters {
		clusterKey := fmt.Sprintf("%s:%s", cluster.Name, cluster.Server)
		newClustersSet.Insert(clusterKey)

		clustersMap[clusterKey] = cluster
	}

	addedClusters := newClustersSet.Difference(currentClustersSet)
	added = appendClusters(clustersMap, addedClusters.List())

	removedClusters := currentClustersSet.Difference(newClustersSet)
	removed = appendClusters(c.clustersMap, removedClusters.List())

	c.clusters = clusters
	c.clustersMap = clustersMap

	return added, removed
}

func appendClusters(clustersMap map[string]Cluster, keys []string) []Cluster {
	clusters := []Cluster{}

	for _, key := range keys {
		clusters = append(clusters, clustersMap[key])
	}

	return clusters
}

// Get will return the cached cluster list.
func (c *Clusters) Get() []Cluster {
	c.Lock()
	defer c.Unlock()

	return c.clusters
}

// Hash will return a hash of the list of clusters by name.
func (c *Clusters) Hash() string {
	names := []string{}

	for _, cluster := range c.clusters {
		names = append(names, cluster.Name)
	}

	sort.Strings(names)

	return strings.Join(names, "")
}

// ClusterNamespaces is a list of namespaces indexed by cluster name.
type ClustersNamespaces struct {
	sync.RWMutex
	namespaces map[string][]v1.Namespace
}

// Set will set the given namespace list in the cache.
// It takes the cluster name as a parameter, so that the list can be indexed by cluster name.
func (cn *ClustersNamespaces) Set(cluster string, namespaces []v1.Namespace) {
	cn.Lock()
	defer cn.Unlock()

	if cn.namespaces == nil {
		cn.namespaces = make(map[string][]v1.Namespace)
	}

	cn.namespaces[cluster] = namespaces
}

// AddNamespace will add the given namespace to the cache.
// It takes the cluster name and the namespace as a parameter.
func (cn *ClustersNamespaces) AddNamespace(cluster string, namespace v1.Namespace) {
	cn.Lock()
	defer cn.Unlock()

	if cn.namespaces == nil {
		cn.namespaces = make(map[string][]v1.Namespace)
	}

	namespaces, ok := cn.namespaces[cluster]
	if !ok {
		namespaces = []v1.Namespace{}
	}

	// Search for the namespace
	index := slices.IndexFunc(namespaces, func(ns v1.Namespace) bool {
		return ns.Name == namespace.Name
	})

	// if the namespace is already in the list, do nothing
	if index != -1 {
		return
	}

	namespaces = append(namespaces, namespace)
	cn.namespaces[cluster] = namespaces
}

// RemoveNamespace will remove the given namespace from the cache.
// It takes the cluster name and the namespace as a parameter.
func (cn *ClustersNamespaces) RemoveNamespace(cluster string, namespace v1.Namespace) {
	cn.Lock()
	defer cn.Unlock()

	if cn.namespaces == nil {
		return
	}

	namespaces, ok := cn.namespaces[cluster]
	if !ok {
		return
	}

	// Search for the namespace
	index := slices.IndexFunc(namespaces, func(ns v1.Namespace) bool {
		return ns.Name == namespace.Name
	})

	if index == -1 {
		return
	}

	namespaces = append(namespaces[:index], namespaces[index+1:]...)

	cn.namespaces[cluster] = namespaces
}

// UpdateNamespace will update the namespace in the cache, and create a new one if it doesn't exist.
// It takes the cluster name and the namespace as a parameter.
func (cn *ClustersNamespaces) UpdateNamespace(cluster string, namespace v1.Namespace) {
	cn.Lock()
	defer cn.Unlock()

	if cn.namespaces == nil {
		return
	}

	namespaces, ok := cn.namespaces[cluster]
	if !ok {
		return
	}

	// Search for the namespace
	index := slices.IndexFunc(namespaces, func(ns v1.Namespace) bool {
		return ns.Name == namespace.Name
	})

	if index == -1 {
		cn.AddNamespace(cluster, namespace)
		return
	}

	namespaces[index] = namespace

	cn.namespaces[cluster] = namespaces
}

// Clear will clear the cache.
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

func (un UsersNamespaces) cacheKey(user *auth.UserPrincipal, cluster string) uint64 {
	return ttlcache.StringKey(fmt.Sprintf("%s:%s", user.ID, cluster))
}
