package automation

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/yaml"
)

type ClusterAutomation struct {
	AppCRD                      Manifest
	GitOpsRuntime               Manifest
	SourceManifest              Manifest
	SystemKustomizationManifest Manifest
	SystemKustResourceManifest  Manifest
	UserKustResourceManifest    Manifest
	WegoAppManifest             Manifest

	// Do I really need this here? This is not needed for automation itself
	// Maybe we need to change the name to something like ClusterManifests
	WegoConfigManifest Manifest
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

func (a *AutomationGen) GenerateClusterAutomation(ctx context.Context, cluster models.Cluster, configURL gitproviders.RepoURL, namespace string) (ClusterAutomation, error) {
	secretRef, err := a.GetSecretRefForPrivateGitSources(ctx, configURL)
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

	m, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: version, Namespace: namespace})
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
		workAroundFluxDroppingDot(git.GetSystemPath(cluster.Name)), namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	userKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-user", cluster.Name)), sourceName,
		workAroundFluxDroppingDot(git.GetUserPath(cluster.Name)), namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustomization := CreateKustomize(cluster.Name, namespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return ClusterAutomation{}, err
	}

	return ClusterAutomation{
		AppCRD: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, AppCRDPath),
			Content: appCRDManifest,
		},
		GitOpsRuntime: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, RuntimePath),
			Content: runtimeManifests,
		},
		SourceManifest: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, SourcePath),
			Content: sourceManifest,
		},
		SystemKustomizationManifest: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, SystemKustomizationPath),
			Content: systemKustomizationManifest,
		},
		SystemKustResourceManifest: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, SystemKustResourcePath),
			Content: systemKustResourceManifest,
		},
		UserKustResourceManifest: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, UserKustResourcePath),
			Content: userKustResourceManifest,
		},
		WegoAppManifest: Manifest{
			Path:    git.GetSystemQualifiedPath(cluster.Name, WegoAppPath),
			Content: wegoAppManifest,
		},
	}, nil
}

func (ca ClusterAutomation) BootstrapManifests() []Manifest {
	return append([]Manifest{ca.AppCRD}, ca.WegoAppManifest, ca.SourceManifest, ca.SystemKustResourceManifest, ca.UserKustResourceManifest)
}

func gitopsConfigMap(fluxNamespace string, wegoNamespace string) (corev1.ConfigMap, error) {
	config := kube.WegoConfig{
		FluxNamespace: fluxNamespace,
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return corev1.ConfigMap{}, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	return corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      WegoConfigMapName,
			Namespace: wegoNamespace,
		},
		Data: map[string]string{
			"config": string(configBytes),
		},
	}, nil

}

func (ca ClusterAutomation) Manifests() []Manifest {
	return append(ca.BootstrapManifests(), ca.GitOpsRuntime, ca.SystemKustomizationManifest)
}

func GetClusterHash(c models.Cluster) string {
	return fmt.Sprintf("wego-%x", md5.Sum([]byte(c.Name)))
}

func workAroundFluxDroppingDot(str string) string {
	return "." + str
}

func GitopsManifests(fluxClient flux.Flux, kubeClient kube.Kube, namespace string, configURL gitproviders.RepoURL) ([][]byte, error) {
	ctx := context.Background()

	clusterName, err := kubeClient.GetClusterName(ctx)
	if err != nil {
		return nil, err
	}

	runtimeManifests, err := fluxClient.Install(namespace, true)
	if err != nil {
		return nil, fmt.Errorf("failed getting runtime manifests: %w", err)
	}

	appCRDManifest := manifests.AppCRD

	version := version.Version
	if os.Getenv("IS_TEST_ENV") != "" {
		version = "latest"
	}

	wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: version, Namespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("error generating wego-app manifest: %w", err)
	}

	wegoAppManifest := bytes.Join(wegoAppManifests, []byte("---\n"))

	sourceName := createClusterSourceName(configURL)

	//sourceManifest, err := flux.CreateSourceGit(sourceName, configURL, configBranch, secretStr, namespace)
	//if err != nil {
	//	return ClusterAutomation{}, err
	//}

	systemKustResourceManifest, err := fluxClient.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-system", clusterName)), sourceName,
		workAroundFluxDroppingDot(git.GetSystemPath(clusterName)), namespace)
	if err != nil {
		return nil, err
	}

	userKustResourceManifest, err := fluxClient.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-user", clusterName)), sourceName,
		workAroundFluxDroppingDot(git.GetUserPath(clusterName)), namespace)
	if err != nil {
		return nil, err
	}

	systemKustomization := CreateKustomize(clusterName, namespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return nil, err
	}

	gitopsConfigMap, err := gitopsConfigMap(namespace, namespace)
	if err != nil {
		return nil, err
	}
	wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	return [][]byte{
		appCRDManifest,
		runtimeManifests,
		systemKustomizationManifest,
		systemKustResourceManifest,
		userKustResourceManifest,
		wegoAppManifest,
		wegoConfigManifest,
	}, nil

	//return ClusterAutomation{
	//	AppCRD: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, AppCRDPath),
	//		Content: appCRDManifest,
	//	},
	//	GitOpsRuntime: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, RuntimePath),
	//		Content: runtimeManifests,
	//	},
	//	// This can't be generated here as we need the Branch and SecretRef, but we
	//	// would need git provider to get the default branch,
	//	// but we need to run the secret logic first to be able to use the gitProvider client
	//	//SourceManifest: AutomationManifest{
	//	//	Path:    git.GetSystemQualifiedPath(clusterName, SourcePath),
	//	//	Content: sourceManifest,
	//	//},
	//	SystemKustomizationManifest: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, SystemKustomizationPath),
	//		Content: systemKustomizationManifest,
	//	},
	//	SystemKustResourceManifest: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, SystemKustResourcePath),
	//		Content: systemKustResourceManifest,
	//	},
	//	UserKustResourceManifest: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, UserKustResourcePath),
	//		Content: userKustResourceManifest,
	//	},
	//	WegoAppManifest: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, WegoAppPath),
	//		Content: wegoAppManifest,
	//	},
	//	WegoConfigManifest: Manifest{
	//		Path:    git.GetSystemQualifiedPath(clusterName, WegoConfigPath),
	//		Content: wegoConfigManifest,
	//	},
	//}, nil
}
