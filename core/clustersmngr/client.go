package clustersmngr

import (
	"context"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client thin wrapper to controller-runtime/client  adding multi clusters context.
type Client interface {
	// Get retrieves an obj for the given object key.
	Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error
	// List retrieves list of objects for a given namespace and list options.
	List(ctx context.Context, cluster string, list client.ObjectList, opts ...client.ListOption) error

	// Create saves the object obj.
	Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error
	// Delete deletes the given obj
	Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error
	// Update updates the given obj.
	Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error
	// Patch patches the given obj
	Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error

	// ClusteredList retrieves list of objects for all clusters.
	ClusteredList(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error

	// ClientsPool returns the clients pool.
	ClientsPool() ClientsPool

	// RestConfig returns a rest.Config for a given cluster
	RestConfig(cluster string) (*rest.Config, error)

	// Scoped returns a client that is scoped to a single cluster
	Scoped(cluster string) (client.Client, error)
}

type clustersClient struct {
	pool       ClientsPool
	namespaces map[string][]v1.Namespace
}

func NewClient(clientsPool ClientsPool, namespaces map[string][]v1.Namespace) Client {
	return &clustersClient{
		pool:       clientsPool,
		namespaces: namespaces,
	}
}

func (c *clustersClient) ClientsPool() ClientsPool {
	return c.pool
}

func (c *clustersClient) Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Get(ctx, key, obj)
}

func (c *clustersClient) List(ctx context.Context, cluster string, list client.ObjectList, opts ...client.ListOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.List(ctx, list, opts...)
}

func (c *clustersClient) ClusteredList(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error {
	wg := sync.WaitGroup{}

	var errs []error

	for clusterName, cc := range c.pool.Clients() {
		for _, ns := range c.namespaces[clusterName] {
			wg.Add(1)

			go func(clusterName string, c client.Client, optsWithNamespace ...client.ListOption) {
				defer wg.Done()

				list := clist.ObjectList(clusterName)

				if err := c.List(ctx, list, optsWithNamespace...); err != nil {
					errs = append(errs, fmt.Errorf("cluster=\"%s\" err=\"%s\"", clusterName, err))
				}
			}(clusterName, cc, append(opts, client.InNamespace(ns.Name))...)
		}
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to list resources: %s", errs)
	}

	return nil
}

func (c *clustersClient) Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Create(ctx, obj, opts...)
}

func (c *clustersClient) Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Delete(ctx, obj, opts...)
}

func (c *clustersClient) Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Update(ctx, obj, opts...)
}

func (c *clustersClient) Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Patch(ctx, obj, patch, opts...)
}

func (c clustersClient) RestConfig(cluster string) (*rest.Config, error) {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return nil, err
	}

	return client.RestConfig(), nil
}

func (c clustersClient) Scoped(cluster string) (client.Client, error) {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type ClusteredObjectList interface {
	ObjectList(cluster string) client.ObjectList
	Lists() map[string]client.ObjectList
}

type ClusteredList struct {
	sync.Mutex

	listFactory func() client.ObjectList
	lists       map[string]client.ObjectList
}

func NewClusteredList(listFactory func() client.ObjectList) ClusteredObjectList {
	return &ClusteredList{
		listFactory: listFactory,
		lists:       make(map[string]client.ObjectList),
	}
}

func (cl *ClusteredList) ObjectList(cluster string) client.ObjectList {
	cl.Lock()
	defer cl.Unlock()

	cl.lists[cluster] = cl.listFactory()

	return cl.lists[cluster]
}

func (cl *ClusteredList) Lists() map[string]client.ObjectList {
	cl.Lock()
	defer cl.Unlock()

	return cl.lists
}
