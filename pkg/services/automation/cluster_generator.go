package automation

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/yaml"
)

type ClusterAutomation struct {
	AppCRD                      AutomationManifest
	GitOpsRuntime               AutomationManifest
	SourceManifest              AutomationManifest
	SystemKustomizationManifest AutomationManifest
	SystemKustResourceManifest  AutomationManifest
	UserKustResourceManifest    AutomationManifest
	WegoAppManifest             AutomationManifest
	WegoConfigManifest          AutomationManifest
}

const (
	AppCRDPath              = "wego-system.yaml"
	RuntimePath             = "gitops-runtime.yaml"
	SourcePath              = "flux-source-resource.yaml"
	SystemKustomizationPath = "kustomization.yaml"
	SystemKustResourcePath  = "flux-system-kustomization-resource.yaml"
	UserKustResourcePath    = "flux-user-kustomization-resource.yaml"
	WegoAppPath             = "wego-app.yaml"
	WegoConfigPath          = "wego-config.yaml"

	WegoConfigMapName = "weave-gitops-config"
)

func createClusterSourceName(gitSourceURL gitproviders.RepoURL) string {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-auto-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return lengthConstrainedName
}

func (a *AutomationGen) GenerateClusterAutomation(ctx context.Context, cluster models.Cluster, configURL gitproviders.RepoURL, namespace string, fluxNamespace string) (ClusterAutomation, error) {
	systemPath := filepath.Join(git.WegoRoot, git.WegoClusterDir, cluster.Name, git.WegoClusterOSWorkloadDir)
	userPath := filepath.Join(git.WegoRoot, git.WegoClusterDir, cluster.Name, git.WegoClusterUserWorkloadDir)

	systemQualifiedPath := func(relativePath string) string {
		return filepath.Join(systemPath, relativePath)
	}

	secretRef, err := a.GetSecretRef(ctx, configURL)
	if err != nil {
		return ClusterAutomation{}, err
	}

	secretStr := secretRef.String()

	configBranch, err := a.GitProvider.GetDefaultBranch(ctx, configURL)
	if err != nil {
		return ClusterAutomation{}, err
	}

	runtimeManifests, err := a.Flux.Install(namespace, true)
	if err != nil {
		return ClusterAutomation{}, err
	}

	appCRDManifest := manifests.AppCRD

	version := version.Version
	if os.Getenv("IS_TEST_ENV") != "" {
		version = "latest"
	}

	m, err := manifests.GenerateWegoAppManifests(manifests.WegoAppParams{Version: version, Namespace: namespace})
	if err != nil {
		return ClusterAutomation{}, fmt.Errorf("error generating wego-app manifest: %w", err)
	}

	wegoAppManifest := bytes.Join(m, []byte("---\n"))

	sourceName := createClusterSourceName(configURL)

	sourceManifest, err := a.Flux.CreateSourceGit(sourceName, configURL, configBranch, secretStr, namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-system", cluster.Name)), sourceName,
		workAroundFluxDroppingDot(systemPath), namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	userKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-user", cluster.Name)), sourceName,
		workAroundFluxDroppingDot(userPath), namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustomization := CreateKustomize(cluster.Name, namespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return ClusterAutomation{}, err
	}

	config := kube.WegoConfig{
		FluxNamespace: fluxNamespace,
		WegoNamespace: namespace,
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return ClusterAutomation{}, fmt.Errorf("failed marshalling wego config: %w", err)
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

	wegoConfigManifest, err := yaml.Marshal(cm)
	if err != nil {
		return ClusterAutomation{}, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	return ClusterAutomation{
		AppCRD: AutomationManifest{
			Path:    systemQualifiedPath(AppCRDPath),
			Content: appCRDManifest,
		},
		GitOpsRuntime: AutomationManifest{
			Path:    systemQualifiedPath(RuntimePath),
			Content: runtimeManifests,
		},
		SourceManifest: AutomationManifest{
			Path:    systemQualifiedPath(SourcePath),
			Content: sourceManifest,
		},
		SystemKustomizationManifest: AutomationManifest{
			Path:    systemQualifiedPath(SystemKustomizationPath),
			Content: systemKustomizationManifest,
		},
		SystemKustResourceManifest: AutomationManifest{
			Path:    systemQualifiedPath(SystemKustResourcePath),
			Content: systemKustResourceManifest,
		},
		UserKustResourceManifest: AutomationManifest{
			Path:    systemQualifiedPath(UserKustResourcePath),
			Content: userKustResourceManifest,
		},
		WegoAppManifest: AutomationManifest{
			Path:    systemQualifiedPath(WegoAppPath),
			Content: wegoAppManifest,
		},
		WegoConfigManifest: AutomationManifest{
			Path:    systemQualifiedPath(WegoConfigPath),
			Content: wegoConfigManifest,
		},
	}, nil
}

func (ca ClusterAutomation) BootstrapManifests() []AutomationManifest {
	return append([]AutomationManifest{ca.AppCRD}, ca.WegoAppManifest, ca.SourceManifest, ca.SystemKustResourceManifest, ca.UserKustResourceManifest, ca.WegoConfigManifest)
}

func (ca ClusterAutomation) Manifests() []AutomationManifest {
	return append(ca.BootstrapManifests(), ca.GitOpsRuntime, ca.SystemKustomizationManifest)
}

func GetClusterHash(c models.Cluster) string {
	return fmt.Sprintf("wego-%x", md5.Sum([]byte(c.Name)))
}

func workAroundFluxDroppingDot(str string) string {
	return "." + str
}
