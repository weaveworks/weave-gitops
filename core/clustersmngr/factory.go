package clustersmngr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientsFactory struct {
	hubClient       client.Client
	clustersFetcher ClusterFetcher

	clustersMutex sync.Mutex
	clusters      []Cluster

	clustersNsMutex    sync.RWMutex
	clustersNamespaces map[string][]v1.Namespace

	userNssMutex    sync.RWMutex
	usersNamespaces map[string]map[string][]v1.Namespace
	nsChecker       nsaccess.Checker

	logger logr.Logger
}

func NewClientFactory(hubClient client.Client, fetcher ClusterFetcher, nsChecker nsaccess.Checker, logger logr.Logger) *ClientsFactory {
	return &ClientsFactory{
		hubClient:          hubClient,
		clustersFetcher:    fetcher,
		nsChecker:          nsChecker,
		clusters:           []Cluster{},
		clustersNamespaces: map[string][]v1.Namespace{},
		usersNamespaces:    map[string]map[string][]v1.Namespace{},
		logger:             logger,
	}
}

func (cf *ClientsFactory) Start(ctx context.Context) error {
	go cf.watchClusters(ctx)
	go cf.watchNamespaces(ctx)

	return nil
}

func (cf *ClientsFactory) watchClusters(ctx context.Context) {
	if err := wait.PollImmediateInfinite(30*time.Second, func() (bool, error) {
		clusters, err := cf.clustersFetcher.Fetch(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to fetch clusters: %w", err)
		}

		cf.clustersMutex.Lock()
		cf.clusters = clusters
		cf.clustersMutex.Unlock()

		return true, nil
	}); err != nil {
		// TODO: log error
		return
	}
}

func (cf *ClientsFactory) watchNamespaces(ctx context.Context) {
	if err := wait.PollImmediateInfinite(30*time.Second, func() (bool, error) {
		clients, err := clientsForClusters(cf.clusters)
		if err != nil {
			cf.logger.Error(err, "failed to create clients for", "clusters", cf.clusters)
			return false, err
		}

		wg := sync.WaitGroup{}

		for clusterName, c := range clients {
			wg.Add(1)

			go func(clusterName string, c client.Client) {
				defer wg.Done()

				nsList := &v1.NamespaceList{}

				if err := c.List(ctx, nsList); err != nil {
					cf.logger.Error(err, "failed listing namespaces", "cluster", clusterName)
				}

				cf.clustersNsMutex.Lock()
				cf.clustersNamespaces[clusterName] = nsList.Items
				cf.clustersNsMutex.Unlock()
			}(clusterName, c)
		}

		wg.Wait()

		return false, nil
	}); err != nil {
		cf.logger.Error(err, "failed pooling namespaces")
	}
}

func (cf *ClientsFactory) GetUserClient(ctx context.Context, user *auth.UserPrincipal) (Client, error) {
	pool := NewClustersClientsPool()

	for _, cluster := range cf.clusters {
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

func (cf *ClientsFactory) userNsList(ctx context.Context, user *auth.UserPrincipal) map[string][]v1.Namespace {
	// if nss, ok := cf.usersNamespaces[user.ID]; ok {
	// 	return nss
	// }

	cf.userNssMutex.Lock()
	cf.usersNamespaces[user.ID] = map[string][]v1.Namespace{}
	cf.userNssMutex.Unlock()

	wg := sync.WaitGroup{}

	for _, cluster := range cf.clusters {
		wg.Add(1)

		go func(cluster Cluster) {
			defer wg.Done()

			cf.clustersNsMutex.RLock()
			clusterNs := cf.clustersNamespaces[cluster.Name]
			cf.clustersNsMutex.RUnlock()

			filteredNs, err := cf.nsChecker.FilterAccessibleNamespaces(ctx, impersonatedConfig(cluster, user), clusterNs)
			if err != nil {
				//todo: log error
			}

			fmt.Println("---------------------------------------------", filteredNs)
			cf.userNssMutex.Lock()
			cf.usersNamespaces[user.ID][cluster.Name] = filteredNs
			cf.userNssMutex.Unlock()
		}(cluster)
	}

	wg.Wait()

	return cf.usersNamespaces[user.ID]
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
			//todo: log error
		}

		clients[cluster.Name] = c
	}

	return clients, nil
}
