package clustersmngr

import (
	"context"
	"errors"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
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

func (c *clustersClient) List(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error {
	for clusterName, c := range c.pool.Clients() {
		list := clist.NewObjectList()

		if err := c.List(ctx, list, opts...); err != nil {
			return fmt.Errorf("failed to list resouces in cluster %s: %w", clusterName, err)
		}

		clist.SetObjectList(clusterName, list)
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
	Lists map[string]*kustomizev1.KustomizationList
}

func (ckl ClusteredKustomizationList) NewObjectList() client.ObjectList {
	return &kustomizev1.KustomizationList{}
}

func (ckl *ClusteredKustomizationList) SetObjectList(cluster string, list client.ObjectList) {
	if ckl.Lists == nil {
		ckl.Lists = make(map[string]*kustomizev1.KustomizationList)
	}

	ckl.Lists[cluster] = list.(*kustomizev1.KustomizationList)
}
