package clustersmngr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
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
)

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

type ClusterPoolFactoryFn func() ClientsPool

type clustersManager struct {
	nsChecker nsaccess.Checker
	log       logr.Logger

	// a collection of the clusters we manage
	clusters ClusterCollection
	// string containing ordered list of cluster names, used to refresh dependent caches
	clustersHash string
	// the lists of all namespaces of each cluster
	clustersNamespaces *ClustersNamespaces
	// lists of namespaces accessible by the user on every cluster
	usersNamespaces *UsersNamespaces

	newClustersPool ClusterPoolFactoryFn
	// list of watchers to notify of clusters updates
	watchers []*ClustersWatcher
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

func NewClustersManager(clusters ClusterCollection, nsChecker nsaccess.Checker, logger logr.Logger) ClustersManager {
	return &clustersManager{
		nsChecker:          nsChecker,
		clusters:           clusters,
		clustersNamespaces: &ClustersNamespaces{},
		usersNamespaces:    &UsersNamespaces{Cache: ttlcache.New(userNamespaceResolution)},
		log:                logger,
		newClustersPool:    NewClustersClientsPool,
		watchers:           []*ClustersWatcher{},
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
	return cf.clusters.GetAll()
}

func (cf *clustersManager) Start(ctx context.Context) {
	go cf.watchClusters(ctx)
	go cf.watchNamespaces(ctx)
}

func (cf *clustersManager) watchClusters(ctx context.Context) {
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
	addedClusters, removedClusters, err := cf.clusters.Update(ctx)
	if err != nil {
		return fmt.Errorf("failed to update clusters: %w", err)
	}

	if len(addedClusters) > 0 || len(removedClusters) > 0 {
		// notify watchers of the changes
		for _, w := range cf.watchers {
			w.Notify(addedClusters, removedClusters)
		}
	}

	return nil
}

func (cf *clustersManager) watchNamespaces(ctx context.Context) {
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
		fmt.Printf("Ugh, %+v\n", clusterName)
		fmt.Printf("Ugh2 %+v\n", cf.clusters)
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

	pool := cf.newClustersPool()
	errChan := make(chan error, len(cf.clusters.GetAll()))

	var wg sync.WaitGroup

	for _, cl := range cf.clusters.GetAll() {
		wg.Add(1)

		go func(cluster cluster.Cluster, pool ClientsPool, errChan chan error) {
			defer wg.Done()

			client, err := cluster.GetUserClient(user)
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

	return NewClient(pool, cf.userNsList(ctx, user)), result.ErrorOrNil()
}

func (cf *clustersManager) GetImpersonatedClientForCluster(ctx context.Context, user *auth.UserPrincipal, clusterName string) (Client, error) {
	if user == nil {
		return nil, errors.New("no user supplied")
	}

	var cl cluster.Cluster

	pool := cf.newClustersPool()
	clusters := cf.clusters.GetAll()

	for _, c := range clusters {
		if c.GetName() == clusterName {
			cl = c
			break
		}
	}

	if cl.GetName() == "" {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	client, err := cl.GetUserClient(user)
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

	for _, cluster := range cf.clusters.GetAll() {
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
	pool := cf.newClustersPool()
	errChan := make(chan error, len(cf.clusters.GetAll()))

	var wg sync.WaitGroup

	for _, cl := range cf.clusters.GetAll() {
		wg.Add(1)

		go func(cluster cluster.Cluster, pool ClientsPool, errChan chan error) {
			defer wg.Done()

			client, err := cluster.GetServerClient()
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

	for _, cl := range cf.clusters.GetAll() {
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
	return cf.usersNamespaces.GetAll(user, cf.clusters.GetAll())
}

func (cf *clustersManager) userNsList(ctx context.Context, user *auth.UserPrincipal) map[string][]v1.Namespace {
	userNamespaces := cf.GetUserNamespaces(user)
	if len(userNamespaces) > 0 {
		return userNamespaces
	}

	cf.UpdateUserNamespaces(ctx, user)

	return cf.GetUserNamespaces(user)
}
