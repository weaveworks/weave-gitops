package kube

import (
	"context"
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/pkg/errors"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

func (c *KubeHTTP) Apply(manifests []byte, namespace string) ([]byte, error) {
	return nil, errors.New("Apply is not implemented for kubeHTTP")
}

func (c *KubeHTTP) GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
	app := wego.Application{}
	if err := c.Client.Get(ctx, name, &app); err != nil {
		return nil, fmt.Errorf("could not get application: %s", err)
	}

	return &app, nil
}

func (c *KubeHTTP) Delete(manifests []byte, namespace string) ([]byte, error) {
	return nil, errors.New("not implemented")
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
		return false, fmt.Errorf("could not find flux namespace: %s", err)
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
		return false, fmt.Errorf("could not get secret: %s", err)
	}

	return true, nil
}

func (c *KubeHTTP) GetApplications(ctx context.Context, namespace string) ([]wego.Application, error) {
	result := wego.ApplicationList{}

	if err := c.Client.List(ctx, &result, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("could not list wego applications: %s", err)
	}

	return result.Items, nil
}

func (c *KubeHTTP) LabelExistsInCluster(ctx context.Context, label string) error {
	return errors.New("LabelExistsInCluster is not implemented for kubeHTTP")
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

// func namespaceOpts(ns string) *client.ListOptions {
// 	opts := client.ListOptions{}
// 	if ns != "" {
// 		opts.Namespace = ns
// 	}

// 	return &opts
// }
