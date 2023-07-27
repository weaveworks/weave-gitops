package fluxsync

import (
	"errors"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	imgautomationv1 "github.com/fluxcd/image-automation-controller/api/v1beta1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
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

// Automation objects are Kustomizations and HelmReleases.
// These are the only object types that can be triggered
// to be reconciled with their source.
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
	case *reflectorv1.ImageRepository:
		return ImageRepositoryAdapter{ImageRepository: o}
	case *imgautomationv1.ImageUpdateAutomation:
		return ImageUpdateAutomationAdapter{ImageUpdateAutomation: o}
	}
	return nil
}

type GitRepositoryAdapter struct {
	*sourcev1.GitRepository
}

func (obj GitRepositoryAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj GitRepositoryAdapter) AsClientObject() client.Object {
	return obj.GitRepository
}

func (obj GitRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
}

func (obj GitRepositoryAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj GitRepositoryAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
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

func (obj BucketAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.BucketKind)
}

func (obj BucketAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj BucketAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
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

func (obj HelmChartAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.HelmChartKind)
}

func (obj HelmChartAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj HelmChartAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
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

func (obj HelmRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.HelmRepositoryKind)
}

func (obj HelmRepositoryAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj HelmRepositoryAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
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

func (obj OCIRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return sourcev1.GroupVersion.WithKind(sourcev1.OCIRepositoryKind)
}

func (obj OCIRepositoryAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj OCIRepositoryAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
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

func (obj HelmReleaseAdapter) SourceRef() SourceRef {
	sr := obj.Spec.Chart.Spec.SourceRef

	return sRef{
		apiVersion: sr.APIVersion,
		name:       sr.Name,
		namespace:  sr.Namespace,
		kind:       sr.Kind,
	}
}

func (obj HelmReleaseAdapter) GroupVersionKind() schema.GroupVersionKind {
	return helmv2.GroupVersion.WithKind(helmv2.HelmReleaseKind)
}

func (obj HelmReleaseAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj HelmReleaseAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
}

type KustomizationAdapter struct {
	*kustomizev1.Kustomization
}

func (obj KustomizationAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj KustomizationAdapter) AsClientObject() client.Object {
	return obj.Kustomization
}

func (obj KustomizationAdapter) SourceRef() SourceRef {
	return sRef{
		apiVersion: obj.Spec.SourceRef.APIVersion,
		name:       obj.Spec.SourceRef.Name,
		namespace:  obj.Spec.SourceRef.Namespace,
		kind:       obj.Spec.SourceRef.Kind,
	}
}

func (obj KustomizationAdapter) GroupVersionKind() schema.GroupVersionKind {
	return kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
}

func (obj KustomizationAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj KustomizationAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
}

type ImageRepositoryAdapter struct {
	*reflectorv1.ImageRepository
}

func (obj ImageRepositoryAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj ImageRepositoryAdapter) AsClientObject() client.Object {
	return obj.ImageRepository
}

func (obj ImageRepositoryAdapter) GroupVersionKind() schema.GroupVersionKind {
	return reflectorv1.GroupVersion.WithKind(reflectorv1.ImageRepositoryKind)
}

func (obj ImageRepositoryAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj ImageRepositoryAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
}

type ImageUpdateAutomationAdapter struct {
	*imgautomationv1.ImageUpdateAutomation
}

func (obj ImageUpdateAutomationAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj ImageUpdateAutomationAdapter) AsClientObject() client.Object {
	return obj.ImageUpdateAutomation
}

func (obj ImageUpdateAutomationAdapter) GroupVersionKind() schema.GroupVersionKind {
	return imgautomationv1.GroupVersion.WithKind(imgautomationv1.ImageUpdateAutomationKind)
}

func (obj ImageUpdateAutomationAdapter) SetSuspended(suspend bool) {
	obj.Spec.Suspend = suspend
}

func (obj ImageUpdateAutomationAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
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

func ToReconcileable(kind string) (client.ObjectList, Reconcilable, error) {
	switch kind {
	case kustomizev1.KustomizationKind:
		return &kustomizev1.KustomizationList{}, NewReconcileable(&kustomizev1.Kustomization{}), nil

	case helmv2.HelmReleaseKind:
		return &helmv2.HelmReleaseList{}, NewReconcileable(&helmv2.HelmRelease{}), nil

	case sourcev1.GitRepositoryKind:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.GitRepository{}), nil

	case sourcev1.BucketKind:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.Bucket{}), nil

	case sourcev1.HelmRepositoryKind:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.HelmRepository{}), nil

	case sourcev1.HelmChartKind:
		return &sourcev1.GitRepositoryList{}, NewReconcileable(&sourcev1.HelmChart{}), nil

	case sourcev1.OCIRepositoryKind:
		return &sourcev1.OCIRepositoryList{}, NewReconcileable(&sourcev1.OCIRepository{}), nil

	case reflectorv1.ImageRepositoryKind:
		return &reflectorv1.ImageRepositoryList{}, NewReconcileable(&reflectorv1.ImageRepository{}), nil

	case imgautomationv1.ImageUpdateAutomationKind:
		return &imgautomationv1.ImageUpdateAutomationList{}, NewReconcileable(&imgautomationv1.ImageUpdateAutomation{}), nil
	}

	return nil, nil, errors.New("could not find source type")
}
