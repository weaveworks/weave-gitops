package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/logger"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	FluxNamespacePartOf = "flux"
)

var (
	KustomizeNameKey      = fmt.Sprintf("%s/name", kustomizev1.GroupVersion.Group)
	KustomizeNamespaceKey = fmt.Sprintf("%s/namespace", kustomizev1.GroupVersion.Group)
	HelmNameKey           = fmt.Sprintf("%s/name", helmv2.GroupVersion.Group)
	HelmNamespaceKey      = fmt.Sprintf("%s/namespace", helmv2.GroupVersion.Group)

	// ErrFluxNamespaceNotFound no flux namespace found
	ErrFluxNamespaceNotFound = errors.New("could not find flux namespace in cluster")
	// ErrListingDeployments no deployments found
	ErrListingDeployments = errors.New("could not list deployments in namespace")

	DefaultFluxNamespace = lookupEnv("WEAVE_GITOPS_FALLBACK_NAMESPACE", "flux-system")
)

func lookupEnv(envVar, fallback string) string {
	if val, ok := os.LookupEnv(envVar); ok {
		return val
	}

	return fallback
}

func (cs *coreServer) ListFluxRuntimeObjects(ctx context.Context, msg *pb.ListFluxRuntimeObjectsRequest) (*pb.ListFluxRuntimeObjectsResponse, error) {
	respErrors := []*pb.ListError{}

	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	var results []*pb.Deployment

	for clusterName, nss := range cs.clustersManager.GetClustersNamespaces() {
		fluxNamepsaces := filterFluxNamespace(nss)
		if len(fluxNamepsaces) == 0 {
			respErrors = append(respErrors, &pb.ListError{ClusterName: clusterName, Namespace: "", Message: ErrFluxNamespaceNotFound.Error()})
			continue
		}

		opts := client.MatchingLabels{
			coretypes.PartOfLabel: FluxNamespacePartOf,
		}

		list := &appsv1.DeploymentList{}

		for _, fluxNs := range fluxNamepsaces {
			if err := clustersClient.List(ctx, clusterName, list, opts, client.InNamespace(fluxNs.Name)); err != nil {
				respErrors = append(respErrors, &pb.ListError{ClusterName: clusterName, Namespace: fluxNs.Name, Message: fmt.Sprintf("%s, %s", ErrListingDeployments.Error(), err)})
				continue
			}

			for _, d := range list.Items {
				r := &pb.Deployment{
					Name:        d.Name,
					Namespace:   d.Namespace,
					Conditions:  []*pb.Condition{},
					ClusterName: clusterName,
					Uid:         string(d.GetUID()),
					Labels:      d.Labels,
				}

				for _, cond := range d.Status.Conditions {
					r.Conditions = append(r.Conditions, &pb.Condition{
						Message: cond.Message,
						Reason:  cond.Reason,
						Status:  string(cond.Status),
						Type:    string(cond.Type),
					})
				}

				for _, img := range d.Spec.Template.Spec.Containers {
					r.Images = append(r.Images, img.Image)
				}

				results = append(results, r)
			}
		}
	}

	return &pb.ListFluxRuntimeObjectsResponse{Deployments: results, Errors: respErrors}, nil
}

func (cs *coreServer) ListFluxCrds(ctx context.Context, msg *pb.ListFluxCrdsRequest) (*pb.ListFluxCrdsResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &apiextensions.CustomResourceDefinitionList{}
	})

	respErrors := []*pb.ListError{}

	opts := client.MatchingLabels{
		coretypes.PartOfLabel: FluxNamespacePartOf,
	}

	if err := clustersClient.ClusteredList(ctx, clist, false, opts); err != nil {
		var errs clustersmngr.ClusteredListError

		if !errors.As(err, &errs) {
			return nil, fmt.Errorf("CRDs clustered list: %w", errs)
		}

		for _, e := range errs.Errors {
			respErrors = append(respErrors, &pb.ListError{
				ClusterName: e.Cluster,
				Message:     e.Err.Error(),
			})
		}
	}

	results := []*pb.Crd{}

	for clusterName, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*apiextensions.CustomResourceDefinitionList)
			if !ok {
				continue
			}

			for _, d := range list.Items {
				version := ""

				for _, v := range d.Spec.Versions {
					// This is the "active" version of the CRD in etcd, and a CRD
					// can only have one version marked as such.
					if v.Storage {
						version = v.Name
						break
					}
				}

				r := &pb.Crd{
					Name: &pb.Crd_Name{
						Plural: d.Spec.Names.Plural,
						Group:  d.Spec.Group,
					},
					Version:     version,
					Kind:        d.Spec.Names.Kind,
					ClusterName: clusterName,
					Uid:         string(d.GetUID()),
				}
				results = append(results, r)
			}
		}
	}

	return &pb.ListFluxCrdsResponse{Crds: results, Errors: respErrors}, nil
}

func filterFluxNamespace(nss []v1.Namespace) []v1.Namespace {
	fluxSystem := []v1.Namespace{}

	for _, ns := range nss {
		if val, ok := ns.Labels[coretypes.PartOfLabel]; ok && val == FluxNamespacePartOf {
			fluxSystem = append(fluxSystem, ns)
			continue
		}

		if ns.Name == DefaultFluxNamespace {
			fluxSystem = append(fluxSystem, ns)
		}
	}

	return fluxSystem
}

func (cs *coreServer) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsRequest) (*pb.GetReconciledObjectsResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var opts client.MatchingLabels

	switch msg.AutomationKind {
	case kustomizev1.KustomizationKind:
		opts = client.MatchingLabels{
			KustomizeNameKey:      msg.AutomationName,
			KustomizeNamespaceKey: msg.Namespace,
		}
	case helmv2.HelmReleaseKind:
		opts = client.MatchingLabels{
			HelmNameKey:      msg.AutomationName,
			HelmNamespaceKey: msg.Namespace,
		}
	default:
		return nil, fmt.Errorf("unsupported application kind: %s", msg.AutomationKind)
	}

	var (
		result   = []unstructured.Unstructured{}
		checkDup = map[types.UID]bool{}
		resultMu = sync.Mutex{}

		respErrors = multierror.Error{}
		errsMu     = sync.Mutex{}

		wg = sync.WaitGroup{}
	)

	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))

	kinds := getUniqueKinds(msg.Kinds)

	for _, namespaces := range clusterUserNamespaces {
		for _, ns := range namespaces {
			for _, gvk := range kinds {
				wg.Add(1)
				go func(namespace, clusterName string, gvk *pb.GroupVersionKind) {
					defer wg.Done()

					listResult := unstructured.UnstructuredList{}

					listResult.SetGroupVersionKind(schema.GroupVersionKind{
						Group:   gvk.Group,
						Kind:    gvk.Kind,
						Version: gvk.Version,
					})

					if err := clustersClient.List(ctx, clusterName, &listResult, opts, client.InNamespace(namespace)); err != nil {
						if k8serrors.IsForbidden(err) {
							cs.logger.V(logger.LogLevelDebug).Info(
								"forbidden list request",
								"cluster", msg.ClusterName,
								"automation", msg.AutomationName,
								"namespace", msg.Namespace,
								"gvk", gvk.String(),
							)
							// Our service account (or impersonated user) may not have the ability to see the resource in question,
							// in the given namespace. We pretend it doesn't exist and keep looping.
							// We need logging to make this error more visible.
							return
						}

						if k8serrors.IsTimeout(err) {
							cs.logger.Error(err, "List timedout", "gvk", gvk.String())

							return
						}

						errsMu.Lock()
						respErrors = *multierror.Append(fmt.Errorf("listing unstructured object: %w", err), respErrors.Errors...)
						errsMu.Unlock()

						cs.logger.Error(err, "failed listing resources", "namespace", namespace, "gvk", gvk.String())
					}

					for _, u := range listResult.Items {
						uid := u.GetUID()
						resultMu.Lock()
						if !checkDup[uid] {
							result = append(result, u)
							checkDup[uid] = true
						}
						resultMu.Unlock()
					}
				}(ns.Name, msg.ClusterName, gvk)
			}
		}
	}

	wg.Wait()

	objects := []*pb.Object{}

	for _, unstructuredObj := range result {
		tenant := GetTenant(unstructuredObj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

		var o *pb.Object

		var obj client.Object = &unstructuredObj

		if unstructuredObj.GetKind() == "Secret" {
			obj, err = sanitizeSecret(&unstructuredObj)
			if err != nil {
				respErrors = *multierror.Append(fmt.Errorf("error sanitizing secrets: %w", err), respErrors.Errors...)
				continue
			}
		}

		o, err = coretypes.K8sObjectToProto(obj, msg.ClusterName, tenant, nil, "")
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting objects: %w", err), respErrors.Errors...)
			continue
		}

		objects = append(objects, o)
	}

	return &pb.GetReconciledObjectsResponse{Objects: objects}, respErrors.ErrorOrNil()
}

func (cs *coreServer) GetChildObjects(ctx context.Context, msg *pb.GetChildObjectsRequest) (*pb.GetChildObjectsResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	opts := client.InNamespace(msg.Namespace)

	listResult := unstructured.UnstructuredList{}

	listResult.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   msg.GroupVersionKind.Group,
		Version: msg.GroupVersionKind.Version,
		Kind:    msg.GroupVersionKind.Kind,
	})

	if err := clustersClient.List(ctx, msg.ClusterName, &listResult, opts); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s", err)
	}

	respErrors := multierror.Error{}
	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))
	objects := []*pb.Object{}

ItemsLoop:
	for _, obj := range listResult.Items {
		refs := obj.GetOwnerReferences()
		if len(refs) == 0 {
			// Ignore items without OwnerReference.
			// for example: dev-weave-gitops-test-connection
			continue ItemsLoop
		}

		for _, ref := range refs {
			if ref.UID != types.UID(msg.ParentUid) {
				// Assuming all owner references have the same parent UID,
				// this is not the child we are looking for.
				// Skip the rest of the operations in Items loops.
				continue ItemsLoop
			}
		}

		tenant := GetTenant(obj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

		obj, err := coretypes.K8sObjectToProto(&obj, msg.ClusterName, tenant, nil, "")

		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting objects: %w", err), respErrors.Errors...)
			continue
		}
		objects = append(objects, obj)
	}

	return &pb.GetChildObjectsResponse{Objects: objects}, nil
}

func sanitizeSecret(obj *unstructured.Unstructured) (client.Object, error) {
	bytes, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshaling secret: %v", err)
	}

	s := &v1.Secret{}

	if err := json.Unmarshal(bytes, s); err != nil {
		return nil, fmt.Errorf("unmarshaling secret: %v", err)
	}

	s.Data = map[string][]byte{"redacted": []byte(nil)}

	return s, nil
}

func getUniqueKinds(gvks []*pb.GroupVersionKind) []*pb.GroupVersionKind {
	uniqueMap := map[string]*pb.GroupVersionKind{}
	for _, gvk := range gvks {
		uniqueMap[gvk.String()] = gvk
	}

	unique := []*pb.GroupVersionKind{}

	for _, gvk := range uniqueMap {
		unique = append(unique, gvk)
	}

	return unique
}
