package kube

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/rest"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func CreateScheme() *apiruntime.Scheme {
	scheme := apiruntime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = kustomizev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	_ = wego.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = extensionsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	return scheme
}

const WeGONamespace = "wego-system"
const WeGOCRDName = "apps.wego.weave.works"
const FluxNamespace = "flux-system"

var (
	GVRSecret         schema.GroupVersionResource = corev1.SchemeGroupVersion.WithResource("secrets")
	GVRApp            schema.GroupVersionResource = wego.GroupVersion.WithResource("apps")
	GVRKustomization  schema.GroupVersionResource = kustomizev1.GroupVersion.WithResource("kustomizations")
	GVRGitRepository  schema.GroupVersionResource = sourcev1.GroupVersion.WithResource("gitrepositories")
	GVRHelmRepository schema.GroupVersionResource = sourcev1.GroupVersion.WithResource("helmrepositories")
	GVRHelmRelease    schema.GroupVersionResource = helmv2.GroupVersion.WithResource("helmreleases")
)

func NewKubeHTTPClient() (Kube, client.Client, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	_, kubeContext, err := initialContexts(cfgLoadingRules)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get initial context: %w", err)
	}

	config, err := rest.InClusterConfig()
	if err == rest.ErrNotInCluster {
		configOverrides := clientcmd.ConfigOverrides{CurrentContext: kubeContext}

		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			cfgLoadingRules,
			&configOverrides,
		).ClientConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("could not create rest config: %w", err)
		}
	}

	scheme := CreateScheme()

	rawClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("kubernetes client initialization failed: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize discovery client: %s", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize dynamic client: %s", err)
	}

	return &KubeHTTP{Client: rawClient, ClusterName: kubeContext, RestMapper: mapper, DynClient: dyn}, rawClient, nil
}

// This is an alternative implementation of the kube.Kube interface,
// specifically designed to query the K8s API directly instead of relying on
// `kubectl` to be present in the PATH.
type KubeHTTP struct {
	Client      client.Client
	ClusterName string
	DynClient   dynamic.Interface
	RestMapper  meta.RESTMapper
}

func (k *KubeHTTP) GetClusterName(ctx context.Context) (string, error) {
	return k.ClusterName, nil
}

func (k *KubeHTTP) GetClusterStatus(ctx context.Context) ClusterStatus {
	tName := types.NamespacedName{
		Name: WeGOCRDName,
	}

	crd := v1.CustomResourceDefinition{}

	if k.Client.Get(ctx, tName, &crd) == nil {
		return GitOpsInstalled
	}

	if ok, _ := k.FluxPresent(ctx); ok {
		return FluxInstalled
	}

	dep := appsv1.Deployment{}
	coreDnsName := types.NamespacedName{
		Name:      "coredns",
		Namespace: "kube-system",
	}

	if err := k.Client.Get(ctx, coreDnsName, &dep); err != nil {
		// Some clusters don't have 'coredns'; if we get a "not found" error, we know we
		// can talk to the cluster
		if apierrors.IsNotFound(err) {
			return Unmodified
		}

		return Unknown
	} else {
		// Request for the coredns namespace was successful.
		return Unmodified
	}
}

func (k *KubeHTTP) Apply(ctx context.Context, manifest []byte, namespace string) error {
	dr, name, data, err := k.getResourceInterface(manifest, namespace)
	if err != nil {
		return fmt.Errorf("failed to dynamic resource interface: %w", err)
	}

	_, err = dr.Patch(ctx, name, types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "wego",
	})
	if err != nil {
		return fmt.Errorf("failed applying %s: %w", string(data), err)
	}

	return nil
}

func (k *KubeHTTP) getResourceInterface(manifest []byte, namespace string) (dynamic.ResourceInterface, string, []byte, error) {
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}

	_, gvk, err := decUnstructured.Decode(manifest, nil, obj)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed decoding manifest: %w", err)
	}

	mapping, err := k.RestMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed getting rest mapping: %w", err)
	}

	var dr dynamic.ResourceInterface

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if namespace == "" {
			namespace = obj.GetNamespace()
		}

		dr = k.DynClient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		dr = k.DynClient.Resource(mapping.Resource)
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed marshalling resource: %w", err)
	}

	return dr, obj.GetName(), data, nil
}

func (k *KubeHTTP) GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
	app := wego.Application{}
	if err := k.Client.Get(ctx, name, &app); err != nil {
		return nil, fmt.Errorf("could not get application: %w", err)
	}

	return &app, nil
}

func (k *KubeHTTP) Delete(ctx context.Context, manifest []byte) error {
	dr, name, data, err := k.getResourceInterface(manifest, "")
	if err != nil {
		return fmt.Errorf("failed to dynamic resource interface: %w", err)
	}

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err = dr.Delete(ctx, name, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed applying %s: %w", string(data), err)
	}

	return nil
}

func (c *KubeHTTP) DeleteByName(ctx context.Context, name string, gvr schema.GroupVersionResource, namespace string) error {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := c.DynClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, deleteOptions); err != nil {
		return fmt.Errorf("failed to delete resource name=%s resource-type=%#v namespace=%s error=%w", name, gvr, namespace, err)
	}

	return nil
}

func (k *KubeHTTP) FluxPresent(ctx context.Context) (bool, error) {
	key := types.NamespacedName{
		Name: FluxNamespace,
	}

	ns := corev1.Namespace{}

	if err := k.Client.Get(ctx, key, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("could not find flux namespace: %w", err)
	}

	return true, nil
}

func (k *KubeHTTP) SecretPresent(ctx context.Context, secretName string, namespace string) (bool, error) {
	name := types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}

	if _, err := k.GetSecret(ctx, name); err != nil {
		return false, fmt.Errorf("error checking secret presence for secret \"%s\": %w", secretName, err)
	}

	// No error, it must exist
	return true, nil
}

func (k KubeHTTP) GetSecret(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	if err := k.Client.Get(ctx, name, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("could not get secret: %w", err)
	}

	return &secret, nil
}

func (k *KubeHTTP) GetApplications(ctx context.Context, namespace string) ([]wego.Application, error) {
	result := wego.ApplicationList{}

	if err := k.Client.List(ctx, &result, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("could not list wego applications: %w", err)
	}

	return result.Items, nil
}

func (k *KubeHTTP) GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error {
	if err := k.Client.Get(ctx, name, resource); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error getting resource: %w", err)
	}

	return nil
}

func initialContexts(cfgLoadingRules *clientcmd.ClientConfigLoadingRules) (contexts []string, currentCtx string, err error) {
	rules, err := cfgLoadingRules.Load()

	if err != nil {
		return contexts, currentCtx, err
	}

	for _, c := range rules.Contexts {
		contexts = append(contexts, c.Cluster)
	}

	return contexts, rules.CurrentContext, nil
}
