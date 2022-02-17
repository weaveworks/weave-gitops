package kube

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

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
	"github.com/weaveworks/weave-gitops/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	kyaml "sigs.k8s.io/yaml"
)

func CreateScheme() *apiruntime.Scheme {
	scheme := apiruntime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = kustomizev2.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = extensionsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)

	return scheme
}

const WeGOCRDName = "apps.wego.weave.works"
const FluxNamespace = "flux-system"

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

func NewKubeHTTPClientWithConfig(config *rest.Config, contextName string, additionalSchemes ...func(*apiruntime.Scheme) error) (Kube, client.Client, error) {
	scheme := CreateScheme()

	for _, add := range additionalSchemes {
		_ = add(scheme)
	}

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
