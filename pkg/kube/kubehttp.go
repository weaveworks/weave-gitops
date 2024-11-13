package kube

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta2"
	imgautomationv1 "github.com/fluxcd/image-automation-controller/api/v1beta1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	notificationv1 "github.com/fluxcd/notification-controller/api/v1"
	notificationv1b2 "github.com/fluxcd/notification-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/pkg/errors"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	appsv1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func CreateScheme() (*apiruntime.Scheme, error) {
	scheme := apiruntime.NewScheme()
	builder := apiruntime.SchemeBuilder{
		clientgoscheme.AddToScheme,
		sourcev1.AddToScheme,
		sourcev1b2.AddToScheme,
		kustomizev1.AddToScheme,
		helmv2.AddToScheme,
		corev1.AddToScheme,
		extensionsv1.AddToScheme,
		appsv1.AddToScheme,
		rbacv1.AddToScheme,
		authv1.AddToScheme,
		notificationv1.AddToScheme,
		notificationv1b2.AddToScheme,
		pacv2beta2.AddToScheme,
		reflectorv1.AddToScheme,
		imgautomationv1.AddToScheme,
	}

	err := builder.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("could not add to scheme: %w", err)
	}

	return scheme, nil
}

const WeGOCRDName = "apps.wego.weave.works"
const FluxNamespace = "flux-system"

const (
	WegoConfigMapName = "weave-gitops-config"
)

var (
	// ErrWegoConfigNotFound indicates weave gitops config could not be found
	ErrWegoConfigNotFound = errors.New("Wego Config not found")
)

// InClusterConfig defines a function for checking if this code is executing in kubernetes.
// This can be overriden if needed.
var InClusterConfig func() (*rest.Config, error) = rest.InClusterConfig

var ErrNamespaceNotFound = errors.New("namespace not found")

func NewKubeHTTPClientWithConfig(config *rest.Config, contextName string, additionalSchemes ...func(*apiruntime.Scheme) error) (*KubeHTTP, error) {
	scheme, err := CreateScheme()
	if err != nil {
		return nil, fmt.Errorf("could not create scheme: %w", err)
	}

	for _, add := range additionalSchemes {
		err = add(scheme)
		if err != nil {
			return nil, fmt.Errorf("could not add to scheme: %w", err)
		}
	}

	rawClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("kubernetes client initialization failed: %w", err)
	}

	return &KubeHTTP{Client: rawClient, ClusterName: contextName}, nil
}

func NewKubeHTTPClient() (*KubeHTTP, error) {
	config, contextName, err := RestConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create default config: %w", err)
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

	return config, InClusterConfigClusterName(), nil
}

func InClusterConfigClusterName() string {
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
