package automation

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	// "strings"
	// "github.com/fluxcd/go-git-providers/gitprovider"
	// sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	// "github.com/fluxcd/source-controller/pkg/sourceignore"
	// "github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	// "github.com/weaveworks/weave-gitops/pkg/kube"
	// "github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	// wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type ClusterAutomation struct {
	GitOpsRuntime               AutomationManifest
	AppCRD                      AutomationManifest
	WegoAppManifest             AutomationManifest
	SourceManifest              AutomationManifest
	SystemKustResourceManifest  AutomationManifest
	UserKustResourceManifest    AutomationManifest
	SystemKustomizationManifest AutomationManifest
}

const (
	RuntimePath             = "gitops-runtime.yaml"
	AppCRDPath              = "wego-system.yaml"
	WegoAppPath             = "wego-app.yaml"
	SourcePath              = "flux-source-resource.yaml"
	SystemKustResourcePath  = "flux-system-kustomization-resource.yaml"
	UserKustResourcePath    = "flux-user-kustomization-resource.yaml"
	SystemKustomizationPath = "kustomization.yaml"
)

func (a *AutomationGen) GenerateClusterAutomation(ctx context.Context, cluster models.Cluster, configURL gitproviders.RepoURL, namespace string) (ClusterAutomation, error) {
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

	sourceManifest, err := a.Flux.CreateSourceGit(secretStr, configURL, configBranch, secretStr, namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-system", cluster.Name)), secretStr,
		workAroundFluxDroppingDot(systemPath), namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	userKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-user", cluster.Name)), secretStr,
		workAroundFluxDroppingDot(userPath), namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustomization := CreateKustomize(cluster.Name, namespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return ClusterAutomation{}, err
	}

	return ClusterAutomation{
		GitOpsRuntime: AutomationManifest{
			Path:    systemQualifiedPath(RuntimePath),
			Content: runtimeManifests,
		},
		AppCRD: AutomationManifest{
			Path:    systemQualifiedPath(AppCRDPath),
			Content: appCRDManifest,
		},
		WegoAppManifest: AutomationManifest{
			Path:    systemQualifiedPath(WegoAppPath),
			Content: wegoAppManifest,
		},
		SourceManifest: AutomationManifest{
			Path:    systemQualifiedPath(SourcePath),
			Content: sourceManifest,
		},
		SystemKustResourceManifest: AutomationManifest{
			Path:    systemQualifiedPath(SystemKustResourcePath),
			Content: systemKustResourceManifest,
		},
		UserKustResourceManifest: AutomationManifest{
			Path:    systemQualifiedPath(UserKustResourcePath),
			Content: userKustResourceManifest,
		},
		SystemKustomizationManifest: AutomationManifest{
			Path:    systemQualifiedPath(SystemKustomizationPath),
			Content: systemKustomizationManifest,
		},
	}, nil
}

func (ca ClusterAutomation) BootstrapManifests() []AutomationManifest {
	return append([]AutomationManifest{ca.AppCRD}, ca.WegoAppManifest, ca.SourceManifest, ca.SystemKustResourceManifest, ca.UserKustResourceManifest)
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
