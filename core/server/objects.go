package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sessionObjectsInfo = "session objects created"
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

	listOptions := []client.ListOption{
		client.InNamespace(msg.Namespace),
	}
	if len(msg.Labels) > 0 {
		listOptions = append(listOptions, client.MatchingLabels(msg.Labels))
	}

	if err := clustersClient.ClusteredList(ctx, clist, true, listOptions...); err != nil {
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

			for _, unstructuredObj := range list.Items {
				tenant := GetTenant(unstructuredObj.GetNamespace(), n, clusterUserNamespaces)

				var obj client.Object = &unstructuredObj

				var inventory []*pb.GroupVersionKind = nil
				var info string

				switch gvk.Kind {
				case "Secret":
					obj, err = sanitizeSecret(&unstructuredObj)
					if err != nil {
						respErrors = append(respErrors, &pb.ListError{ClusterName: n, Message: fmt.Sprintf("error sanitizing secrets: %v", err)})
						continue
					}
				case v2beta1.HelmReleaseKind:
					inventory, err = getUnstructuredHelmReleaseInventory(ctx, unstructuredObj, clustersClient, n)
					if err != nil {
						respErrors = append(respErrors, &pb.ListError{ClusterName: n, Message: err.Error()})
						inventory = nil // We can still display most things without inventory

						cs.logger.V(logger.LogLevelDebug).Info("Couldn't grab inventory for helm release", "error", err)
					}
				case "StatefulSet":
					clusterName, kind, err := parseSessionInfo(unstructuredObj)
					if err != nil {
						break
					}

					created, _ := cs.sessionObjectsCreated(ctx, clusterName, "flux-system", kind)

					if created {
						info = sessionObjectsInfo
					}
				}

				o, err := types.K8sObjectToProto(obj, n, tenant, inventory, info)
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
		SearchedNamespaces: GetClusterUserNamespacesNames(clusterUserNamespaces),
	}, nil
}


func parseSessionInfo(unstructuredObj unstructured.Unstructured) (string, string, error) {
	var set v1.StatefulSet

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(), &set)
	if err != nil {
		return "", "", fmt.Errorf("converting unstructured to statefulset: %w", err)
	}

	labels := set.GetLabels()

	if labels[types.AppLabel] != "vcluster" || labels[types.PartOfLabel] != "gitops-run" {
		return "", "", fmt.Errorf("unexpected format of labels")
	}

	annotations := set.GetAnnotations()

	var kind string
	if annotations["run.weave.works/automation-kind"] == "ks" {
		kind = kustomizev1.KustomizationKind
	} else {
		kind = v2beta1.HelmReleaseKind
	}

	ns := annotations["run.weave.works/namespace"]
	if ns == "" {
		return "", "", fmt.Errorf("empty session namespace")
	}

	clusterName := ns + "/" + set.GetName()

	return clusterName, kind, nil
}

func (cs *coreServer) sessionObjectsCreated(ctx context.Context, clusterName, objectNamespace, automationKind string) (bool, error) {
	automationName := constants.RunDevHelmName

	if automationKind == kustomizev1.KustomizationKind {
		automationName = constants.RunDevKsName
	}

	automation, err := cs.GetObject(ctx, &pb.GetObjectRequest{
		Name:        automationName,
		Namespace:   objectNamespace,
		Kind:        automationKind,
		ClusterName: clusterName,
	})

	if err != nil {
		return false, err
	}

	src, err := cs.GetObject(ctx, &pb.GetObjectRequest{
		Name:        constants.RunDevBucketName,
		Namespace:   objectNamespace,
		Kind:        "Bucket",
		ClusterName: clusterName,
	})

	if err != nil {
		return false, err
	}

	return automation != nil && src != nil, nil
}

func (cs *coreServer) GetObject(ctx context.Context, msg *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	gvk, err := cs.primaryKinds.Lookup(msg.Kind)
	if err != nil {
		return nil, err
	}

	unstructuredObj := unstructured.Unstructured{}
	unstructuredObj.SetGroupVersionKind(*gvk)

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := clustersClient.Get(ctx, msg.ClusterName, key, &unstructuredObj); err != nil {
		return nil, err
	}

	var inventory []*pb.GroupVersionKind = nil

	var obj client.Object = &unstructuredObj

	switch gvk.Kind {
	case "Secret":
		obj, err = sanitizeSecret(&unstructuredObj)
		if err != nil {
			return nil, fmt.Errorf("error sanitizing secrets: %w", err)
		}
	case v2beta1.HelmReleaseKind:
		inventory, err = getUnstructuredHelmReleaseInventory(ctx, unstructuredObj, clustersClient, msg.ClusterName)
		if err != nil {
			inventory = nil // We can still display most things without inventory

			cs.logger.V(logger.LogLevelDebug).Info("Couldn't grab inventory for helm release", "error", err)
		}
	}

	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))

	tenant := GetTenant(obj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

	res, err := types.K8sObjectToProto(obj, msg.ClusterName, tenant, inventory, "")

	if err != nil {
		return nil, fmt.Errorf("converting object to proto: %w", err)
	}

	return &pb.GetObjectResponse{Object: res}, nil
}
