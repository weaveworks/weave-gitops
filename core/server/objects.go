package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getUnstructuredHelmReleaseInventory(ctx context.Context, obj unstructured.Unstructured, c clustersmngr.Client, cluster string) ([]*pb.GroupVersionKind, error) {
	var release v2beta1.HelmRelease

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &release)
	if err != nil {
		return nil, fmt.Errorf("converting unstructured to helmrelease: %w", err)
	}

	inventory, err := getHelmReleaseInventory(ctx, release, c, cluster)
	if err != nil {
		return nil, fmt.Errorf("get helmrelease inventory: %w", err)
	}

	return inventory, nil
}

func (cs *coreServer) ListObjects(ctx context.Context, msg *pb.ListObjectsRequest) (*pb.ListObjectsResponse, error) {
	respErrors := []*pb.ListError{}

	gvk, err := cs.primaryKinds.Lookup(msg.Kind)
	if err != nil {
		return nil, err
	}

	var clustersClient clustersmngr.Client

	if msg.ClusterName != "" {
		clustersClient, err = cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), msg.ClusterName)
	} else {
		clustersClient, err = cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	}

	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		list := unstructured.UnstructuredList{}
		list.SetGroupVersionKind(*gvk)
		return &list
	})

	if err := clustersClient.ClusteredList(ctx, clist, true, client.InNamespace(msg.Namespace)); err != nil {
		var errs clustersmngr.ClusteredListError
		if !errors.As(err, &errs) {
			return nil, err
		}

		for _, e := range errs.Errors {
			respErrors = append(respErrors, &pb.ListError{ClusterName: e.Cluster, Namespace: e.Namespace, Message: e.Err.Error()})
		}
	}

	var results []*pb.Object

	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*unstructured.UnstructuredList)
			if !ok {
				continue
			}

			for _, object := range list.Items {
				tenant := GetTenant(object.GetNamespace(), n, clusterUserNamespaces)

				var inventory []*pb.GroupVersionKind = nil

				if gvk.Kind == v2beta1.HelmReleaseKind {
					inventory, err = getUnstructuredHelmReleaseInventory(ctx, object, clustersClient, n)
					if err != nil {
						respErrors = append(respErrors, &pb.ListError{ClusterName: n, Message: err.Error()})
						inventory = nil // We can still display most things without inventory

						cs.logger.V(logger.LogLevelDebug).Info("Couldn't grab inventory for helm release", "error", err)
					}
				}

				o, err := types.K8sObjectToProto(&object, n, tenant, inventory)
				if err != nil {
					respErrors = append(respErrors, &pb.ListError{ClusterName: n, Message: "converting items: " + err.Error()})
					continue
				}

				results = append(results, o)
			}
		}
	}

	return &pb.ListObjectsResponse{
		Objects: results,
		Errors:  respErrors,
	}, nil
}

func (cs *coreServer) GetObject(ctx context.Context, msg *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	gvk, err := cs.primaryKinds.Lookup(msg.Kind)
	if err != nil {
		return nil, err
	}

	obj := unstructured.Unstructured{}
	obj.SetGroupVersionKind(*gvk)

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := clustersClient.Get(ctx, msg.ClusterName, key, &obj); err != nil {
		return nil, err
	}

	var inventory []*pb.GroupVersionKind = nil

	if gvk.Kind == v2beta1.HelmReleaseKind {
		inventory, err = getUnstructuredHelmReleaseInventory(ctx, obj, clustersClient, msg.ClusterName)
		if err != nil {
			inventory = nil // We can still display most things without inventory

			cs.logger.V(logger.LogLevelDebug).Info("Couldn't grab inventory for helm release", "error", err)
		}
	}

	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))

	tenant := GetTenant(obj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

	res, err := types.K8sObjectToProto(&obj, msg.ClusterName, tenant, inventory)

	if err != nil {
		return nil, fmt.Errorf("converting object to proto: %w", err)
	}

	return &pb.GetObjectResponse{Object: res}, nil
}
