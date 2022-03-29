package clustersmngr

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrClusterNotFound error = errors.New("cluster not found")
)

// Client mimics the client interface of controller-runtime as much as possible but adding multi clusters context.
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
}

type clustersClient struct {
	pool ClientsPool
}

func NewClient(clientsPool ClientsPool) Client {
	return &clustersClient{
		pool: clientsPool,
	}
}

func (c *clustersClient) ClientsPool() ClientsPool {
	return c.pool
}

func (c *clustersClient) Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error {
	client := c.pool.Clients()[cluster]
	if client == nil {
		return ErrClusterNotFound
	}

	return client.Get(ctx, key, obj)
}

func (c *clustersClient) List(ctx context.Context, cluster string, list client.ObjectList, opts ...client.ListOption) error {
	client := c.pool.Clients()[cluster]
	if client == nil {
		return ErrClusterNotFound
	}

	return client.List(ctx, list, opts...)
}

func (c *clustersClient) ClusteredList(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error {
	wg := sync.WaitGroup{}

	var errs []error

	for clusterName, c := range c.pool.Clients() {
		wg.Add(1)

		go func(clusterName string, c client.Client) {
			defer wg.Done()

			list := clist.ObjectList(clusterName)

			if err := c.List(ctx, list, opts...); err != nil {
				errs = append(errs, fmt.Errorf("cluster=\"%s\" err=\"%s\"", clusterName, err))
			}
		}(clusterName, c)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to list resources: %s", errs)
	}

	return nil
}

func (c *clustersClient) Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error {
	client := c.pool.Clients()[cluster]
	if client == nil {
		return ErrClusterNotFound
	}

	return client.Create(ctx, obj, opts...)
}

func (c *clustersClient) Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error {
	client := c.pool.Clients()[cluster]
	if client == nil {
		return ErrClusterNotFound
	}

	return client.Delete(ctx, obj, opts...)
}

func (c *clustersClient) Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error {
	client := c.pool.Clients()[cluster]
	if client == nil {
		return ErrClusterNotFound
	}

	return client.Update(ctx, obj, opts...)
}

func (c *clustersClient) Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	client := c.pool.Clients()[cluster]
	if client == nil {
		return ErrClusterNotFound
	}

	return client.Patch(ctx, obj, patch, opts...)
}

type ClusteredObjectList interface {
	ObjectList(cluster string) client.ObjectList
	Lists() map[string]interface{}
}

type ClusteredList struct {
	sync.Mutex

	listFactory func() client.ObjectList
	lists       map[string]interface{}
}

func NewClusteredList(listFactory func() client.ObjectList) ClusteredObjectList {
	return &ClusteredList{
		listFactory: listFactory,
		lists:       make(map[string]interface{}),
	}
}

func (cl *ClusteredList) ObjectList(cluster string) client.ObjectList {
	cl.Lock()
	defer cl.Unlock()

	cl.lists[cluster] = cl.listFactory()

	return cl.lists[cluster].(client.ObjectList)
}

func (cl *ClusteredList) Lists() map[string]interface{} {
	cl.Lock()
	defer cl.Unlock()

	return cl.lists
}
