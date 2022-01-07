package install

import (
	"bytes"
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"

	"github.com/weaveworks/weave-gitops/pkg/models"

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
	gitOpsDirWriter   gitopswriter.GitOpsDirectoryWriter
}

func NewInstaller(fluxClient flux.Flux, kubeClient kube.Kube, gitClient git.Git, gitProviderClient gitproviders.GitProvider, log logger.Logger, osysClient osys.Osys, gitOpsDirWriter gitopswriter.GitOpsDirectoryWriter) Installer {
	return &Install{
		fluxClient:        fluxClient,
		kubeClient:        kubeClient,
		gitClient:         gitClient,
		gitProviderClient: gitProviderClient,
		log:               log,
		osysClient:        osysClient,
		gitOpsDirWriter:   gitOpsDirWriter,
	}
}

func (i *Install) Install(namespace string, configURL gitproviders.RepoURL, autoMerge bool) error {
	ctx := context.Background()

	clusterName, err := i.kubeClient.GetClusterName(ctx)
	if err != nil {
		return fmt.Errorf("failed getting cluster name: %w", err)
	}

	if _, err = i.fluxClient.Install(namespace, false); err != nil {
		return fmt.Errorf("failed intalling flux: %w", err)
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

	cluster := models.Cluster{Name: clusterName}
	err = i.gitOpsDirWriter.AssociateCluster(ctx, i.fluxClient, i.gitProviderClient, cluster, configURL, namespace, autoMerge)
	if err != nil {
		return fmt.Errorf("failed associating cluster: %w", err)
	}

	return nil
}
