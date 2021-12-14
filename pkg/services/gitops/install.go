package gitops

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kyaml "sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
)

// InstallParams are used to configure the installation the GitOps runtime.
type InstallParams struct {
	Namespace         string
	DryRun            bool
	ConfigRepo        string
	FluxHTTPSUsername string
	FluxHTTPSPassword string
}

func (g *Gitops) Install(params InstallParams) (map[string][]byte, error) {
	ctx := context.Background()

	// Avoid outputting anything before this line as it will mess up the
	// output from validate in case of a dry-run.
	// If you must, output something to Debug which can be disabled.
	if err := g.validateWegoInstall(ctx, params); err != nil {
		return nil, err
	}

	// TODO apply these manifests instead of generating them again
	var fluxManifests []byte

	var err error

	if params.ConfigRepo != "" || params.DryRun {
		// We need to get the manifests to persist in the repo and
		// non-dry run install doesn't return them
		fluxManifests, err = g.flux.Install(params.Namespace, true)
		if err != nil {
			return nil, fmt.Errorf("error on flux install %w", err)
		}
	}

	if !params.DryRun {
		_, err = g.flux.Install(params.Namespace, false)
		if err != nil {
			return nil, fmt.Errorf("error on flux install %w", err)
		}
	}

	systemManifests := make(map[string][]byte)
	systemManifests["gitops-runtime.yaml"] = fluxManifests
	systemManifests["wego-system.yaml"] = manifests.AppCRD

	if params.DryRun {
		return systemManifests, nil
	} else {
		if err := g.kube.Apply(ctx, manifests.AppCRD, params.Namespace); err != nil {
			return nil, fmt.Errorf("could not apply App CRD: %w", err)
		}

		version := version.Version
		if os.Getenv("IS_TEST_ENV") != "" {
			version = "latest"
		}

		wegoAppManifests, err := manifests.GenerateManifests(manifests.Params{
			AppVersion: version,
			Namespace:  params.Namespace,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating wego-app manifests, %w", err)
		}
		for _, m := range wegoAppManifests {
			if err := g.kube.Apply(ctx, m, params.Namespace); err != nil {
				return nil, fmt.Errorf("error applying wego-app manifest \n%s: %w", m, err)
			}
		}

		systemManifests["wego-app.yaml"] = bytes.Join(wegoAppManifests, []byte("---\n"))
	}

	wegoConfigCM, err := g.saveWegoConfig(ctx, params)
	if err != nil {
		return nil, err
	}

	configBytes, err := kyaml.Marshal(wegoConfigCM)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	systemManifests["wego-config.yaml"] = configBytes

	return systemManifests, nil
}

func (g *Gitops) StoreManifests(gitClient git.Git, gitProvider gitproviders.GitProvider, params InstallParams, systemManifests map[string][]byte) (map[string][]byte, error) {
	ctx := context.Background()

	if !params.DryRun && params.ConfigRepo != "" {
		cname, err := g.kube.GetClusterName(ctx)
		if err != nil {
			g.logger.Warningf("Cluster name not found, using default : %v", err)

			cname = "default"
		}

		goatManifests, err := g.storeManifests(gitClient, gitProvider, params, systemManifests, cname)
		if err != nil {
			return nil, fmt.Errorf("failed to store cluster manifests: %w", err)
		}

		g.logger.Actionf("Applying manifests to the cluster")
		// only apply the system manifests as the others will get picked up once flux is running
		if err := g.applyManifestsToK8s(ctx, params.Namespace, goatManifests); err != nil {
			return nil, fmt.Errorf("failed applying system manifests to cluster %s :%w", cname, err)
		}
	}

	return systemManifests, nil
}

func (g *Gitops) validateWegoInstall(ctx context.Context, params InstallParams) error {
	status := g.kube.GetClusterStatus(ctx)

	switch status {
	case kube.FluxInstalled:
		return errors.New("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	case kube.Unknown:
		return errors.New("Weave GitOps cannot talk to the cluster")
	}

	wegoConfig, err := g.kube.GetWegoConfig(ctx, "")
	if err != nil {
		if !errors.Is(err, kube.ErrWegoConfigNotFound) {
			return fmt.Errorf("failed getting wego config: %w", err)
		}
	}

	if wegoConfig.WegoNamespace != "" && wegoConfig.WegoNamespace != params.Namespace {
		return errors.New("You cannot install Weave GitOps into a different namespace")
	}

	return nil
}

func (g *Gitops) storeManifests(gitClient git.Git, gitProvider gitproviders.GitProvider, params InstallParams, systemManifests map[string][]byte, cname string) (map[string][]byte, error) {
	ctx := context.Background()

	normalizedURL, err := gitproviders.NewRepoURL(params.ConfigRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert app config repo %q : %w", params.ConfigRepo, err)
	}

	configBranch, err := gitProvider.GetDefaultBranch(ctx, normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("could not determine default branch for config repository: %q %w", params.ConfigRepo, err)
	}

	remover, _, err := gitrepo.CloneRepo(ctx, gitClient, normalizedURL, configBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to clone configuration repo to store manifests: %w", err)
	}

	defer remover()

	manifests := make(map[string][]byte, 3)
	clusterPath := filepath.Join(git.WegoRoot, git.WegoClusterDir, cname)

	gitsource, sourceName, err := g.genSource(configBranch, params.Namespace, normalizedURL, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create source manifest: %w", err)
	}
	// filepath.Join doesn't support a starting "."
	prefixForFlux := func(s string) string {
		return "." + s
	}
	manifests["flux-source-resource.yaml"] = gitsource

	system, err := g.genKustomize(automation.ConstrainResourceName(fmt.Sprintf("%s-system", cname)), sourceName,
		prefixForFlux(filepath.Join(".", clusterPath, git.WegoClusterOSWorkloadDir)), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create system kustomization manifest: %w", err)
	}

	manifests["flux-system-kustomization-resource.yaml"] = system

	user, err := g.genKustomize(automation.ConstrainResourceName(fmt.Sprintf("%s-user", cname)), sourceName,
		prefixForFlux(filepath.Join(".", clusterPath, git.WegoClusterUserWorkloadDir)), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user kustomization manifest: %w", err)
	}
	manifests["flux-user-kustomization-resource.yaml"] = user

	g.logger.Actionf("Writing manifests to disk")
	if err := g.writeManifestsToGit(gitClient, filepath.Join(clusterPath, "system"), manifests); err != nil {
		return nil, fmt.Errorf("failed to write manifests: %w", err)
	}

	if err := g.writeManifestsToGit(gitClient, filepath.Join(clusterPath, "system"), systemManifests); err != nil {
		return nil, fmt.Errorf("failed to write system manifests: %w", err)
	}
	// store a .keep file in the user dir
	userKeep := map[string][]byte{
		".keep": strconv.AppendQuote(nil, "# keep"),
	}
	if err := g.writeManifestsToGit(gitClient, filepath.Join(clusterPath, "user"), userKeep); err != nil {
		return nil, fmt.Errorf("failed to write user manifests: %w", err)
	}

	return manifests, gitrepo.CommitAndPush(ctx, gitClient, "Add GitOps runtime manifests", g.logger)
}

func (g *Gitops) genSource(branch string, namespace string, normalizedUrl gitproviders.RepoURL, params InstallParams) ([]byte, string, error) {
	secretRef := automation.CreateRepoSecretName(normalizedUrl).String()

	sourceManifest, err := g.flux.CreateSourceGit(secretRef, normalizedUrl, branch, secretRef, namespace, credsFromParams(params))
	if err != nil {
		return nil, secretRef, fmt.Errorf("could not create git source for repo %s : %w", normalizedUrl.String(), err)
	}

	return sourceManifest, secretRef, nil
}

func (g *Gitops) genKustomize(name, cname, path string, params InstallParams) ([]byte, error) {
	sourceManifest, err := g.flux.CreateKustomization(name, cname, path, params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not create flux kustomization for path %q : %w", path, err)
	}

	return sourceManifest, nil
}

func (g *Gitops) writeManifestsToGit(gitClient git.Git, path string, manifests map[string][]byte) error {
	for k, m := range manifests {
		if err := gitClient.Write(filepath.Join(path, k), m); err != nil {
			g.logger.Warningf("failed to write manifest %s : %v", k, err)
			return err
		}
	}

	return nil
}

func (g *Gitops) applyManifestsToK8s(ctx context.Context, namespace string, manifests map[string][]byte) error {
	for k, manifest := range manifests {
		if err := g.kube.Apply(ctx, manifest, namespace); err != nil {
			return fmt.Errorf("could not apply manifest %q : %w", k, err)
		}
	}

	return nil
}

const LabelPartOf = "app.kubernetes.io/part-of"

var ErrNamespaceNotFound = errors.New("namespace not found")

func (g *Gitops) fetchNamespaceWithLabel(ctx context.Context, key string, value string) (string, error) {
	selector := labels.NewSelector()

	partOf, err := labels.NewRequirement(key, selection.Equals, []string{value})
	if err != nil {
		return "", fmt.Errorf("bad requirement: %w", err)
	}

	selector = selector.Add(*partOf)

	options := client.ListOptions{
		LabelSelector: selector,
	}

	nsl := &corev1.NamespaceList{}
	if err := g.kube.Raw().List(ctx, nsl, &options); err != nil {
		return "", fmt.Errorf("error setting resource: %w", err)
	}

	namespaces := []string{}
	for _, n := range nsl.Items {
		namespaces = append(namespaces, n.Name)
	}

	if len(namespaces) == 0 {
		return "", ErrNamespaceNotFound
	}

	if len(namespaces) > 1 {
		return "", fmt.Errorf("found multiple namespaces %s with %s=%s, we are unable to define the correct one", namespaces, key, value)
	}

	return namespaces[0], nil
}

func (g *Gitops) saveWegoConfig(ctx context.Context, params InstallParams) (*corev1.ConfigMap, error) {
	fluxNamespace, err := g.fetchNamespaceWithLabel(ctx, LabelPartOf, "flux")
	if err != nil {
		if !errors.Is(err, ErrNamespaceNotFound) {
			return nil, fmt.Errorf("failed fetching flux namespace: %w", err)
		}
	}

	cm, err := g.kube.SetWegoConfig(ctx, kube.WegoConfig{
		FluxNamespace: fluxNamespace,
		WegoNamespace: params.Namespace,
	}, params.Namespace)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func credsFromParams(p InstallParams) *flux.HTTPSCreds {
	var creds *flux.HTTPSCreds
	if p.FluxHTTPSUsername != "" || p.FluxHTTPSPassword != "" {
		creds = &flux.HTTPSCreds{
			Username: p.FluxHTTPSUsername,
			Password: p.FluxHTTPSPassword,
		}
	}
	return creds
}
