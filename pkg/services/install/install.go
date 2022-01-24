package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
)

type Installer interface {
	Install(namespace string, configURL gitproviders.RepoURL, autoMerge bool) error
}

type install struct {
	fluxClient        flux.Flux
	kubeClient        kube.Kube
	gitClient         git.Git
	gitProviderClient gitproviders.GitProvider
	log               logger.Logger
	repoWriter        gitopswriter.RepoWriter
}

// NewInstaller instantiate a new installer
func NewInstaller(fluxClient flux.Flux, kubeClient kube.Kube, gitClient git.Git, gitProviderClient gitproviders.GitProvider, log logger.Logger, repoWriter gitopswriter.RepoWriter) Installer {
	return &install{
		fluxClient:        fluxClient,
		kubeClient:        kubeClient,
		gitClient:         gitClient,
		gitProviderClient: gitProviderClient,
		log:               log,
		repoWriter:        repoWriter,
	}
}

// Install generates gitops manifests, save them to the config repository and applies them to the cluster. In case auto-merge is true it creates a PR instead of writing directly to the default branch.
func (i *install) Install(namespace string, configURL gitproviders.RepoURL, autoMerge bool) error {
	ctx := context.Background()

	if err := validateWegoInstall(ctx, i.kubeClient, namespace); err != nil {
		return fmt.Errorf("failed validating wego installation: %w", err)
	}

	clusterName, err := i.kubeClient.GetClusterName(ctx)
	if err != nil {
		return fmt.Errorf("failed getting cluster name: %w", err)
	}

	if _, err = i.fluxClient.Install(namespace, false); err != nil {
		return fmt.Errorf("failed installing flux: %w", err)
	}

	fluxNamespace, err := i.kubeClient.FetchNamespaceWithLabel(ctx, flux.PartOfLabelKey, flux.PartOfLabelValue)
	if err != nil {
		if !errors.Is(err, kube.ErrNamespaceNotFound) {
			return fmt.Errorf("failed fetching flux namespace: %w", err)
		}
	}

	bootstrapManifests, err := models.BootstrapManifests(i.fluxClient, models.BootstrapManifestsParams{
		ClusterName:   clusterName,
		WegoNamespace: namespace,
		FluxNamespace: fluxNamespace.Name,
		ConfigRepo:    configURL,
	})
	if err != nil {
		return fmt.Errorf("failed getting bootstrap manifests: %w", err)
	}

	defaultBranch, err := i.gitProviderClient.GetDefaultBranch(ctx, configURL)
	if err != nil {
		return fmt.Errorf("failed getting default branch: %w", err)
	}

	for _, manifest := range bootstrapManifests {
		ms := bytes.Split(manifest.Content, []byte("---\n"))

		for _, m := range ms {
			if len(bytes.Trim(m, " \t\n")) == 0 {
				continue
			}

			if err := i.kubeClient.Apply(ctx, m, namespace); err != nil {
				return fmt.Errorf("error applying manifest %s: %w", manifest.Path, err)
			}
		}
	}

	gitopsManifests, err := models.GitopsManifests(ctx, bootstrapManifests, models.GitopsManifestsParams{
		FluxClient:    i.fluxClient,
		GitProvider:   i.gitProviderClient,
		ClusterName:   clusterName,
		WegoNamespace: namespace,
		ConfigRepo:    configURL,
	})
	if err != nil {
		return fmt.Errorf("failed generating gitops manifests: %w", err)
	}

	i.log.Actionf("Associating cluster %q", clusterName)

	if autoMerge {
		err = i.repoWriter.Write(ctx, configURL, defaultBranch, models.ConvertManifestsToCommitFiles(gitopsManifests))
		if err != nil {
			return fmt.Errorf("failed writting to default branch %w", err)
		}

		return nil
	}

	pullRequestInfo := gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("GitOps associate %s", clusterName),
		Description:   fmt.Sprintf("Add gitops automation manifests for cluster %s", clusterName),
		CommitMessage: gitopswriter.ClusterCommitMessage,
		NewBranch:     models.GetClusterHash(clusterName),
		TargetBranch:  defaultBranch,
		Files:         models.ConvertManifestsToCommitFiles(gitopsManifests),
	}

	pr, err := i.gitProviderClient.CreatePullRequest(ctx, configURL, pullRequestInfo)
	if err != nil {
		return fmt.Errorf("failed creating pull request: %w", err)
	}

	i.log.Println("Pull Request created: %s\n", pr.Get().WebURL)

	return nil
}

func validateWegoInstall(ctx context.Context, kubeClient kube.Kube, namespace string) error {
	status := kubeClient.GetClusterStatus(ctx)

	switch status {
	case kube.FluxInstalled:
		return errors.New("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	case kube.Unknown:
		return errors.New("Weave GitOps cannot talk to the cluster")
	}

	wegoConfig, err := kubeClient.GetWegoConfig(ctx, "")
	if err != nil {
		if !errors.Is(err, kube.ErrWegoConfigNotFound) {
			return fmt.Errorf("failed getting wego config: %w", err)
		}
	}

	if wegoConfig.WegoNamespace != "" && wegoConfig.WegoNamespace != namespace {
		return errors.New("You cannot install Weave GitOps into a different namespace")
	}

	return nil
}
