package run

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	helmRepositoryName      = "ww-gitops"
	helmReleaseName         = "ww-gitops"
	helmChartName           = "weave-gitops"
	helmChartNamespacedName = wego.DefaultNamespace + "-ww-gitops"
	podName                 = "ww-gitops-weave-gitops"
	helmRepositoryUrl       = "https://helm.gitops.weave.works"
)

// InstallDashboard installs the GitOps Dashboard.
func InstallDashboard(log logger.Logger, ctx context.Context, kubeClient client.Client, kubeConfigArgs *genericclioptions.ConfigFlags) error {
	log.Actionf("Installing the GitOps Dashboard ...")

	password, err := utils.ReadPasswordFromStdin(log, "Please enter your password to generate the secret: ")
	if err != nil {
		log.Failuref("Could not read password")
		return err
	}

	secret, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Failuref("Error generating secret from password")
		return err
	}

	log.Successf("Secret has been generated:")
	fmt.Println(string(secret))

	log.Actionf("Installing the GitOps Dashboard ...")

	helmRepository := makeHelmRepository()
	helmRelease, err := makeHelmRelease(log, string(secret))

	if err != nil {
		log.Failuref("Creating HelmRelease failed")
		return err
	}

	manifests, err := generateManifests(log, string(secret), helmRepository, helmRelease)
	if err != nil {
		log.Failuref("Generating GitOps Dashboard manifests failed")
		return err
	}

	log.Successf("Generated GitOps Dashboard manifests")

	applyOutput, err := apply(log, ctx, kubeClient, kubeConfigArgs, manifests)
	if err != nil {
		log.Failuref("GitOps Dashboard install failed")
		return err
	}

	log.Successf("GitOps Dashboard has been installed")

	fmt.Println(applyOutput)

	return nil
}

// IsDashboardInstalled checks if the GitOps Dashboard is installed.
func IsDashboardInstalled(log logger.Logger, ctx context.Context, kubeClient client.Client) bool {
	helmChart := sourcev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: wego.DefaultNamespace,
			Name:      helmChartNamespacedName,
		},
	}
	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&helmChart), &helmChart); err != nil {
		return false
	}

	return true
}

// EnablePortForwardingForDashboard enables port forwarding for the GitOps Dashboard.
func EnablePortForwardingForDashboard(log logger.Logger, kubeClient client.Client, config *rest.Config, namespace string, dashboardPort string) (func(), error) {
	specMap := &PortForwardSpec{
		Namespace:     namespace,
		Name:          podName,
		Kind:          "service",
		HostPort:      dashboardPort,
		ContainerPort: server.DefaultPort,
	}
	// get pod from specMap
	pod, err := GetPodFromSpecMap(specMap, kubeClient, corev1.PodRunning)
	if err != nil {
		log.Failuref("Error getting pod from specMap: %v", err)
	}

	if pod != nil {
		waitFwd := make(chan struct{}, 1)
		readyChannel := make(chan struct{})
		cancelPortFwd := func() {
			close(waitFwd)
		}

		log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

		go func() {
			if err := ForwardPort(pod, config, specMap, waitFwd, readyChannel); err != nil {
				log.Failuref("Error forwarding port: %v", err)
			}
		}()
		<-readyChannel

		log.Successf("Port forwarding for dashboard is ready.")

		return cancelPortFwd, nil
	}

	return nil, fmt.Errorf("dashboard pod not found")
}

// ReconcileDashboard reconciles the dashboard.
func ReconcileDashboard(kubeClient client.Client, namespace string, timeout time.Duration, dashboardPort string) error {
	const interval = 3 * time.Second / 2

	// reconcile dashboard
	sourceRequestedAt, err := RequestReconciliation(context.Background(), kubeClient,
		types.NamespacedName{
			Namespace: namespace,
			Name:      helmChartNamespacedName,
		}, schema.GroupVersionKind{
			Group:   "source.toolkit.fluxcd.io",
			Version: "v1beta2",
			Kind:    "HelmChart",
		})
	if err != nil {
		return err
	}

	// wait for the reconciliation of dashboard to be done
	if err := wait.Poll(interval, timeout, func() (bool, error) {
		dashboard := &sourcev1.HelmChart{}
		if err := kubeClient.Get(context.Background(), types.NamespacedName{
			Namespace: namespace,
			Name:      helmChartNamespacedName,
		}, dashboard); err != nil {
			return false, err
		}

		return dashboard.Status.GetLastHandledReconcileRequest() == sourceRequestedAt, nil
	}); err != nil {
		return err
	}

	// wait for dashboard to be ready
	if err := wait.Poll(interval, timeout, func() (bool, error) {
		specMap := &PortForwardSpec{
			Namespace:     wego.DefaultNamespace,
			Name:          podName,
			Kind:          "service",
			HostPort:      dashboardPort,
			ContainerPort: server.DefaultPort,
		}

		dashboard, err := GetPodFromSpecMap(specMap, kubeClient, "")
		if err != nil {
			return false, err
		}

		return IsPodStatusConditionPresentAndEqual(dashboard.Status.Conditions, corev1.PodReady, corev1.ConditionTrue), nil
	}); err != nil {
		return err
	}

	return nil
}

// generateManifests generates dashboard manifests from objects.
func generateManifests(log logger.Logger, secret string, helmRepository *sourcev1.HelmRepository, helmRelease *helmv2.HelmRelease) ([]byte, error) {
	helmRepositoryData, err := yaml.Marshal(helmRepository)
	if err != nil {
		log.Failuref("Error generating HelmRepository manifest from object")
		return nil, err
	}

	helmReleaseData, err := yaml.Marshal(helmRelease)
	if err != nil {
		log.Failuref("Error generating HelmRelease manifest from object")
		return nil, err
	}

	divider := []byte("---\n")

	content := append(divider, helmRepositoryData...)
	content = append(content, divider...)
	content = append(content, helmReleaseData...)

	return content, nil
}

// makeHelmRepository creates a HelmRepository object for installing the GitOps Dashboard.
func makeHelmRepository() *sourcev1.HelmRepository {
	helmRepository := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmRepositoryName,
			Namespace: wego.DefaultNamespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: helmRepositoryUrl,
			Interval: metav1.Duration{
				Duration: time.Minute,
			},
		},
	}

	return helmRepository
}

// makeHelmRelease creates a HelmRelease object for installing the GitOps Dashboard.
func makeHelmRelease(log logger.Logger, secret string) (*helmv2.HelmRelease, error) {
	helmRelease := &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmReleaseName,
			Namespace: wego.DefaultNamespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{
				Duration: time.Minute,
			},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   helmChartName,
					Version: "2.0.6",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "HelmRepository",
						Name: helmRepositoryName,
					},
					ReconcileStrategy: "ChartVersion",
				},
			},
		},
	}

	valuesData, err := makeValues(secret)
	if err != nil {
		log.Failuref("Error generating values from secret")
		return nil, err
	}

	helmRelease.Spec.Values = &apiextensionsv1.JSON{Raw: valuesData}

	return helmRelease, nil
}

// makeValues creates a values object for installing the GitOps Dashboard.
func makeValues(secret string) ([]byte, error) {
	valuesMap := map[string]interface{}{
		"adminUser": map[string]interface{}{
			"create":       true,
			"passwordHash": secret,
			"username":     "admin",
		},
	}

	jsonRaw, err := json.Marshal(valuesMap)
	if err != nil {
		return nil, fmt.Errorf("encoding values failed: %w", err)
	}

	return jsonRaw, nil
}
