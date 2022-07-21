package run

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Installs the GitOps Dashboard.
func InstallDashboard(log logger.Logger, ctx context.Context, kubeClient *kube.KubeHTTP, kubeConfigArgs *genericclioptions.ConfigFlags) error {
	password, err := utils.ReadPasswordFromStdin("Please enter your password to generate your secret: ")
	if err != nil {
		return fmt.Errorf("could not read password: %w", err)
	}

	secret, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	log.Successf("Secret has been generated:")
	fmt.Println(string(secret))

	log.Actionf("Installing GitOps Dashboard ...")

	helmRepository := makeHelmRepository()
	helmRelease := makeHelmRelease(string(secret))

	manifests, err := generateManifests(string(secret), helmRepository, helmRelease)
	if err != nil {
		log.Failuref("Creating GitOps Dashboard manifests failed")
		return err
	}

	applyOutput, err := apply(log, ctx, kubeClient, kubeConfigArgs, manifests)
	if err != nil {
		log.Failuref("GitOps Dashboard install failed")
		return err
	}

	log.Successf("GitOps Dashboard has been installed")

	fmt.Println(applyOutput)

	return nil
}

// Checks if the GitOps Dashboard is installed.
func IsDashboardInstalled(log logger.Logger, ctx context.Context, kubeClient *kube.KubeHTTP) bool {
	helmChart := sourcev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: wego.DefaultNamespace,
			Name:      fmt.Sprintf("%s-ww-gitops", wego.DefaultNamespace),
		},
	}
	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&helmChart), &helmChart); err != nil {
		return false
	}

	return true
}

const (
	dashboard = "ww-gitops-weave-gitops"
)

// Checks if the GitOps Dashboard is installed.
func OpenDashboardPort(log logger.Logger, kubeClient *kube.KubeHTTP, config *rest.Config, dashboardPort string) (func(), error) {
	specMap := &PortForwardSpec{
		Namespace:     "flux-system",
		Name:          dashboard,
		Kind:          "service",
		HostPort:      dashboardPort,
		ContainerPort: server.DefaultPort,
	}
	// get pod from specMap
	pod, err := GetPodFromSpecMap(specMap, kubeClient)
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

		log.Successf("Port forwarding for dev-bucket is ready.")

		return cancelPortFwd, nil
	}

	return nil, fmt.Errorf("pod not found")
}

// Generates GitOps Dashboard manifests from objects.
func generateManifests(secret string, helmRepository *sourcev1.HelmRepository, helmRelease *helmv2.HelmRelease) ([]byte, error) {
	helmRepositoryData, err := yaml.Marshal(helmRepository)
	if err != nil {
		return nil, err
	}

	helmReleaseData, err := yaml.Marshal(helmRelease)
	if err != nil {
		return nil, err
	}

	divider := []byte("---\n")

	content := append(divider, helmRepositoryData...)
	content = append(content, divider...)
	content = append(content, helmReleaseData...)

	return content, nil
}

// Creates a HelmRepository object for installing the GitOps Dashboard.
func makeHelmRepository() *sourcev1.HelmRepository {
	helmRepository := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ww-gitops",
			Namespace: "flux-system",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: "https://helm.gitops.weave.works",
			Interval: metav1.Duration{
				Duration: time.Minute,
			},
		},
	}

	return helmRepository
}

// Creates a HelmRelease object for installing the GitOps Dashboard.
func makeHelmRelease(secret string) *helmv2.HelmRelease {
	helmRelease := &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ww-gitops",
			Namespace: "flux-system",
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{
				Duration: time.Minute,
			},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   "weave-gitops",
					Version: "2.0.6",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "HelmRepository",
						Name: "ww-gitops",
					},
					ReconcileStrategy: "ChartVersion",
				},
			},
			Suspend: false,
		},
	}

	valuesData, _ := makeValues(secret)

	helmRelease.Spec.Values = &apiextensionsv1.JSON{Raw: valuesData}

	return helmRelease
}

// Creates a values object for installing the GitOps Dashboard.
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
		return nil, fmt.Errorf("marshaling values failed: %w", err)
	}

	return jsonRaw, nil
}
