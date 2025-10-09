package fluxsync

import (
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	imgautomationv1 "github.com/fluxcd/image-automation-controller/api/v1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	SetSuspended(suspend bool) error
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

func (obj GitRepositoryAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj BucketAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj HelmChartAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj HelmRepositoryAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj OCIRepositoryAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj HelmReleaseAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj KustomizationAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj ImageRepositoryAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
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

func (obj ImageUpdateAutomationAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
}

func (obj ImageUpdateAutomationAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
}

// UnstructuredAdapter implements the Reconcilable interface for unstructured resources.
// The underlying resource gvk should have the standard flux object sync/suspend fields
type UnstructuredAdapter struct {
	*unstructured.Unstructured
}

func (obj UnstructuredAdapter) GetLastHandledReconcileRequest() string {
	if val, found, _ := unstructured.NestedString(obj.Object, "status", "lastHandledReconcileAt"); found {
		return val
	}
	return ""
}

func (obj UnstructuredAdapter) GetConditions() []metav1.Condition {
	conditionsSlice, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found || err != nil {
		return nil
	}

	conditions := make([]metav1.Condition, len(conditionsSlice))
	for i, c := range conditionsSlice {
		var condition metav1.Condition
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(c.(map[string]interface{}), &condition); err != nil {
			continue
		}
		conditions[i] = condition
	}

	return conditions
}

func (obj UnstructuredAdapter) AsClientObject() client.Object {
	// Important for the controller-runtime type reflection to work
	// We can't return just `obj` here otherwise we get a
	// panic: reflect: call of reflect.Value.Elem on struct Value
	return obj.Unstructured
}

func (obj UnstructuredAdapter) SetSuspended(suspend bool) error {
	return unstructured.SetNestedField(obj.Object, suspend, "spec", "suspend")
}

func (obj UnstructuredAdapter) DeepCopyClientObject() client.Object {
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

// ToReconcileable takes a GVK and returns a "Reconcilable" for it.
// The reconcilable can be passed to a controller-runtime client to fetch it
// from the cluster. Once fetched we can query it for the last sync time, whether
// its suspended etc, using the Reconcilable interface.
//
// The generic unstructured case handles "flux like" objects that we don't explicitly
// know about, but which follow the same patterns for suspend/sync as a stadard flux object.
// E.g. `spec.suspend` and `status.lastHandledReconcileRequest` etc.
func ToReconcileable(gvk schema.GroupVersionKind) Reconcilable {
	switch gvk.Kind {
	case kustomizev1.KustomizationKind:
		return KustomizationAdapter{Kustomization: &kustomizev1.Kustomization{}}
	case helmv2.HelmReleaseKind:
		return HelmReleaseAdapter{HelmRelease: &helmv2.HelmRelease{}}
	// TODO: remove all these and let them fall through to the Unstructured case?
	case sourcev1.GitRepositoryKind:
		return GitRepositoryAdapter{GitRepository: &sourcev1.GitRepository{}}
	case sourcev1.BucketKind:
		return BucketAdapter{Bucket: &sourcev1.Bucket{}}
	case sourcev1.HelmRepositoryKind:
		return HelmRepositoryAdapter{HelmRepository: &sourcev1.HelmRepository{}}
	case sourcev1.HelmChartKind:
		return HelmChartAdapter{HelmChart: &sourcev1.HelmChart{}}
	case sourcev1.OCIRepositoryKind:
		return OCIRepositoryAdapter{OCIRepository: &sourcev1.OCIRepository{}}
	case reflectorv1.ImageRepositoryKind:
		return ImageRepositoryAdapter{ImageRepository: &reflectorv1.ImageRepository{}}
	case imgautomationv1.ImageUpdateAutomationKind:
		return ImageUpdateAutomationAdapter{ImageUpdateAutomation: &imgautomationv1.ImageUpdateAutomation{}}
	}

	// Return the UnstructuredAdapter for flux-like resources
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	return UnstructuredAdapter{Unstructured: obj}
}
