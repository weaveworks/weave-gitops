package clustersmngr

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	userNamespaceTTL = 30 * time.Second
	// How often we need to stop the world and remove outdated records.
	userNamespaceResolution = 30 * time.Second
	watchClustersFrequency  = 30 * time.Second
	watchNamespaceFrequency = 30 * time.Second
	usersClientResolution   = 30 * time.Second
)

var (
	usersClientsTTL = getEnvDuration("WEAVE_GITOPS_USERS_CLIENTS_TTL", 30*time.Minute)
)

func getEnvDuration(key string, defaultDuration time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultDuration
	}

	d, err := time.ParseDuration(val)

	// on error return the default duration
	if err != nil {
		return defaultDuration
	}

	return d
}

var (
	opsUpdateClusters = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "gitops",
			Subsystem: "clustersmngr",
			Name:      "update_clusters_total",
			Help:      "The number of times clusters have been refreshed",
		})
	opsClustersCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gitops",
			Subsystem: "clustersmngr",
			Name:      "clusters_total",
			Help:      "The number of clusters currently being tracked",
		})
	opsUpdateNamespaces = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "gitops",
			Subsystem: "clustersmngr",
			Name:      "update_namespaces_total",
			Help:      "The number of times namespaces have been refreshed",
		})
	opsNamespacesCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gitops",
			Subsystem: "clustersmngr",
			Name:      "namespaces_count",
			Help:      "The number of namespaces currently being tracked",
		},
		[]string{
			// Which cluster has these namespaces
			"cluster",
		},
	)
	opsCreateServerClient = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gitops",
			Subsystem: "clustersmngr",
			Name:      "create_server_client_total",
			Help:      "The number of times a server client has been created",
		},
		[]string{
			// Which cluster created the client
			"cluster",
		},
	)
	opsCreateUserClient = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gitops",
			Subsystem: "clustersmngr",
			Name:      "create_user_client_total",
			Help:      "The number of times a server client has been created",
		},
		[]string{
			// Which cluster created the client
			"cluster",
		},
	)

	Registry = prometheus.NewRegistry()
)

func registerMetrics() {
	_ = Registry.Register(opsUpdateClusters)
	_ = Registry.Register(opsClustersCount)
	_ = Registry.Register(opsUpdateNamespaces)
	_ = Registry.Register(opsNamespacesCount)
	_ = Registry.Register(opsCreateServerClient)
	_ = Registry.Register(opsCreateUserClient)
}

// ClientError is an error returned by the GetImpersonatedClient function which contains
// the details of the cluster that caused the error.
type ClientError struct {
	ClusterName string
	Err         error
}

// Error() returns the error message of the underlying error.
func (ce *ClientError) Error() string {
	return ce.Err.Error()
}

// ClustersManager is a manager for creating clients for clusters
//
//counterfeiter:generate . ClustersManager
type ClustersManager interface {
	// GetImpersonatedClient returns the clusters client for the given user
	GetImpersonatedClient(ctx context.Context, user *auth.UserPrincipal) (Client, error)
	// GetImpersonatedClientForCluster returns the client for the given user and cluster
	GetImpersonatedClientForCluster(ctx context.Context, user *auth.UserPrincipal, clusterName string) (Client, error)
	// GetImpersonatedDiscoveryClient returns the discovery for the given user and for the given cluster
	GetImpersonatedDiscoveryClient(ctx context.Context, user *auth.UserPrincipal, clusterName string) (discovery.DiscoveryInterface, error)
	// UpdateClusters updates the clusters list
	UpdateClusters(ctx context.Context) error
	// UpdateNamespaces updates the namespaces all namespaces for all clusters
	UpdateNamespaces(ctx context.Context) error
	// UpdateUserNamespaces updates the cache of accessible namespaces for the user
	UpdateUserNamespaces(ctx context.Context, user *auth.UserPrincipal)
	// GetServerClient returns the cluster client with gitops server permissions
	GetServerClient(ctx context.Context) (Client, error)
	// GetClustersNamespaces returns the namespaces for all clusters
	GetClustersNamespaces() map[string][]v1.Namespace
	// GetUserNamespaces returns the accessible namespaces for the user
	GetUserNamespaces(user *auth.UserPrincipal) map[string][]v1.Namespace
	// Start starts go routines to keep clusters and namespaces lists up to date
	Start(ctx context.Context)
	// Subscribe returns a new ClustersWatcher
	Subscribe() *ClustersWatcher
	// RemoveWatcher removes the given ClustersWatcher from the list of watchers
	RemoveWatcher(cw *ClustersWatcher)
	// GetClusters returns all the currently known clusters
	GetClusters() []cluster.Cluster
}

type clustersManager struct {
	clustersFetchers clusterFetchers
	nsChecker        nsaccess.Checker
	log              logr.Logger

	// list of clusters returned by the clusters fetcher
	clusters *Clusters
	// string containing ordered list of cluster names, used to refresh dependent caches
	clustersHash string
	// the lists of all namespaces of each cluster
	clustersNamespaces *ClustersNamespaces
	// lists of namespaces accessible by the user on every cluster
	usersNamespaces *UsersNamespaces
	usersClients    *UsersClients

	initialClustersLoad chan bool
	// list of watchers to notify of clusters updates
	watchers []*ClustersWatcher

	// whether to use the server client to lookup the list of namespaces on each
	// cluster or whether to use the user client. When using the user client, the
	// namespaces are looked up during the request
	useUserClientForNamespaces bool
}

// ClusterListUpdate records the changes to the cluster state managed by the factory.
type ClusterListUpdate struct {
	Added   []cluster.Cluster
	Removed []cluster.Cluster
}

// ClustersWatcher watches for cluster list updates and notifies the registered clients.
type ClustersWatcher struct {
	Updates chan ClusterListUpdate
	cf      *clustersManager
}

// Notify publishes cluster updates to the current watcher.
func (cw *ClustersWatcher) Notify(addedClusters, removedClusters []cluster.Cluster) {
	cw.Updates <- ClusterListUpdate{Added: addedClusters, Removed: removedClusters}
}

// Unsubscribe removes the given ClustersWatcher from the list of watchers.
func (cw *ClustersWatcher) Unsubscribe() {
	cw.cf.RemoveWatcher(cw)
	close(cw.Updates)
}

func NewClustersManager(fetchers []ClusterFetcher, nsChecker nsaccess.Checker, logger logr.Logger) ClustersManager {
	registerMetrics()

	useUserClientForNamespaces := featureflags.Get("WEAVE_GITOPS_FEATURE_USE_USER_CLIENT_FOR_NAMESPACES") == "true"
	logger.Info("Use user client for namespaces", "enabled", useUserClientForNamespaces)

	return &clustersManager{
		clustersFetchers:           fetchers,
		nsChecker:                  nsChecker,
		clusters:                   &Clusters{},
		clustersNamespaces:         &ClustersNamespaces{},
		usersNamespaces:            &UsersNamespaces{Cache: ttlcache.New(userNamespaceResolution)},
		usersClients:               &UsersClients{Cache: ttlcache.New(usersClientResolution)},
		log:                        logger,
		initialClustersLoad:        make(chan bool),
		watchers:                   []*ClustersWatcher{},
		useUserClientForNamespaces: useUserClientForNamespaces,
	}
}

// Subscribe returns a new ClustersWatcher.
func (cf *clustersManager) Subscribe() *ClustersWatcher {
	cw := &ClustersWatcher{cf: cf, Updates: make(chan ClusterListUpdate, 1)}
	cf.watchers = append(cf.watchers, cw)

	return cw
}

// RemoveWatcher removes the given ClustersWatcher from the list of watchers.
func (cf *clustersManager) RemoveWatcher(cw *ClustersWatcher) {
	watchers := []*ClustersWatcher{}
	for _, w := range cf.watchers {
		if cw != w {
			watchers = append(watchers, w)
		}
	}

	cf.watchers = watchers
}

func (cf *clustersManager) GetClusters() []cluster.Cluster {
	return cf.clusters.Get()
}

func (cf *clustersManager) Start(ctx context.Context) {
	go cf.watchClusters(ctx)

	if !cf.useUserClientForNamespaces {
		go cf.watchNamespaces(ctx)
	}
}

func (cf *clustersManager) watchClusters(ctx context.Context) {
	if err := cf.UpdateClusters(ctx); err != nil {
		cf.log.Error(err, "failed updating clusters")
	}

	cf.initialClustersLoad <- true

	if err := wait.PollImmediateInfinite(watchClustersFrequency, func() (bool, error) {
		if err := cf.UpdateClusters(ctx); err != nil {
			cf.log.Error(err, "Failed to update clusters")
		}

		return false, nil
	}); err != nil {
		cf.log.Error(err, "failed polling clusters")
	}
}

// UpdateClusters updates the clusters list and notifies the registered watchers.
func (cf *clustersManager) UpdateClusters(ctx context.Context) error {
	clusters, err := cf.clustersFetchers.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch clusters: %w", err)
	}

	addedClusters, removedClusters := cf.clusters.Set(clusters)

	opsUpdateClusters.Inc()
	opsClustersCount.Set(float64(len(clusters)))

	if len(addedClusters) > 0 || len(removedClusters) > 0 {
		// notify watchers of the changes
		for _, w := range cf.watchers {
			w.Notify(addedClusters, removedClusters)
		}
	}

	return nil
}

func (cf *clustersManager) watchNamespaces(ctx context.Context) {
	// waits the first load of cluster to start watching namespaces
	<-cf.initialClustersLoad

	if err := wait.PollImmediateInfinite(watchNamespaceFrequency, func() (bool, error) {
		if err := cf.UpdateNamespaces(ctx); err != nil {
			if merr, ok := err.(*multierror.Error); ok {
				for _, cerr := range merr.Errors {
					cf.log.Error(cerr, "failed to update namespaces")
				}
			}
		}

		return false, nil
	}); err != nil {
		cf.log.Error(err, "failed polling namespaces")
	}
}

// Used when all clusters have kubeconfig secrets that contain credentials to connect to the cluster
// and list their namespaces.
func (cf *clustersManager) UpdateNamespaces(ctx context.Context) error {
	return cf.updateNamespacesWithClient(ctx, func() (Client, error) {
		return cf.GetServerClient(ctx)
	})
}

// Used when clusters have kubeconfig secrets that DO NOT contain credentials to connect to the cluster.
// We need to use the user's token from the incoming request to connect to the cluster and list its namespaces
// instead.
func (cf *clustersManager) updateNamespacesFromUserClient(ctx context.Context, user *auth.UserPrincipal) error {
	return cf.updateNamespacesWithClient(ctx, func() (Client, error) {
		return cf.getUserClientWithNamespaces(ctx, user, false)
	})
}

func (cf *clustersManager) updateNamespacesWithClient(ctx context.Context, createClient func() (Client, error)) error {
	var result *multierror.Error

	clientset, err := createClient()
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*ClientError); ok {
					result = multierror.Append(result, fmt.Errorf("%w, cluster: %v", cerr, cerr.ClusterName))
				}
			}
		}
	}

	cf.syncCaches()

	nsList := NewClusteredList(func() client.ObjectList {
		return &v1.NamespaceList{}
	})

	if err := clientset.ClusteredList(ctx, nsList, false); err != nil {
		result = multierror.Append(result, err)
	}

	for clusterName, lists := range nsList.Lists() {
		// This is the "namespace loop", but namespaces aren't
		// namespaced so only 1 item
		for _, l := range lists {
			list, ok := l.(*v1.NamespaceList)
			if !ok {
				continue
			}

			cf.clustersNamespaces.Set(clusterName, list.Items)
			opsNamespacesCount.WithLabelValues(clusterName).Set(float64(len(list.Items)))
		}
	}

	opsUpdateNamespaces.Inc()

	return result.ErrorOrNil()
}

func (cf *clustersManager) GetClustersNamespaces() map[string][]v1.Namespace {
	return cf.clustersNamespaces.namespaces
}

func (cf *clustersManager) syncCaches() {
	newHash := cf.clusters.Hash()

	if newHash != cf.clustersHash {
		cf.log.Info("Clearing namespace caches")
		cf.clustersNamespaces.Clear()
		cf.usersNamespaces.Clear()
		cf.clustersHash = newHash
	}
}

func (cf *clustersManager) GetImpersonatedClient(ctx context.Context, user *auth.UserPrincipal) (Client, error) {
	return cf.getUserClientWithNamespaces(ctx, user, true)
}

// getUserClientWithNamespaces returns a client for the user and optionally fetches the namespaces
// the user has access to. If lookupNamespaces is false, the namespaces will not be fetched and the client will only be
// able to make non-namespaced requests. (e.g. listing the namespaces themselves etc).
func (cf *clustersManager) getUserClientWithNamespaces(ctx context.Context, user *auth.UserPrincipal, lookupNamespaces bool) (Client, error) {
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	pool := NewClustersClientsPool()
	errChan := make(chan error, len(cf.clusters.Get()))

	var wg sync.WaitGroup

	for _, cl := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster cluster.Cluster, pool ClientsPool, errChan chan error) {
			defer wg.Done()

			client, err := cf.getOrCreateClient(ctx, user, cluster)
			if err != nil {
				errChan <- &ClientError{ClusterName: cluster.GetName(), Err: fmt.Errorf("failed creating user client to pool: %w", err)}
				return
			}

			if err := pool.Add(client, cluster); err != nil {
				errChan <- &ClientError{ClusterName: cluster.GetName(), Err: fmt.Errorf("failed adding cluster client to pool: %w", err)}
			}
		}(cl, pool, errChan)
	}

	wg.Wait()
	close(errChan)

	var result *multierror.Error

	for err := range errChan {
		result = multierror.Append(result, err)
	}

	var nss map[string][]v1.Namespace
	if lookupNamespaces {
		nss = cf.userNsList(ctx, user)
	}

	return NewClient(pool, nss), result.ErrorOrNil()
}

func (cf *clustersManager) GetImpersonatedClientForCluster(ctx context.Context, user *auth.UserPrincipal, clusterName string) (Client, error) {
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	var cl cluster.Cluster

	pool := NewClustersClientsPool()
	clusters := cf.clusters.Get()

	for _, c := range clusters {
		if c.GetName() == clusterName {
			cl = c
			break
		}
	}

	if cl == nil {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	client, err := cf.getOrCreateClient(ctx, user, cl)
	if err != nil {
		return nil, fmt.Errorf("failed creating cluster client: %w", err)
	}

	if err := pool.Add(client, cl); err != nil {
		return nil, fmt.Errorf("failed adding cluster client to pool: %w", err)
	}

	return NewClient(pool, cf.userNsList(ctx, user)), nil
}

func (cf *clustersManager) GetImpersonatedDiscoveryClient(ctx context.Context, user *auth.UserPrincipal, clusterName string) (discovery.DiscoveryInterface, error) {
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	for _, cluster := range cf.clusters.Get() {
		if cluster.GetName() == clusterName {
			var err error

			clientset, err := cluster.GetUserClientset(user)
			if err != nil {
				return nil, fmt.Errorf("error creating client for cluster: %w", err)
			}

			return clientset.Discovery(), nil
		}
	}

	return nil, fmt.Errorf("cluster not found: %s", clusterName)
}

func (cf *clustersManager) GetServerClient(ctx context.Context) (Client, error) {
	pool := NewClustersClientsPool()
	errChan := make(chan error, len(cf.clusters.Get()))

	var wg sync.WaitGroup

	for _, cl := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster cluster.Cluster, pool ClientsPool, errChan chan error) {
			defer wg.Done()

			client, err := cf.getOrCreateClient(ctx, nil, cluster)
			if err != nil {
				errChan <- &ClientError{ClusterName: cluster.GetName(), Err: fmt.Errorf("failed creating server client to pool: %w", err)}
				return
			}

			if err := pool.Add(client, cluster); err != nil {
				errChan <- &ClientError{ClusterName: cluster.GetName(), Err: fmt.Errorf("failed adding cluster client to pool: %w", err)}
			}
		}(cl, pool, errChan)
	}

	wg.Wait()
	close(errChan)

	var result *multierror.Error

	for err := range errChan {
		result = multierror.Append(result, err)
	}

	return NewClient(pool, cf.clustersNamespaces.namespaces), result.ErrorOrNil()
}

func (cf *clustersManager) UpdateUserNamespaces(ctx context.Context, user *auth.UserPrincipal) {
	wg := sync.WaitGroup{}

	for _, cl := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster cluster.Cluster) {
			defer wg.Done()

			clusterNs := cf.clustersNamespaces.Get(cluster.GetName())

			clientset, err := cluster.GetUserClientset(user)
			if err != nil {
				cf.log.Error(err, "failed creating clientset", "cluster", cluster.GetName(), "user", user.ID)
				return
			}

			filteredNs, err := cf.nsChecker.FilterAccessibleNamespaces(ctx, clientset.AuthorizationV1(), clusterNs)
			if err != nil {
				cf.log.Error(err, "failed filtering namespaces", "cluster", cluster.GetName(), "user", user.ID)
				return
			}

			cf.usersNamespaces.Set(user, cluster.GetName(), filteredNs)
		}(cl)
	}

	wg.Wait()
}

func (cf *clustersManager) GetUserNamespaces(user *auth.UserPrincipal) map[string][]v1.Namespace {
	return cf.usersNamespaces.GetAll(user, cf.clusters.Get())
}

func (cf *clustersManager) userNsList(ctx context.Context, user *auth.UserPrincipal) map[string][]v1.Namespace {
	userNamespaces := cf.GetUserNamespaces(user)
	if len(userNamespaces) > 0 {
		return userNamespaces
	}

	if cf.useUserClientForNamespaces {
		err := cf.updateNamespacesFromUserClient(ctx, user)
		if err != nil {
			// This may not completely fail the request, e.g. if some of the clusters
			// are able to respond with their namespaces. So log the error and continue.
			cf.log.Error(err, "failed updating namespaces from user client")
		}
	}

	cf.UpdateUserNamespaces(ctx, user)

	return cf.GetUserNamespaces(user)
}

func (cf *clustersManager) getOrCreateClient(ctx context.Context, user *auth.UserPrincipal, cluster cluster.Cluster) (client.Client, error) {
	isServer := false

	if user == nil {
		user = &auth.UserPrincipal{
			ID: "weave-gitops-server",
		}
		isServer = true
	}

	if client, found := cf.usersClients.Get(user, cluster.GetName()); found {
		return client, nil
	}

	var (
		client client.Client
		err    error
	)

	if isServer {
		opsCreateServerClient.WithLabelValues(cluster.GetName()).Inc()
		client, err = cluster.GetServerClient()
	} else {
		opsCreateUserClient.WithLabelValues(cluster.GetName()).Inc()
		client, err = cluster.GetUserClient(user)
	}

	if err != nil {
		return nil, fmt.Errorf("failed creating client for cluster=%s: %w", cluster.GetName(), err)
	}

	cf.usersClients.Set(user, cluster.GetName(), client)

	return client, nil
}
