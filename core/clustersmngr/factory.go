package clustersmngr

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cli-utils/pkg/flowcontrol"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	userNamespaceTTL = 30 * time.Second
	// How often we need to stop the world and remove outdated records.
	userNamespaceResolution = 30 * time.Second
	watchClustersFrequency  = 30 * time.Second
	watchNamespaceFrequency = 30 * time.Second
	kubeClientDialTimeout   = 5 * time.Second
	kubeClientDialKeepAlive = 30 * time.Second
)

var (
	kubeClientTimeout = getEnvDuration("WEAVE_GITOPS_KUBE_CLIENT_TIMEOUT", 30*time.Second)
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
	GetImpersonatedDiscoveryClient(ctx context.Context, user *auth.UserPrincipal, clusterName string) (*discovery.DiscoveryClient, error)
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
}

var DefaultKubeConfigOptions = []KubeConfigOption{WithFlowControl}

type ClusterPoolFactoryFn func(*apiruntime.Scheme) ClientsPool
type KubeConfigOption func(*rest.Config) (*rest.Config, error)
type ClientFactoryFn func(ctx context.Context, user *auth.UserPrincipal, cfgFunc ClusterClientConfigFunc, cluster Cluster, scheme *apiruntime.Scheme) (client.Client, error)

type clustersManager struct {
	clustersFetcher ClusterFetcher
	nsChecker       nsaccess.Checker
	log             logr.Logger

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
	scheme              *apiruntime.Scheme
	newClustersPool     ClusterPoolFactoryFn
	createClient        ClientFactoryFn
	kubeConfigOptions   []KubeConfigOption
	// list of watchers to notify of clusters updates
	watchers []*ClustersWatcher

	usersLock sync.Map
}

// ClusterListUpdate records the changes to the cluster state managed by the factory.
type ClusterListUpdate struct {
	Added   []Cluster
	Removed []Cluster
}

// ClustersWatcher watches for cluster list updates and notifies the registered clients.
type ClustersWatcher struct {
	Updates chan ClusterListUpdate
	cf      *clustersManager
}

// Notify publishes cluster updates to the current watcher.
func (cw *ClustersWatcher) Notify(addedClusters, removedClusters []Cluster) {
	cw.Updates <- ClusterListUpdate{Added: addedClusters, Removed: removedClusters}
}

// Unsubscribe removes the given ClustersWatcher from the list of watchers.
func (cw *ClustersWatcher) Unsubscribe() {
	cw.cf.RemoveWatcher(cw)
	close(cw.Updates)
}

func NewClustersManager(fetcher ClusterFetcher, nsChecker nsaccess.Checker, logger logr.Logger, scheme *apiruntime.Scheme, clusterPoolFactory ClusterPoolFactoryFn, clientFactory ClientFactoryFn, kubeConfigOptions []KubeConfigOption) ClustersManager {
	return &clustersManager{
		clustersFetcher:     fetcher,
		nsChecker:           nsChecker,
		clusters:            &Clusters{},
		clustersNamespaces:  &ClustersNamespaces{},
		usersNamespaces:     &UsersNamespaces{Cache: ttlcache.New(userNamespaceResolution)},
		usersClients:        &UsersClients{Cache: ttlcache.New(userNamespaceResolution)},
		log:                 logger,
		initialClustersLoad: make(chan bool),
		scheme:              scheme,
		newClustersPool:     clusterPoolFactory,
		createClient:        clientFactory,
		kubeConfigOptions:   []KubeConfigOption{},
		watchers:            []*ClustersWatcher{},
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

func (cf *clustersManager) Start(ctx context.Context) {
	go cf.watchClusters(ctx)
	go cf.watchNamespaces(ctx)
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
	clusters, err := cf.clustersFetcher.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch clusters: %w", err)
	}

	addedClusters, removedClusters := cf.clusters.Set(clusters)

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

func (cf *clustersManager) UpdateNamespaces(ctx context.Context) error {
	var result *multierror.Error

	serverClient, err := cf.GetServerClient(ctx)
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

	if err := serverClient.ClusteredList(ctx, nsList, false); err != nil {
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
		}
	}

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
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	pool := cf.newClustersPool(cf.scheme)
	errChan := make(chan error, len(cf.clusters.Get()))

	var wg sync.WaitGroup

	for _, cluster := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster Cluster, pool ClientsPool, errChan chan error) {
			defer wg.Done()

			client, err := cf.getOrCreateClient(ctx, user, ClientConfigWithUser(user, cf.kubeConfigOptions...), cluster)
			if err != nil {
				errChan <- &ClientError{ClusterName: cluster.Name, Err: fmt.Errorf("failed creating user client to pool: %w", err)}
				return
			}

			if err := pool.Add(client, cluster); err != nil {
				errChan <- &ClientError{ClusterName: cluster.Name, Err: fmt.Errorf("failed adding cluster client to pool: %w", err)}
			}
		}(cluster, pool, errChan)
	}

	wg.Wait()
	close(errChan)

	var result *multierror.Error

	for err := range errChan {
		result = multierror.Append(result, err)
	}

	return NewClient(pool, cf.userNsList(ctx, user)), result.ErrorOrNil()
}

func (cf *clustersManager) GetImpersonatedClientForCluster(ctx context.Context, user *auth.UserPrincipal, clusterName string) (Client, error) {
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	pool := cf.newClustersPool(cf.scheme)
	clusters := cf.clusters.Get()

	var cl Cluster

	for _, c := range clusters {
		if c.Name == clusterName {
			cl = c
			break
		}
	}

	if cl.Name == "" {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	client, err := cf.getOrCreateClient(ctx, user, ClientConfigWithUser(user, cf.kubeConfigOptions...), cl)
	if err != nil {
		return nil, fmt.Errorf("failed creating cluster client to pool: %w", err)
	}

	if err := pool.Add(client, cl); err != nil {
		return nil, fmt.Errorf("failed adding cluster client to pool: %w", err)
	}

	return NewClient(pool, cf.userNsList(ctx, user)), nil
}

func (cf *clustersManager) GetImpersonatedDiscoveryClient(ctx context.Context, user *auth.UserPrincipal, clusterName string) (*discovery.DiscoveryClient, error) {
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	var config *rest.Config

	for _, cluster := range cf.clusters.Get() {
		if cluster.Name == clusterName {
			var err error

			config, err = ClientConfigWithUser(user, cf.kubeConfigOptions...)(cluster)
			if err != nil {
				return nil, fmt.Errorf("error creating client for cluster: %w", err)
			}

			break
		}
	}

	if config == nil {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client for config: %w", err)
	}

	return dc, nil
}

func (cf *clustersManager) GetServerClient(ctx context.Context) (Client, error) {
	pool := cf.newClustersPool(cf.scheme)
	errChan := make(chan error, len(cf.clusters.Get()))

	var wg sync.WaitGroup

	for _, cluster := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster Cluster, pool ClientsPool, errChan chan error) {
			defer wg.Done()

			client, err := cf.getOrCreateClient(ctx, nil, ClientConfigAsServer(cf.kubeConfigOptions...), cluster)
			if err != nil {
				errChan <- &ClientError{ClusterName: cluster.Name, Err: fmt.Errorf("failed creating server client to pool: %w", err)}
				return
			}

			if err := pool.Add(client, cluster); err != nil {
				errChan <- &ClientError{ClusterName: cluster.Name, Err: fmt.Errorf("failed adding cluster client to pool: %w", err)}
			}
		}(cluster, pool, errChan)
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

	for _, cluster := range cf.clusters.Get() {
		wg.Add(1)

		go func(cluster Cluster) {
			defer wg.Done()

			clusterNs, found := cf.clustersNamespaces.Get(cluster.Name)
			if !found {
				cf.log.Error(nil, "failed to get cluster namespaces", "cluster", cluster.Name)
				return
			}

			cfg, err := ClientConfigWithUser(user, cf.kubeConfigOptions...)(cluster)
			if err != nil {
				cf.log.Error(err, "failed creating client config", "cluster", cluster.Name, "user", user.ID)
				return
			}

			filteredNs, err := cf.nsChecker.FilterAccessibleNamespaces(ctx, cfg, clusterNs)
			if err != nil {
				cf.log.Error(err, "failed filtering namespaces", "cluster", cluster.Name, "user", user.ID)
				return
			}

			cf.usersNamespaces.Set(user, cluster.Name, filteredNs)
		}(cluster)
	}

	wg.Wait()
}

func (cf *clustersManager) UserLock(user *auth.UserPrincipal) *sync.Mutex {
	actual, _ := cf.usersLock.LoadOrStore(user.Hash(), &sync.Mutex{})
	lock := actual.(*sync.Mutex)
	lock.Lock()
	return lock
}

func (cf *clustersManager) GetUserNamespaces(user *auth.UserPrincipal) map[string][]v1.Namespace {
	return cf.usersNamespaces.GetAll(user, cf.clusters.Get())
}

func (cf *clustersManager) userNsList(ctx context.Context, user *auth.UserPrincipal) map[string][]v1.Namespace {
	userLock := cf.UserLock(user)
	defer userLock.Unlock()

	userNamespaces := cf.GetUserNamespaces(user)
	if len(userNamespaces) > 0 {
		return userNamespaces
	}

	cf.UpdateUserNamespaces(ctx, user)

	return cf.GetUserNamespaces(user)
}

func (cf *clustersManager) getOrCreateClient(ctx context.Context, user *auth.UserPrincipal, cfgFunc ClusterClientConfigFunc, cluster Cluster) (client.Client, error) {
	if user == nil {
		user = &auth.UserPrincipal{
			ID: "weave-gitops-server",
		}
	}

	if client, found := cf.usersClients.Get(user, cluster.Name); found {
		return client, nil
	}

	client, err := cf.createClient(ctx, user, cfgFunc, cluster, cf.scheme)
	if err != nil {
		return nil, fmt.Errorf("failed creating client for cluster=%s: %w", cluster.Name, err)
	}

	cf.usersClients.Set(user, cluster.Name, client)

	return client, nil
}

func ClientFactory(ctx context.Context, user *auth.UserPrincipal, cfgFunc ClusterClientConfigFunc, cluster Cluster, scheme *apiruntime.Scheme) (client.Client, error) {
	config, err := cfgFunc(cluster)
	if err != nil {
		return nil, fmt.Errorf("error building cluster client config: %w", err)
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return nil, fmt.Errorf("could not create RESTMapper from config: %w", err)
	}

	client, err := client.New(config, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create leaf client: %w", err)
	}

	return client, nil
}

func CachedClientFactory(ctx context.Context, user *auth.UserPrincipal, cfgFunc ClusterClientConfigFunc, cluster Cluster, scheme *apiruntime.Scheme) (client.Client, error) {
	config, err := cfgFunc(cluster)
	if err != nil {
		return nil, fmt.Errorf("error building cluster client config: %w", err)
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return nil, fmt.Errorf("could not create RESTMapper from config: %w", err)
	}

	leafClient, err := client.New(config, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create leaf client: %w", err)
	}

	cache, err := cache.New(config, cache.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating client cache: %w", err)
	}

	delegatingCache := NewDelegatingCache(leafClient, cache, scheme)

	delegatingClient, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader: delegatingCache,
		Client:      leafClient,
		// Non-exact field matches are not supported by the cache.
		// https://github.com/kubernetes-sigs/controller-runtime/issues/612
		// TODO: Research if we can change the way we query those events so we can enable the cache for it.
		UncachedObjects:   []client.Object{&corev1.Event{}},
		CacheUnstructured: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating DelegatingClient: %w", err)
	}

	ctx = context.Background()

	go delegatingCache.Start(ctx)

	if ok := delegatingCache.WaitForCacheSync(ctx); !ok {
		return nil, errors.New("failed syncing client cache")
	}

	return delegatingClient, nil
}

func ApplyKubeConfigOptions(config *rest.Config, options ...KubeConfigOption) (*rest.Config, error) {
	for _, o := range options {
		_, err := o(config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

// restConfigFromCluster creates a generic rest.Config for a given cluster.
// You should not call this directly, but rather via
// ClientConfigAsServer or ClientConfigWithUser
func restConfigFromCluster(cluster Cluster) *rest.Config {
	return &rest.Config{
		Host:            cluster.Server,
		TLSClientConfig: cluster.TLSConfig,
		QPS:             ClientQPS,
		Burst:           ClientBurst,
		Timeout:         kubeClientTimeout,
		Dial: (&net.Dialer{
			Timeout: kubeClientDialTimeout,
			// KeepAlive is default to 30s within client-go.
			KeepAlive: kubeClientDialKeepAlive,
		}).DialContext,
	}
}

func WithFlowControl(config *rest.Config) (*rest.Config, error) {
	// flowcontrol.IsEnabled makes a request to the K8s API of the cluster stored in the config.
	// It does a HEAD request to /livez/ping which uses the config.Dial timeout. We can use this
	// function to error early rather than wait to call client.New.
	enabled, err := flowcontrol.IsEnabled(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("error querying cluster for flowcontrol config: %w", err)
	}

	if enabled {
		// Enabled & negative QPS and Burst indicate that the client would use the rate limit set by the server.
		// Ref: https://github.com/kubernetes/kubernetes/blob/v1.24.0/staging/src/k8s.io/client-go/rest/config.go#L354-L364
		config.QPS = -1
		config.Burst = -1

		return config, nil
	}

	config.QPS = ClientQPS
	config.Burst = ClientBurst

	return config, nil
}

// clientConfigAsServer returns a *rest.Config for a given cluster
// as the server service acconut
func ClientConfigAsServer(options ...KubeConfigOption) ClusterClientConfigFunc {
	return func(cluster Cluster) (*rest.Config, error) {
		config := restConfigFromCluster(cluster)

		config.BearerToken = cluster.BearerToken

		return ApplyKubeConfigOptions(config, options...)
	}
}

// ClientConfigWithUser returns a function that returns a *rest.Config with the relevant
// user authentication details pre-defined for a given cluster.
func ClientConfigWithUser(user *auth.UserPrincipal, options ...KubeConfigOption) ClusterClientConfigFunc {
	return func(cluster Cluster) (*rest.Config, error) {
		config := restConfigFromCluster(cluster)

		if !user.Valid() {
			return nil, fmt.Errorf("no user ID or Token found in UserPrincipal")
		} else if tok := user.Token(); tok != "" {
			config.BearerToken = tok
		} else {
			config.BearerToken = cluster.BearerToken
			config.Impersonate = rest.ImpersonationConfig{
				UserName: user.ID,
				Groups:   user.Groups,
			}
		}

		return ApplyKubeConfigOptions(config, options...)
	}
}
