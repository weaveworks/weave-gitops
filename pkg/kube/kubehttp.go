package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/pkg/errors"
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
	kyaml "sigs.k8s.io/yaml"
)

func CreateScheme() *apiruntime.Scheme {
	scheme := apiruntime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = kustomizev2.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	_ = wego.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = extensionsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	return scheme
}

const WeGOCRDName = "apps.wego.weave.works"
const FluxNamespace = "flux-system"

var (
	GVRSecret         schema.GroupVersionResource = corev1.SchemeGroupVersion.WithResource("secrets")
	GVRApp            schema.GroupVersionResource = wego.GroupVersion.WithResource("apps")
	GVRKustomization  schema.GroupVersionResource = kustomizev2.GroupVersion.WithResource("kustomizations")
	GVRGitRepository  schema.GroupVersionResource = sourcev1.GroupVersion.WithResource("gitrepositories")
	GVRHelmRepository schema.GroupVersionResource = sourcev1.GroupVersion.WithResource("helmrepositories")
	GVRHelmRelease    schema.GroupVersionResource = helmv2.GroupVersion.WithResource("helmreleases")
)

const (
	WegoConfigMapName = "weave-gitops-config"
)

var (
	//ErrWegoConfigNotFound indicates weave gitops config could not be found
	ErrWegoConfigNotFound = errors.New("Wego Config not found")
)

// InClusterConfig defines a function for checking if this code is executing in kubernetes.
// This can be overriden if needed.
var InClusterConfig func() (*rest.Config, error) = func() (*rest.Config, error) {
	return rest.InClusterConfig()
}

var ErrNamespaceNotFound = errors.New("namespace not found")

func NewKubeHTTPClientWithConfig(config *rest.Config, contextName string) (Kube, client.Client, error) {
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

	return &KubeHTTP{Client: rawClient, ClusterName: contextName, RestMapper: mapper, DynClient: dyn}, rawClient, nil
}

func NewKubeHTTPClient() (Kube, client.Client, error) {
	config, contextName, err := RestConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not create default config: %w", err)
	}

	return NewKubeHTTPClientWithConfig(config, contextName)
}

func RestConfig() (*rest.Config, string, error) {
	config, err := InClusterConfig()
	if err != nil {
		if err == rest.ErrNotInCluster {
			return outOfClusterConfig()
		}
		// Handle other errors
		return nil, "", fmt.Errorf("could not create in-cluster config: %w", err)
	}

	return config, inClusterConfigClusterName(), nil
}

func inClusterConfigClusterName() string {
	// kube clusters don't really know their own names
	// try and read a unique name from the env, fall back to "default"
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "default"
	}

	return clusterName
}

func outOfClusterConfig() (*rest.Config, string, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeContext, clusterName, err := initialContext(cfgLoadingRules)
	if err != nil {
		return nil, "", fmt.Errorf("could not get initial context: %w", err)
	}

	configOverrides := clientcmd.ConfigOverrides{CurrentContext: kubeContext}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		cfgLoadingRules,
		&configOverrides,
	).ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("could not create rest config: %w", err)
	}

	return config, clusterName, nil
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

func (k *KubeHTTP) Raw() client.Client {
	return k.Client
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

	force := true
	_, err = dr.Patch(ctx, name, types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "wego",
		Force:        &force,
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
		return fmt.Errorf("failed deleting %s: %w", string(data), err)
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

func (k *KubeHTTP) NamespacePresent(ctx context.Context, namespace string) (bool, error) {
	key := types.NamespacedName{
		Name: namespace,
	}

	ns := corev1.Namespace{}

	if err := k.Client.Get(ctx, key, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("could not find namespace: %w", err)
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
		return nil, fmt.Errorf("could not list gitops applications: %w", err)
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

func (k *KubeHTTP) SetResource(ctx context.Context, resource Resource) error {
	if err := k.Client.Update(ctx, resource); err != nil {
		return fmt.Errorf("error setting resource: %w", err)
	}

	return nil
}

func (k *KubeHTTP) SetWegoConfig(ctx context.Context, config WegoConfig, namespace string) (*corev1.ConfigMap, error) {
	configBytes, err := kyaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	name := types.NamespacedName{Name: WegoConfigMapName, Namespace: namespace}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: map[string]string{
			"config": string(configBytes),
		},
	}

	orginalCm := cm.DeepCopy()

	if err := k.Client.Get(ctx, name, cm); err != nil {
		if apierrors.IsNotFound(err) {
			if err = k.Client.Create(ctx, cm); err != nil {
				return nil, fmt.Errorf("failed creating weave-gitops configmap: %w", err)
			}

			return orginalCm, nil
		}

		return nil, fmt.Errorf("failed getting weave-gitops configmap: %w", err)
	}

	cm.Data["config"] = string(configBytes)

	if err = k.Client.Update(ctx, cm); err != nil {
		return nil, fmt.Errorf("failed updating weave-gitops configmap: %w", err)
	}

	return orginalCm, nil
}

// GetWegoConfig fetches the wego config saved in the cluster in a given namespace.
// If an empty namespace is passed it will search in all namespaces and return the first one it finds.
func (k *KubeHTTP) GetWegoConfig(ctx context.Context, namespace string) (*WegoConfig, error) {
	name := types.NamespacedName{Name: WegoConfigMapName, Namespace: namespace}

	cm := &corev1.ConfigMap{}

	if namespace != "" {
		if err := k.Client.Get(ctx, name, cm); err != nil {
			if apierrors.IsNotFound(err) {
				return &WegoConfig{}, ErrWegoConfigNotFound
			}

			return nil, fmt.Errorf("failed getting weave-gitops configmap: %w", err)
		}
	} else {
		configMap, err := k.getWegoConfigMapFromAllNamespaces(ctx)
		if err != nil {
			return &WegoConfig{}, err
		}

		cm = configMap
	}

	wegoConfig := &WegoConfig{}
	if err := kyaml.Unmarshal([]byte(cm.Data["config"]), wegoConfig); err != nil {
		return nil, err
	}

	return wegoConfig, nil
}

func (k *KubeHTTP) getWegoConfigMapFromAllNamespaces(ctx context.Context) (*corev1.ConfigMap, error) {
	cml := &corev1.ConfigMapList{}
	if err := k.Client.List(ctx, cml); err != nil {
		return nil, fmt.Errorf("could not list weave-gitops config: %w", err)
	}

	for _, cm := range cml.Items {
		if cm.Name == WegoConfigMapName {
			return &cm, nil
		}
	}

	return nil, ErrWegoConfigNotFound
}

// FetchNamespaceWithLabel fetches a namespace where a label follows key=value
func (k *KubeHTTP) FetchNamespaceWithLabel(ctx context.Context, key string, value string) (*corev1.Namespace, error) {
	selector := labels.NewSelector()

	partOf, err := labels.NewRequirement(key, selection.Equals, []string{value})
	if err != nil {
		return nil, fmt.Errorf("bad requirement: %w", err)
	}

	selector = selector.Add(*partOf)

	options := client.ListOptions{
		LabelSelector: selector,
	}

	nsl := &corev1.NamespaceList{}
	if err := k.Client.List(ctx, nsl, &options); err != nil {
		return nil, fmt.Errorf("failed getting namespaces list: %w", err)
	}

	switch totalNamespaces := len(nsl.Items); {
	case totalNamespaces == 0:
		return nil, ErrNamespaceNotFound
	case totalNamespaces > 1:
		namespaces := make([]string, 0)
		for _, n := range nsl.Items {
			namespaces = append(namespaces, n.Name)
		}

		return nil, fmt.Errorf("found multiple namespaces %s with %s=%s, we are unable to define the correct one", namespaces, key, value)
	default:
		return &nsl.Items[0], nil
	}
}

func initialContext(cfgLoadingRules *clientcmd.ClientConfigLoadingRules) (currentCtx, clusterName string, err error) {
	rules, err := cfgLoadingRules.Load()
	if err != nil {
		return currentCtx, clusterName, err
	}

	if rules.CurrentContext == "" {
		return currentCtx, clusterName, fmt.Errorf("current context not found in kubeconfig file")
	}

	c := rules.Contexts[rules.CurrentContext]

	return rules.CurrentContext, sanitizeClusterName(c.Cluster), nil
}

func sanitizeClusterName(s string) string {
	// remove leading email address or username prefix from context
	if strings.Contains(s, "@") {
		return s[strings.LastIndex(s, "@")+1:]
	}

	s = strings.ReplaceAll(s, "_", "-")

	return s
}
