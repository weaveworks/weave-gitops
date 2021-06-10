package utils

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	notificationv1 "github.com/fluxcd/notification-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
)

func CreateScheme() *apiruntime.Scheme {
	scheme := apiruntime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = sourcev1.AddToScheme(scheme)
	_ = kustomizev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)
	_ = notificationv1.AddToScheme(scheme)

	return scheme
}

func NewKubeClient(kubeContext string) (client.Client, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
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

	return kubeClient, nil
}
