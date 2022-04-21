package clustersmngr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	userNamespaceTTL = 1 * time.Hour
)

// ClientsFactory is a factory for creating clients for clusters
//counterfeiter:generate . ClientsFactory
type ClientsFactory interface {
	// GetUserClient returns the clusters client for the given user
	GetUserClient(ctx context.Context, user *auth.UserPrincipal) (Client, error)
	// UpdateClusters updates the clusters list
	UpdateClusters(ctx context.Context) error
	// UpdateNamespaces updates the namespaces all namespaces for all clusters
	UpdateNamespaces(ctx context.Context) error
	// Start starts go routines to keep clusters and namespaces lists up to date
	Start(ctx context.Context)
}

type clientsFactory struct {
	hubClient       client.Client
	clustersFetcher ClusterFetcher
	nsChecker       nsaccess.Checker
	log             logr.Logger

	// list of clusters returned by the clusters fetcher
	clusters *Clusters
	// the lists of all namespaces of each cluster
	clustersNamespaces *ClustersNamespaces
	// lists of namespaces accessible by the user on every cluster
	usersNamespaces *UsersNamespaces
}

func NewClientFactory(hubClient client.Client, fetcher ClusterFetcher, nsChecker nsaccess.Checker, logger logr.Logger) ClientsFactory {
	return &clientsFactory{
		hubClient:          hubClient,
		clustersFetcher:    fetcher,
		nsChecker:          nsChecker,
		clusters:           &Clusters{},
		clustersNamespaces: &ClustersNamespaces{},
		usersNamespaces:    &UsersNamespaces{Cache: ttlcache.New(24 * time.Hour)},
		log:                logger,
	}
}

func (cf *clientsFactory) Start(ctx context.Context) {
	go cf.watchClusters(ctx)
	go cf.watchNamespaces(ctx)
}

func (cf *clientsFactory) watchClusters(ctx context.Context) {
	if err := wait.PollImmediateInfinite(30*time.Second, func() (bool, error) {
		if err := cf.UpdateClusters(ctx); err != nil {
			return false, err
		}

		return true, nil
	}); err != nil {
		cf.log.Error(err, "failed pooling clusters")
	}
}

func (cf *clientsFactory) UpdateClusters(ctx context.Context) error {
	clusters, err := cf.clustersFetcher.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch clusters: %w", err)
	}

	cf.clusters.Set(clusters)

	return nil
}

func (cf *clientsFactory) watchNamespaces(ctx context.Context) {
	if err := wait.PollImmediateInfinite(30*time.Second, func() (bool, error) {
		if err := cf.UpdateNamespaces(ctx); err != nil {
			return false, err
		}

		return false, nil
	}); err != nil {
		cf.log.Error(err, "failed pooling namespaces")
	}
}

func (cf *clientsFactory) UpdateNamespaces(ctx context.Context) error {
	clients, err := clientsForClusters(cf.clusters.Get())
	if err != nil {
		cf.log.Error(err, "failed to create clients for", "clusters", cf.clusters.Get())
		return err
	}

	wg := sync.WaitGroup{}

	for clusterName, c := range clients {
		wg.Add(1)

		go func(clusterName string, c client.Client) {
			defer wg.Done()

			nsList := &v1.NamespaceList{}

			if err := c.List(ctx, nsList); err != nil {
				cf.log.Error(err, "failed listing namespaces", "cluster", clusterName)
			}

			cf.clustersNamespaces.Set(clusterName, nsList.Items)
		}(clusterName, c)
	}

	wg.Wait()

	return nil
}

func (cf *clientsFactory) GetUserClient(ctx context.Context, user *auth.UserPrincipal) (Client, error) {
	pool := NewClustersClientsPool()

	for _, cluster := range cf.clusters.Get() {
		if err := pool.Add(ClientConfigWithUser(user), cluster); err != nil {
			return nil, fmt.Errorf("failed adding cluster client to pool: %w", err)
		}
	}

	return NewClient(pool, cf.userNsList(ctx, user)), nil
}

func restConfigFromCluster(cluster Cluster) *rest.Config {
	return &rest.Config{
		Host:            cluster.Server,
		BearerToken:     cluster.BearerToken,
		TLSClientConfig: cluster.TLSConfig,
	}
}

func (cf *clientsFactory) userNsList(ctx context.Context, user *auth.UserPrincipal) map[string][]v1.Namespace {
	userNamespaces := cf.usersNamespaces.GetAll(user, cf.clusters.Get())
	if len(userNamespaces) > 0 {
		return userNamespaces
	}

	wg := sync.WaitGroup{}

	for _, cluster := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster Cluster) {
			defer wg.Done()

			clusterNs := cf.clustersNamespaces.Get(cluster.Name)

			filteredNs, err := cf.nsChecker.FilterAccessibleNamespaces(ctx, impersonatedConfig(cluster, user), clusterNs)
			if err != nil {
				cf.log.Error(err, "failed filtering namespaces", "cluster", cluster.Name, "user", user.ID)
			}

			cf.usersNamespaces.Set(user, cluster.Name, filteredNs)
		}(cluster)
	}

	wg.Wait()

	return cf.usersNamespaces.GetAll(user, cf.clusters.Get())
}

func impersonatedConfig(cluster Cluster, user *auth.UserPrincipal) *rest.Config {
	shallowCopy := *restConfigFromCluster(cluster)

	shallowCopy.Impersonate = rest.ImpersonationConfig{
		UserName: user.ID,
		Groups:   user.Groups,
	}

	return &shallowCopy
}

func clientsForClusters(clusters []Cluster) (map[string]client.Client, error) {
	clients := map[string]client.Client{}

	for _, cluster := range clusters {
		c, err := client.New(restConfigFromCluster(cluster), client.Options{})
		if err != nil {
			return nil, fmt.Errorf("failed creating client for cluster %s: %w", cluster.Name, err)
		}

		clients[cluster.Name] = c
	}

	return clients, nil
}

type Clusters struct {
	sync.RWMutex
	clusters []Cluster
}

func (c *Clusters) Set(clusters []Cluster) {
	c.Lock()
	defer c.Unlock()

	c.clusters = clusters
}

func (c *Clusters) Get() []Cluster {
	c.Lock()
	defer c.Unlock()

	return c.clusters
}

type ClustersNamespaces struct {
	sync.RWMutex
	Namespaces map[string][]v1.Namespace
}

func (cn *ClustersNamespaces) Set(cluster string, namespaces []v1.Namespace) {
	cn.Lock()
	defer cn.Unlock()

	if cn.Namespaces == nil {
		cn.Namespaces = make(map[string][]v1.Namespace)
	}

	cn.Namespaces[cluster] = namespaces
}

func (cn *ClustersNamespaces) Get(cluster string) []v1.Namespace {
	cn.Lock()
	defer cn.Unlock()

	return cn.Namespaces[cluster]
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

func (un *UsersNamespaces) GetAll(user *auth.UserPrincipal, clusters []Cluster) map[string][]v1.Namespace {
	namespaces := map[string][]v1.Namespace{}

	for _, cluster := range clusters {
		if nsList, found := un.Get(user, cluster.Name); found {
			namespaces[cluster.Name] = nsList
		}
	}

	return namespaces
}

func (u UsersNamespaces) cacheKey(user *auth.UserPrincipal, cluster string) uint64 {
	return ttlcache.StringKey(fmt.Sprintf("%s:%s", user.ID, cluster))
}
