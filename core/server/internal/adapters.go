package internal

import (
	"errors"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcilable represents a Kubernetes object that Flux can reconcile
type Reconcilable interface {
	client.Object
	meta.ObjectWithConditions
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
	GetLastHandledReconcileRequest() string
	AsClientObject() client.Object
	GroupVersionKind() schema.GroupVersionKind
}

type SourceRef interface {
	APIVersion() string
	Kind() string
	Name() string
	Namespace() string
}

type Automation interface {
	Reconcilable
	SourceRef() SourceRef
}

func NewReconcileableSource(obj client.Object) Reconcilable {
	switch o := obj.(type) {
	case *sourcev1.GitRepository:
		return GitRepositoryAdapter{GitRepository: o}
	case *sourcev1.HelmRepository:
		return HelmRepositoryAdapter{HelmRepository: o}
	case *sourcev1.Bucket:
		return BucketAdapter{Bucket: o}
	case *sourcev1.HelmChart:
		return HelmChartAdapter{HelmChart: o}
	}

	return nil
}

type GitRepositoryAdapter struct {
	*sourcev1.GitRepository
}

func (o GitRepositoryAdapter) GetLastHandledReconcileRequest() string {
	return o.Status.GetLastHandledReconcileRequest()
}

func (o GitRepositoryAdapter) AsClientObject() client.Object {
	return o.GitRepository
}

func (o GitRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
}

type BucketAdapter struct {
	*sourcev1.Bucket
}

func (obj BucketAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj BucketAdapter) AsClientObject() client.Object {
	return obj.Bucket
}

func (o BucketAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.BucketKind)
}

type HelmChartAdapter struct {
	*sourcev1.HelmChart
}

func (obj HelmChartAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj HelmChartAdapter) AsClientObject() client.Object {
	return obj.HelmChart
}

func (o HelmChartAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.HelmChartKind)
}

type HelmRepositoryAdapter struct {
	*sourcev1.HelmRepository
}

func (obj HelmRepositoryAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj HelmRepositoryAdapter) AsClientObject() client.Object {
	return obj.HelmRepository
}

func (o HelmRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.HelmChartKind)
}

type HelmReleaseAdapter struct {
	*helmv2.HelmRelease
}

func (obj HelmReleaseAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj HelmReleaseAdapter) AsClientObject() client.Object {
	return obj.HelmRelease
}

func (o HelmReleaseAdapter) SourceRef() SourceRef {
	sr := o.Spec.Chart.Spec.SourceRef

	return sRef{
		apiVersion: sr.APIVersion,
		name:       sr.Name,
		namespace:  sr.Namespace,
		kind:       sr.Kind,
	}
}

func (o HelmReleaseAdapter) GroupVersionKind() schema.GroupVersionKind {
	return helmv2.GroupVersion.WithKind(helmv2.HelmReleaseKind)
}

type KustomizationAdapter struct {
	*kustomizev1.Kustomization
}

func (o KustomizationAdapter) GetLastHandledReconcileRequest() string {
	return o.Status.GetLastHandledReconcileRequest()
}

func (o KustomizationAdapter) AsClientObject() client.Object {
	return o.Kustomization
}

func (o KustomizationAdapter) SourceRef() SourceRef {
	return sRef{
		apiVersion: o.Spec.SourceRef.APIVersion,
		name:       o.Spec.SourceRef.Name,
		namespace:  o.Spec.SourceRef.Namespace,
		kind:       o.Spec.SourceRef.Kind,
	}
}

func (o KustomizationAdapter) GroupVersionKind() schema.GroupVersionKind {
	return kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
}

type sRef struct {
	apiVersion string
	name       string
	kind       string
	namespace  string
}

func (s sRef) APIVersion() string {
	return s.apiVersion
}

func (s sRef) Name() string {
	return s.name
}

func (s sRef) Namespace() string {
	return s.namespace
}

func (s sRef) Kind() string {
	return s.kind
}

func ToReconcileableSource(sourceType pb.SourceRef_SourceKind) (client.ObjectList, Reconcilable, error) {
	switch sourceType {
	case pb.SourceRef_GitRepository:
		return &sourcev1.GitRepositoryList{}, NewReconcileableSource(&sourcev1.GitRepository{}), nil

	case pb.SourceRef_Bucket:
		return &sourcev1.GitRepositoryList{}, NewReconcileableSource(&sourcev1.Bucket{}), nil

	case pb.SourceRef_HelmRepository:
		return &sourcev1.GitRepositoryList{}, NewReconcileableSource(&sourcev1.HelmRepository{}), nil

	case pb.SourceRef_HelmChart:
		return &sourcev1.GitRepositoryList{}, NewReconcileableSource(&sourcev1.HelmChart{}), nil
	}

	return nil, nil, errors.New("could not find source type")
}
