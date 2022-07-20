package run

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"
)

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

	manifests, err := createManifests(string(secret), helmRepository, helmRelease)
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

func createManifests(secret string, helmRepository *sourcev1.HelmRepository, helmRelease *helmv2.HelmRelease) ([]byte, error) {
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

func makeValues(secret string) ([]byte, error) {
	valuesMap := make(map[string]interface{})
	valuesMap["adminUser"] = map[string]interface{}{
		"create":       true,
		"passwordHash": secret,
		"username":     "admin",
	}

	jsonRaw, err := json.Marshal(valuesMap)
	if err != nil {
		return nil, fmt.Errorf("marshaling values failed: %w", err)
	}

	return jsonRaw, nil
}
