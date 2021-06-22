package kube

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func CreateScheme() *apiruntime.Scheme {
	scheme := apiruntime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = kustomizev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	_ = wego.AddToScheme(scheme)

	return scheme
}

func NewKubeHTTPClient() (Kube, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	_, kubeContext, err := initialContexts(cfgLoadingRules)
	if err != nil {
		return nil, fmt.Errorf("could not get initial context: %s", err)
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

	return &KubeHTTP{Client: kubeClient, ClusterName: kubeContext}, nil
}

// This is an alternative implementation of the kube.Kube interface,
// specifically designed to query the K8s API directly instead of relying on
// `kubectl` to be present in the PATH.
type KubeHTTP struct {
	Client      client.Client
	ClusterName string
}

func (c *KubeHTTP) GetClusterName() (string, error) {
	return c.ClusterName, nil
}

func (c *KubeHTTP) GetClusterStatus() ClusterStatus {
	tName := types.NamespacedName{
		Name: "apps.wego.weave.works",
	}

	crd := v1.CustomResourceDefinition{}

	if c.Client.Get(context.Background(), tName, &crd) == nil {
		return WeGOInstalled
	}

	return Unknown
}

func (c *KubeHTTP) Apply(manifests []byte, namespace string) ([]byte, error) {
	return nil, errors.New("Apply not implemented for kubeHTTP")
}

func (c *KubeHTTP) GetApplication(name string) (*wego.Application, error) {
	tName := types.NamespacedName{
		Name:      name,
		Namespace: "default",
	}
	app := wego.Application{}
	if err := c.Client.Get(context.Background(), tName, &app); err != nil {
		return nil, fmt.Errorf("could not get application: %s", err)
	}

	return &app, nil
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
