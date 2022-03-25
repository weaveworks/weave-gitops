package clustersmngr

import (
	"context"
	"errors"
	"fmt"
	"sync"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
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
}

type clustersClient struct {
	pool ClientsPool
}

func NewClient(clientsPool ClientsPool) Client {
	return &clustersClient{
		pool: clientsPool,
	}
}

func (c *clustersClient) Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error {
	cClient := c.pool.Clients()[cluster]
	if cClient == nil {
		return ErrClusterNotFound
	}

	return cClient.Get(ctx, key, obj)
}

type ClusteredObjectListError struct {
	errs []error
}

func (e *ClusteredObjectListError) Add(err error) {
	e.errs = append(e.errs, err)
}

func (e *ClusteredObjectListError) Error() string {
	if len(e.errs) > 0 {
		return fmt.Sprintf("listing resources on clusters failed: %s", e.errs)
	}

	return ""
}

func (c *clustersClient) List(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error {
	wg := sync.WaitGroup{}

	var errs []error

	for clusterName, c := range c.pool.Clients() {
		wg.Add(1)

		go func(clusterName string, c client.Client) {
			defer wg.Done()

			list := clist.NewObjectList()

			if err := c.List(ctx, list, opts...); err != nil {
				errs = append(errs, fmt.Errorf("cluster=\"%s\" err=\"%s\"", clusterName, err))
			}

			clist.SetObjectList(clusterName, list)
		}(clusterName, c)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to list resources: %s", errs)
	}

	return nil
}

func (c *clustersClient) Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error {
	cClient := c.pool.Clients()[cluster]
	if cClient == nil {
		return ErrClusterNotFound
	}

	return cClient.Create(ctx, obj, opts...)
}

func (c *clustersClient) Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error {
	cClient := c.pool.Clients()[cluster]
	if cClient == nil {
		return ErrClusterNotFound
	}

	return cClient.Delete(ctx, obj, opts...)
}

func (c *clustersClient) Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error {
	cClient := c.pool.Clients()[cluster]
	if cClient == nil {
		return ErrClusterNotFound
	}

	return cClient.Update(ctx, obj, opts...)
}

func (c *clustersClient) Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	cClient := c.pool.Clients()[cluster]
	if cClient == nil {
		return ErrClusterNotFound
	}

	return cClient.Patch(ctx, obj, patch, opts...)
}

type ClusteredObjectList interface {
	NewObjectList() client.ObjectList
	SetObjectList(cluster string, list client.ObjectList)
}

type ClusteredKustomizationList struct {
	sync.Mutex

	Lists map[string]*kustomizev1.KustomizationList
}

func (ckl *ClusteredKustomizationList) NewObjectList() client.ObjectList {
	return &kustomizev1.KustomizationList{}
}

func (ckl *ClusteredKustomizationList) SetObjectList(cluster string, list client.ObjectList) {
	ckl.Lock()
	defer ckl.Unlock()

	if ckl.Lists == nil {
		ckl.Lists = make(map[string]*kustomizev1.KustomizationList)
	}

	ckl.Lists[cluster] = list.(*kustomizev1.KustomizationList)
}

type ClusteredGitRepositoryList struct {
	sync.Mutex

	Lists map[string]*sourcev1.GitRepositoryList
}

func (ckl *ClusteredGitRepositoryList) NewObjectList() client.ObjectList {
	return &kustomizev1.KustomizationList{}
}

func (ckl *ClusteredGitRepositoryList) SetObjectList(cluster string, list client.ObjectList) {
	ckl.Lock()
	defer ckl.Unlock()

	if ckl.Lists == nil {
		ckl.Lists = make(map[string]*sourcev1.GitRepositoryList)
	}

	ckl.Lists[cluster] = list.(*sourcev1.GitRepositoryList)
}
