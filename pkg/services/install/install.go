package install

import (
	"bytes"
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/osys"

	"github.com/weaveworks/weave-gitops/pkg/logger"

	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
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
	osysClient        osys.Osys
}

func NewInstaller(fluxClient flux.Flux, kubeClient kube.Kube, gitClient git.Git, gitProviderClient gitproviders.GitProvider, log logger.Logger, osysClient osys.Osys) Installer {
	return &Install{
		fluxClient:        fluxClient,
		kubeClient:        kubeClient,
		gitClient:         gitClient,
		gitProviderClient: gitProviderClient,
		log:               log,
		osysClient:        osysClient,
	}
}

func (i *Install) Install(namespace string, configURL gitproviders.RepoURL, autoMerge bool) error {
	ctx := context.Background()

	clusterName, err := i.kubeClient.GetClusterName(ctx)
	if err != nil {
		return fmt.Errorf("failed getting cluster name: %w", err)
	}

	manifests, err := automation.BootstrapManifests(i.fluxClient, clusterName, namespace, configURL)
	if err != nil {
		return fmt.Errorf("failed getting gitops manifests: %w", err)
	}

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

	// This might be better outside this Install function
	// What if we want to write to a different place
	cluster := models.Cluster{Name: clusterName}
	repoWriter := gitrepo.NewRepoWriter(configURL, i.gitProviderClient, i.gitClient, i.log)
	automationGen := automation.NewAutomationGenerator(i.gitProviderClient, i.fluxClient, i.log)
	gitOpsDirWriter := gitopswriter.NewGitOpsDirectoryWriter(automationGen, repoWriter, i.osysClient, i.log)
	err = gitOpsDirWriter.AssociateCluster(ctx, cluster, configURL, namespace, namespace, autoMerge)
	if err != nil {
		return fmt.Errorf("failed associating cluster: %w", err)
	}

	return nil
}
