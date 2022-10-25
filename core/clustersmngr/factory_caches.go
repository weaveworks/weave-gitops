package clustersmngr

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/cheshir/ttlcache"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ClusterCollection
type ClusterCollection interface {
	Update(ctx context.Context) (added, removed []cluster.Cluster, error error)
	Get(name string) (cluster.Cluster, error)
	GetAll() []cluster.Cluster
	Hash() string
}

type staticClusters struct {
	clusters []cluster.Cluster
}

// This returns a ClusterCollection that cannot update its cluster list
func NewStaticClusterCollection(clusters ...cluster.Cluster) ClusterCollection {
	return &staticClusters{clusters}
}

func (c *staticClusters) Update(ctx context.Context) (added, removed []cluster.Cluster, error error) {
	return []cluster.Cluster{}, []cluster.Cluster{}, nil
}

func (c *staticClusters) Get(name string) (cluster.Cluster, error) {
	for _, cluster := range c.clusters {
		if name == cluster.GetName() {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("couldn't find cluster %s", name)
}

func (c *staticClusters) GetAll() []cluster.Cluster {
	return c.clusters
}

func (c *staticClusters) Hash() string {
	names := []string{}

	for _, cluster := range c.clusters {
		names = append(names, cluster.GetName())
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
func (un *UsersNamespaces) GetAll(user *auth.UserPrincipal, clusters []cluster.Cluster) map[string][]v1.Namespace {
	namespaces := map[string][]v1.Namespace{}

	for _, cluster := range clusters {
		if nsList, found := un.Get(user, cluster.GetName()); found {
			namespaces[cluster.GetName()] = nsList
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
