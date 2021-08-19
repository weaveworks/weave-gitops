package kube

import (
	"context"
	"encoding/json"
	"fmt"

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
	"github.com/weaveworks/weave-gitops/pkg/utils"
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
	GVRHelmRepository schema.GroupVersionResource = helmv2.GroupVersion.WithResource("helmrepositories")
	GVRHelmRelease    schema.GroupVersionResource = helmv2.GroupVersion.WithResource("helmreleases")
)

func NewKubeHTTPClient() (Kube, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	_, kubeContext, err := initialContexts(cfgLoadingRules)
	if err != nil {
		return nil, fmt.Errorf("could not get initial context: %w", err)
	}

	configOverrides := clientcmd.ConfigOverrides{CurrentContext: kubeContext}

	restCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		cfgLoadingRules,
		&configOverrides,
	).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("could not create rest config: %w", err)
	}

	scheme := CreateScheme()

	kubeClient, err := client.New(restCfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("kubernetes client initialization failed: %w", err)
	}

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discovery client: %s", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dynamic client: %s", err)
	}

	return &KubeHTTP{Client: kubeClient, ClusterName: kubeContext, restMapper: mapper, dynClient: dyn}, nil
}

// This is an alternative implementation of the kube.Kube interface,
// specifically designed to query the K8s API directly instead of relying on
// `kubectl` to be present in the PATH.
type KubeHTTP struct {
	Client      client.Client
	ClusterName string
	dynClient   dynamic.Interface
	restMapper  *restmapper.DeferredDiscoveryRESTMapper
}

func (c *KubeHTTP) GetClusterName(ctx context.Context) (string, error) {
	return c.ClusterName, nil
}

func (c *KubeHTTP) GetClusterStatus(ctx context.Context) ClusterStatus {
	tName := types.NamespacedName{
		Name: WeGOCRDName,
	}

	crd := v1.CustomResourceDefinition{}

	if c.Client.Get(ctx, tName, &crd) == nil {
		return WeGOInstalled
	}

	if ok, _ := c.FluxPresent(ctx); ok {
		return FluxInstalled
	}

	dep := appsv1.Deployment{}
	coreDnsName := types.NamespacedName{
		Name:      "coredns",
		Namespace: "kube-system",
	}

	if err := c.Client.Get(ctx, coreDnsName, &dep); err != nil {
		// Couldn't find the coredns deployment.
		// We don't know what state the cluster is in.
		return Unknown
	} else {
		// Request for the coredns namespace was successfull.
		return Unmodified
	}
}

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func (c *KubeHTTP) Apply(ctx context.Context, manifest []byte, namespace string) error {
	dr, name, data, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
	}

	_, err = dr.Patch(ctx, name, types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "wego",
	})
	if err != nil {
		return fmt.Errorf("failed applying %s: %w", string(data), err)
	}

	return nil
}

func (c *KubeHTTP) getResourceInterface(manifest []byte) (dynamic.ResourceInterface, string, []byte, error) {
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(manifest), nil, obj)
	if err != nil {
		return nil, "", nil, err
	}

	// 4. Find GVR
	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, "", nil, err
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = c.dynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = c.dynClient.Resource(mapping.Resource)
	}

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, "", nil, err
	}

	return dr, obj.GetName(), data, nil
}

func (c *KubeHTTP) GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
	app := wego.Application{}
	if err := c.Client.Get(ctx, name, &app); err != nil {
		return nil, fmt.Errorf("could not get application: %w", err)
	}

	return &app, nil
}

func (c *KubeHTTP) Delete(ctx context.Context, manifest []byte, namespace string) error {
	dr, name, data, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
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

	if err := c.dynClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, deleteOptions); err != nil {
		return fmt.Errorf("failed to delete resource name=%s resource-type=%#v namespace=%s error=%w", name, gvr, namespace, err)
	}

	return nil
}

func (c *KubeHTTP) FluxPresent(ctx context.Context) (bool, error) {
	key := types.NamespacedName{
		Name: FluxNamespace,
	}

	ns := corev1.Namespace{}

	if err := c.Client.Get(ctx, key, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("could not find flux namespace: %w", err)
	}

	return true, nil
}

func (c *KubeHTTP) SecretPresent(ctx context.Context, secretName string, namespace string) (bool, error) {
	name := types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}

	secret := corev1.Secret{}
	if err := c.Client.Get(ctx, name, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("could not get secret: %w", err)
	}

	return true, nil
}

func (c *KubeHTTP) GetApplications(ctx context.Context, namespace string) ([]wego.Application, error) {
	result := wego.ApplicationList{}

	if err := c.Client.List(ctx, &result, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("could not list wego applications: %w", err)
	}

	return result.Items, nil
}

func (c *KubeHTTP) AppExistsInCluster(ctx context.Context, namespace string, appHash string) error {
	apps, err := c.GetApplications(ctx, namespace)
	if err != nil {
		return err
	}

	for _, app := range apps {
		existingHash, err := utils.GetAppHash(app)
		if err != nil {
			return err
		}

		if appHash == existingHash {
			return fmt.Errorf("unable to create resource, resource already exists in cluster")
		}
	}

	return nil
}

func (c *KubeHTTP) GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error {
	if err := c.Client.Get(ctx, name, resource); err != nil {
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
