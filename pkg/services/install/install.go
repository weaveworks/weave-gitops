package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/models"

	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"

	"github.com/weaveworks/weave-gitops/pkg/logger"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type Installer interface {
	Install(namespace string, configURL gitproviders.RepoURL, autoMerge bool) error
}

type Install struct {
	fluxClient        flux.Flux
	kubeClient        kube.Kube
	gitClient         git.Git
	gitProviderClient gitproviders.GitProvider
	log               logger.Logger
	repoWriter        gitopswriter.RepoWriter
}

func NewInstaller(fluxClient flux.Flux, kubeClient kube.Kube, gitClient git.Git, gitProviderClient gitproviders.GitProvider, log logger.Logger, repoWriter gitopswriter.RepoWriter) Installer {
	return &Install{
		fluxClient:        fluxClient,
		kubeClient:        kubeClient,
		gitClient:         gitClient,
		gitProviderClient: gitProviderClient,
		log:               log,
		repoWriter:        repoWriter,
	}
}

func (i *Install) Install(namespace string, configURL gitproviders.RepoURL, autoMerge bool) error {
	ctx := context.Background()

	if err := validateWegoInstall(ctx, i.kubeClient, namespace); err != nil {
		return fmt.Errorf("failed validating wego installation: %w", err)
	}

	clusterName, err := i.kubeClient.GetClusterName(ctx)
	if err != nil {
		return fmt.Errorf("failed getting cluster name: %w", err)
	}

	if _, err = i.fluxClient.Install(namespace, false); err != nil {
		return fmt.Errorf("failed intalling flux: %w", err)
	}

	manifests, err := models.BootstrapManifests(i.fluxClient, clusterName, namespace, configURL)
	if err != nil {
		return fmt.Errorf("failed getting bootstrap manifests: %w", err)
	}

	defaultBranch, err := i.gitProviderClient.GetDefaultBranch(ctx, configURL)
	if err != nil {
		return fmt.Errorf("failed getting default branch: %w", err)
	}

	// TODO: Move this to bootstrap after refactoring getGitClients
	source, err := models.GetSourceManifest(ctx, i.fluxClient, i.gitProviderClient, clusterName, namespace, configURL, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed getting git source: %w", err)
	}

	manifests = append(manifests, source)

	for _, manifest := range manifests {
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

	gitopsManifests, err := models.GitopsManifests(ctx, i.fluxClient, i.gitProviderClient, clusterName, namespace, configURL)
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
		Files:         models.ConvertManifestsToCommitFiles(manifests),
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
