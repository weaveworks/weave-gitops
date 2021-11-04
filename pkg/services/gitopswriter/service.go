package gitopswriter

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	"sigs.k8s.io/yaml"
)

const (
	AddCommitMessage    = "Add application manifests"
	RemoveCommitMessage = "Remove application manifests"
)

var _ GitOpsDirectoryWriter = &GitOpsDirectoryWriterSvc{}

type GitOpsDirectoryWriter interface {
	AddApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error
	RemoveApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error
}

type GitOpsDirectoryWriterSvc struct {
	Automation automation.AutomationService
	RepoWriter gitrepo.RepoWriter
	Osys       osys.Osys
	Logger     logger.Logger
}

func NewGitOpsDirectoryWriter(automationSvc automation.AutomationService, repoWriter gitrepo.RepoWriter, osys osys.Osys, logger logger.Logger) GitOpsDirectoryWriter {
	return &GitOpsDirectoryWriterSvc{Automation: automationSvc, RepoWriter: repoWriter, Osys: osys, Logger: logger}
}

func (dw *GitOpsDirectoryWriterSvc) AddApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error {
	manifests, err := dw.Automation.GenerateAutomation(ctx, app, clusterName)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	remover, repoDir, err := dw.RepoWriter.CloneRepo(ctx, app.Branch)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	defer remover()

	kManifest, err := addKustomizeResources(app, repoDir, clusterName, appKustomizeReference(app))
	if err != nil {
		return err
	}

	manifests = append(manifests, kManifest)

	if !autoMerge {
		files := []gitprovider.CommitFile{}

		for _, manifest := range manifests {
			content := string(manifest.Content)

			files = append(files, gitprovider.CommitFile{Path: &manifest.Path, Content: &content})
		}

		defaultBranch, err := dw.RepoWriter.GetDefaultBranch(ctx)
		if err != nil {
			return fmt.Errorf("failed to retrieve default branch for repository: %w", err)
		}

		prApp := gitproviders.PullRequestInfo{
			Title:         fmt.Sprintf("Gitops add %s", app.Name),
			Description:   fmt.Sprintf("Added yamls for %s", app.Name),
			CommitMessage: AddCommitMessage,
			TargetBranch:  defaultBranch,
			NewBranch:     automation.GetAppHash(app),
			Files:         files,
		}

		if err := dw.RepoWriter.CreatePullRequest(ctx, prApp); err != nil {
			return fmt.Errorf("failed creating pull request: %w", err)
		}
	} else {
		if err := dw.RepoWriter.WriteAndMerge(ctx, repoDir, AddCommitMessage, manifests); err != nil {
			return fmt.Errorf("failed writing automation to disk: %w", err)
		}
	}

	return nil
}

func (dw *GitOpsDirectoryWriterSvc) RemoveApplication(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error {
	branch, err := dw.RepoWriter.GetDefaultBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain config branch: %w", err)
	}

	remover, repoDir, err := dw.RepoWriter.CloneRepo(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	dw.Logger.Actionf("Removing application from cluster and repository")

	appDir := filepath.Join(repoDir, automation.AppYamlDir(app))

	resourcePaths, err := dw.Osys.ReadDir(appDir)
	if err != nil {
		return fmt.Errorf("failed to read resource files: %w", err)
	}

	for _, resourcePath := range resourcePaths {
		if !utils.Exists(filepath.Join(repoDir)) {
			continue
		}

		if err := dw.RepoWriter.Remove(ctx, resourcePath.Name()); err != nil {
			return fmt.Errorf("failed to remove app resource from repository: %w", err)
		}
	}

	// Remove reference in kustomization file
	kManifest, err := removeKustomizeResources(app, repoDir, clusterName, appKustomizeReference(app))
	if err != nil {
		return fmt.Errorf("failed to remove app reference from user kustomize file: %w", err)
	}

	if err = dw.RepoWriter.Write(ctx, kManifest.Path, kManifest.Content); err != nil {
		return fmt.Errorf("failed to write updated kustomize file: %w", err)
	}

	return dw.RepoWriter.CommitAndPush(ctx, RemoveCommitMessage)
}

func addKustomizeResources(app models.Application, repoDir, clusterName string, resources ...string) (automation.AutomationManifest, error) {
	userKustomizationRepoPath := filepath.Join(git.WegoRoot, git.WegoClusterDir, clusterName, "user", "kustomization.yaml")
	userKustomization := filepath.Join(repoDir, userKustomizationRepoPath)

	k, err := automation.GetOrCreateKustomize(userKustomization, app.Name, app.Namespace)
	if err != nil {
		return automation.AutomationManifest{}, err
	}

	k.Resources = append(k.Resources, resources...)

	userKustomizationManifest, err := yaml.Marshal(&k)
	if err != nil {
		return automation.AutomationManifest{}, fmt.Errorf("failed to marshal kustomize %v : %w", k, err)
	}

	return automation.AutomationManifest{
		Path:    userKustomizationRepoPath,
		Content: userKustomizationManifest,
	}, nil
}

func removeKustomizeResources(app models.Application, repoDir, clusterName string, resources ...string) (automation.AutomationManifest, error) {
	userKustomizationRepoPath := filepath.Join(git.WegoRoot, git.WegoClusterDir, clusterName, "user", "kustomization.yaml")
	userKustomization := filepath.Join(repoDir, userKustomizationRepoPath)

	k, err := automation.GetOrCreateKustomize(userKustomization, app.Name, app.Namespace)
	if err != nil {
		return automation.AutomationManifest{}, err
	}

	oldResources := k.Resources
	newResources := []string{}

	for _, oldResource := range oldResources {
		keep := true

		for _, resource := range resources {
			if resource == oldResource {
				keep = false
				break
			}
		}

		if keep {
			newResources = append(newResources, oldResource)
		}
	}

	k.Resources = newResources

	userKustomizationManifest, err := yaml.Marshal(&k)
	if err != nil {
		return automation.AutomationManifest{}, fmt.Errorf("failed to marshal kustomize %v : %w", k, err)
	}

	return automation.AutomationManifest{
		Path:    userKustomizationRepoPath,
		Content: userKustomizationManifest,
	}, nil
}

func appKustomizeReference(app models.Application) string {
	return "../../../apps/" + automation.AppDeployName(app)
}
