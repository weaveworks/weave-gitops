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
	userNamespaceTTL        = 1 * time.Hour
	watchClustersFrequency  = 30 * time.Second
	watchNamespaceFrequency = 30 * time.Second
)

// ClientsFactory is a factory for creating clients for clusters
//counterfeiter:generate . ClientsFactory
type ClientsFactory interface {
	// GetImpersonatedClient returns the clusters client for the given user
	GetImpersonatedClient(ctx context.Context, user *auth.UserPrincipal) (Client, error)
	// UpdateClusters updates the clusters list
	UpdateClusters(ctx context.Context) error
	// UpdateNamespaces updates the namespaces all namespaces for all clusters
	UpdateNamespaces(ctx context.Context) error
	// GetServerClient returns the cluster client with gitops server permissions
	GetServerClient(ctx context.Context) (Client, error)
	// GetClustersNamespaces returns the namespaces for all clusters
	GetClustersNamespaces() map[string][]v1.Namespace
	// Start starts go routines to keep clusters and namespaces lists up to date
	Start(ctx context.Context)
}

type clientsFactory struct {
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

func NewClientFactory(fetcher ClusterFetcher, nsChecker nsaccess.Checker, logger logr.Logger) ClientsFactory {
	return &clientsFactory{
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
	if err := wait.PollImmediateInfinite(watchClustersFrequency, func() (bool, error) {
		if err := cf.UpdateClusters(ctx); err != nil {
			return false, err
		}

		return true, nil
	}); err != nil {
		cf.log.Error(err, "failed polling clusters")
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
	if err := wait.PollImmediateInfinite(watchNamespaceFrequency, func() (bool, error) {
		if err := cf.UpdateNamespaces(ctx); err != nil {
			return false, err
		}

		return false, nil
	}); err != nil {
		cf.log.Error(err, "failed polling namespaces")
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

func (cf *clientsFactory) GetImpersonatedClient(ctx context.Context, user *auth.UserPrincipal) (Client, error) {
	pool := NewClustersClientsPool()

	for _, cluster := range cf.clusters.Get() {
		if err := pool.Add(ClientConfigWithUser(user), cluster); err != nil {
			return nil, fmt.Errorf("failed adding cluster client to pool: %w", err)
		}
	}

	return NewClient(pool, cf.userNsList(ctx, user)), nil
}

func (cf *clientsFactory) GetServerClient(ctx context.Context) (Client, error) {
	pool := NewClustersClientsPool()

	for _, cluster := range cf.clusters.Get() {
		if err := pool.Add(restConfigFromCluster, cluster); err != nil {
			return nil, fmt.Errorf("failed adding cluster client to pool: %w", err)
		}
	}

	return NewClient(pool, cf.clustersNamespaces.namespaces), nil
}

func (cf *clientsFactory) GetClustersNamespaces() map[string][]v1.Namespace {
	return cf.clustersNamespaces.namespaces
}

func restConfigFromCluster(cluster Cluster) *rest.Config {
	return &rest.Config{
		Host:            cluster.Server,
		BearerToken:     cluster.BearerToken,
		TLSClientConfig: cluster.TLSConfig,
		QPS:             ClientQPS,
		Burst:           ClientBurst,
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
