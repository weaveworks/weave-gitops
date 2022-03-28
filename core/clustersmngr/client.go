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

type Client interface {
	Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error
	List(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error

	Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error
	Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error
	Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error
	Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error

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

func (c *clustersClient) List(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error {
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
}

type ClusteredList[T client.ObjectList] struct {
	sync.Mutex

	listFactory func() T
	lists       map[string]T
}

func NewClusteredList[T client.ObjectList](listFactory func() T) *ClusteredList[T] {
	return &ClusteredList[T]{
		lists:       make(map[string]T),
		listFactory: listFactory,
	}
}

func (cl *ClusteredList[T]) ObjectList(cluster string) client.ObjectList {
	cl.Lock()
	defer cl.Unlock()

	cl.lists[cluster] = cl.listFactory()

	return cl.lists[cluster]
}

func (cl *ClusteredList[T]) Lists() map[string]T {
	cl.Lock()
	defer cl.Unlock()

	return cl.lists
}

func (cl *ClusteredList[T]) List(cluster string) T {
	cl.Lock()
	defer cl.Unlock()

	return cl.lists[cluster]
}
