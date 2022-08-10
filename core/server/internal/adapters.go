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
	SetSuspended(suspend bool)
	DeepCopyClientObject() client.Object
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

func NewReconcileable(obj client.Object) Reconcilable {
	switch o := obj.(type) {
	case *kustomizev1.Kustomization:
		return KustomizationAdapter{Kustomization: o}
	case *helmv2.HelmRelease:
		return HelmReleaseAdapter{HelmRelease: o}
	case *sourcev1.GitRepository:
		return GitRepositoryAdapter{GitRepository: o}
	case *sourcev1.HelmRepository:
		return HelmRepositoryAdapter{HelmRepository: o}
	case *sourcev1.Bucket:
		return BucketAdapter{Bucket: o}
	case *sourcev1.HelmChart:
		return HelmChartAdapter{HelmChart: o}
	case *sourcev1.OCIRepository:
		return OCIRepositoryAdapter{OCIRepository: o}
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

func (o GitRepositoryAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o GitRepositoryAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
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

func (o BucketAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o BucketAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
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

func (o HelmChartAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o HelmChartAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
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
	return sourcev1.GroupVersion.WithKind(sourcev1.HelmRepositoryKind)
}

func (o HelmRepositoryAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o HelmRepositoryAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
}

type OCIRepositoryAdapter struct {
	*sourcev1.OCIRepository
}

func (obj OCIRepositoryAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj OCIRepositoryAdapter) AsClientObject() client.Object {
	return obj.OCIRepository
}

func (o OCIRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.OCIRepositoryKind)
}

func (o OCIRepositoryAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o OCIRepositoryAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
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

func (o HelmReleaseAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o HelmReleaseAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
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

func (o KustomizationAdapter) SetSuspended(suspend bool) {
	o.Spec.Suspend = suspend
}

func (o KustomizationAdapter) DeepCopyClientObject() client.Object {
	return o.DeepCopy()
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

func ToReconcileable(kind pb.FluxObjectKind) (client.ObjectList, Reconcilable, error) {
	switch kind {
	case pb.FluxObjectKind_KindKustomization:
		return &kustomizev1.KustomizationList{}, NewReconcileable(&kustomizev1.Kustomization{}), nil

	case pb.FluxObjectKind_KindHelmRelease:
		return &helmv2.HelmReleaseList{}, NewReconcileable(&helmv2.HelmRelease{}), nil

	case pb.FluxObjectKind_KindGitRepository:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.GitRepository{}), nil

	case pb.FluxObjectKind_KindBucket:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.Bucket{}), nil

	case pb.FluxObjectKind_KindHelmRepository:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.HelmRepository{}), nil

	case pb.FluxObjectKind_KindHelmChart:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.HelmChart{}), nil

	case pb.FluxObjectKind_KindOCIRepository:
		return &sourcev1.OCIRepositoryList{}, NewReconcileable(&sourcev1.OCIRepository{}), nil
	}

	return nil, nil, errors.New("could not find source type")
}
