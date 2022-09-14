package run

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
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
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/api/resource"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

const (
	helmChartName     = "weave-gitops"
	helmRepositoryURL = "https://helm.gitops.weave.works"
)

func ReadPassword(log logger.Logger) (string, error) {
	password, err := utils.ReadPasswordFromStdin(log, "Please enter your password to generate the secret: ")
	if err != nil {
		log.Failuref("Could not read password")
		return "", err
	}

	return password, nil
}

func GenerateSecret(log logger.Logger, password string) (string, error) {
	secret, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Failuref("Error generating secret from password")
		return "", err
	}

	log.Successf("Secret has been generated:")

	return string(secret), nil
}

// CreateDashboardObjects creates HelmRepository and HelmRelease objects for the GitOps Dashboard installation.
func CreateDashboardObjects(log logger.Logger, name string, namespace string, username string, secret string, chartVersion string) ([]byte, error) {
	log.Actionf("Creating GitOps Dashboard objects ...")

	helmRepository := makeHelmRepository(name, namespace)
	helmRelease, err := makeHelmRelease(log, name, namespace, username, secret, chartVersion)

	if err != nil {
		log.Failuref("Creating HelmRelease failed")
		return nil, err
	}

	log.Generatef("Generating GitOps Dashboard manifests ...")

	manifests, err := generateManifestsForDashboard(log, helmRepository, helmRelease)
	if err != nil {
		log.Failuref("Generating GitOps Dashboard manifests failed")
		return nil, err
	}

	return manifests, nil
}

// InstallDashboard installs the GitOps Dashboard.
func InstallDashboard(log logger.Logger, ctx context.Context, manager ResourceManagerForApply, manifests []byte) error {
	log.Actionf("Installing the GitOps Dashboard ...")

	applyOutput, err := apply(log, ctx, manager, manifests)
	if err != nil {
		log.Failuref("GitOps Dashboard install failed")
		return err
	}

	log.Successf("GitOps Dashboard has been installed")

	fmt.Println(applyOutput)

	return nil
}

// IsDashboardInstalled checks if the GitOps Dashboard is installed.
func IsDashboardInstalled(log logger.Logger, ctx context.Context, kubeClient client.Client, name string, namespace string) bool {
	return getDashboardHelmChart(log, ctx, kubeClient, name, namespace) != nil
}

// GetDashboardHelmChart checks if the GitOps Dashboard is installed.
func getDashboardHelmChart(log logger.Logger, ctx context.Context, kubeClient client.Client, name string, namespace string) *sourcev1.HelmChart {
	helmChart := sourcev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace + "-" + name,
			Namespace: namespace,
		},
	}

	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&helmChart), &helmChart); err != nil {
		return nil
	}

	return &helmChart
}

// EnablePortForwardingForDashboard enables port forwarding for the GitOps Dashboard.
func EnablePortForwardingForDashboard(log logger.Logger, kubeClient client.Client, config *rest.Config, namespace string, podName string, dashboardPort string) (func(), error) {
	specMap := &PortForwardSpec{
		Namespace:     namespace,
		Name:          podName,
		Kind:          "deployment",
		HostPort:      dashboardPort,
		ContainerPort: server.DefaultPort,
	}
	// get pod from specMap
	namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

	pod, err := GetPodFromResourceDescription(namespacedName, specMap.Kind, kubeClient)
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

	return nil, ErrDashboardPodNotFound
}

// ReconcileDashboard reconciles the dashboard.
func ReconcileDashboard(kubeClient client.Client, name string, namespace string, podName string, timeout time.Duration) error {
	const interval = 3 * time.Second / 2

	helmChartName := namespace + "-" + name

	// reconcile dashboard
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      helmChartName,
	}
	gvk := schema.GroupVersionKind{
		Group:   "source.toolkit.fluxcd.io",
		Version: "v1beta2",
		Kind:    "HelmChart",
	}

	var sourceRequestedAt string

	if err := wait.Poll(interval, timeout, func() (bool, error) {
		var err error
		sourceRequestedAt, err = requestReconciliation(context.Background(), kubeClient,
			namespacedName, gvk)

		return err == nil, nil
	}); err != nil {
		return err
	}

	// wait for the reconciliation of dashboard to be done
	if err := wait.Poll(interval, timeout, func() (bool, error) {
		dashboard := &sourcev1.HelmChart{}
		if err := kubeClient.Get(context.Background(), types.NamespacedName{
			Namespace: namespace,
			Name:      helmChartName,
		}, dashboard); err != nil {
			return false, err
		}

		return dashboard.Status.GetLastHandledReconcileRequest() == sourceRequestedAt, nil
	}); err != nil {
		return err
	}

	// wait for dashboard to be ready
	if err := wait.Poll(interval, timeout, func() (bool, error) {
		namespacedName := types.NamespacedName{Namespace: namespace, Name: podName}

		dashboard, _ := GetPodFromResourceDescription(namespacedName, "deployment", kubeClient)
		if dashboard == nil {
			return false, nil
		}

		return isPodStatusConditionPresentAndEqual(dashboard.Status.Conditions, corev1.PodReady, corev1.ConditionTrue), nil
	}); err != nil {
		return err
	}

	return nil
}

// generateManifestsForDashboard generates dashboard manifests from objects.
func generateManifestsForDashboard(log logger.Logger, helmRepository *sourcev1.HelmRepository, helmRelease *helmv2.HelmRelease) ([]byte, error) {
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

	sanitizedHelmRepositoryData, err := sanitizeResourceData(log, helmRepositoryData)
	if err != nil {
		log.Failuref("Error sanitizing HelmRepository data")
		return nil, err
	}

	sanitizedHelmReleaseData, err := sanitizeResourceData(log, helmReleaseData)
	if err != nil {
		log.Failuref("Error sanitizing HelmRelease data")
		return nil, err
	}

	divider := []byte("---\n")

	content := sanitizedHelmRepositoryData
	content = append(content, divider...)
	content = append(content, sanitizedHelmReleaseData...)

	return content, nil
}

// makeHelmRepository creates a HelmRepository object for installing the GitOps Dashboard.
func makeHelmRepository(name string, namespace string) *sourcev1.HelmRepository {
	helmRepository := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "weave-gitops-dashboard",
				"app.kubernetes.io/component":  "ui",
				"app.kubernetes.io/part-of":    "weave-gitops",
				"app.kubernetes.io/created-by": "weave-gitops-cli",
			},
			Annotations: map[string]string{
				"metadata.weave.works/description": "This is the Weave GitOps Dashboard.  It provides a simple way to get insights into your GitOps workloads.",
			},
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: helmRepositoryURL,
			Interval: metav1.Duration{
				Duration: time.Minute * 60,
			},
		},
	}

	return helmRepository
}

// makeHelmRelease creates a HelmRelease object for installing the GitOps Dashboard.
func makeHelmRelease(log logger.Logger, name string, namespace string, username string, secret string, chartVersion string) (*helmv2.HelmRelease, error) {
	helmRelease := &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{
				Duration: 60 * time.Minute,
			},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: helmChartName,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "HelmRepository",
						Name: name,
					},
				},
			},
		},
	}

	if chartVersion != "" {
		helmRelease.Spec.Chart.Spec.Version = chartVersion
	}

	if username != "" && secret != "" {
		values, err := makeValues(username, secret)
		if err != nil {
			log.Failuref("Error generating values from secret")
			return nil, err
		}

		helmRelease.Spec.Values = &apiextensionsv1.JSON{Raw: values}
	}

	return helmRelease, nil
}

// makeValues creates a values object for installing the GitOps Dashboard.
func makeValues(username string, secret string) ([]byte, error) {
	valuesMap := map[string]interface{}{
		"adminUser": map[string]interface{}{
			"create":       true,
			"username":     username,
			"passwordHash": secret,
		},
	}

	jsonRaw, err := json.Marshal(valuesMap)
	if err != nil {
		return nil, fmt.Errorf("encoding values failed: %w", err)
	}

	return jsonRaw, nil
}

func sanitizeResourceData(log logger.Logger, resourceData []byte) ([]byte, error) {
	// remove status
	resNode, err := kyaml.Parse(string(resourceData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource data: %v", err.Error())
	}

	res := &resource.Resource{RNode: *resNode}

	err = res.PipeE(kyaml.FieldClearer{Name: "status"})
	if err != nil {
		return nil, fmt.Errorf("failed to remove status: %v", err.Error())
	}

	// remove creationTimestamp
	metadataNode, err := res.Pipe(kyaml.Get("metadata"))
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %v", err.Error())
	}

	metadataRes := &resource.Resource{RNode: *metadataNode}

	err = metadataRes.PipeE(kyaml.FieldClearer{Name: "creationTimestamp"})
	if err != nil {
		return nil, fmt.Errorf("failed to remove creationTimestamp: %v", err.Error())
	}

	err = res.PipeE(kyaml.FieldSetter{Name: "metadata", Value: &metadataRes.RNode})
	if err != nil {
		return nil, fmt.Errorf("failed to set metadata: %v", err.Error())
	}

	resourceData, err = res.AsYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to YAML: %v", err.Error())
	}

	return resourceData, nil
}
