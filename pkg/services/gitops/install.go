package gitops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type InstallParams struct {
	Namespace    string
	DryRun       bool
	AppConfigURL string
}

func (g *Gitops) Install(params InstallParams) ([]byte, error) {
	ctx := context.Background()
	status := g.kube.GetClusterStatus(ctx)

	switch status {
	case kube.FluxInstalled:
		return nil, errors.New("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	case kube.Unknown:
		return nil, errors.New("Weave GitOps cannot talk to the cluster")
	}

	// TODO apply these manifests instead of generating them again
	var fluxManifests []byte
	var err error
	if params.AppConfigURL != "" {
		fluxManifests, err = g.flux.Install(params.Namespace, true)
		if err != nil {
			return fluxManifests, fmt.Errorf("error on flux install %s", err)
		}
	}
	_, err = g.flux.Install(params.Namespace, params.DryRun)
	if err != nil {
		return fluxManifests, fmt.Errorf("error on flux install %s", err)
	}
	systemManifests := make(map[string][]byte, 3)
	systemManifests["gitops-runtime.yaml"] = fluxManifests

	systemManifests["wego-system.yaml"] = manifests.AppCRD
	if params.DryRun {
		fluxManifests = append(fluxManifests, manifests.AppCRD...)
	} else {

		for _, manifest := range manifests.Manifests {
			if err := g.kube.Apply(ctx, manifest, params.Namespace); err != nil {
				return nil, fmt.Errorf("could not apply manifest: %w", err)
			}
		}

		version := version.Version
		if os.Getenv("IS_TEST_ENV") != "" {
			version = "latest"
		}

		wegoAppDeploymentManifest, err := manifests.GenerateWegoAppDeploymentManifest(version)
		if err != nil {
			return nil, fmt.Errorf("error generating wego-app deployment, %w", err)
		}
		systemManifests["wego-app.yaml"] = wegoAppDeploymentManifest
		if err := g.kube.Apply(ctx, wegoAppDeploymentManifest, params.Namespace); err != nil {
			return nil, fmt.Errorf("could not apply wego-app deployment manifest: %w", err)
		}
	}
	if params.AppConfigURL != "" {
		if err := g.storeManifests(params, systemManifests); err != nil {
			return nil, fmt.Errorf("failed to store cluster manifests: %v", err)
		}
	}

	return fluxManifests, nil
}

func (g *Gitops) storeManifests(params InstallParams, systemManifests map[string][]byte) error {
	configBranch, err := g.gitProvider.GetDefaultBranch(params.AppConfigURL)
	ctx := context.Background()
	if err != nil {
		return fmt.Errorf("could not determine default branch for config repository: %v %w", params.AppConfigURL, err)
	}
	cname, err := g.kube.GetClusterName(ctx)
	if err != nil {
		g.logger.Warningf("Cluster name not found, using default : %v", err)
		cname = "default"
	}
	normalizedURL, err := gitproviders.NewNormalizedRepoURL(params.AppConfigURL)
	if err != nil {
		return fmt.Errorf("failed to normalize URL %s: %w", params.AppConfigURL, err)
	}
	// TODO: pass context
	// TODO: actual cluster name

	if g.gitClient == nil {
		authsvc, err := apputils.GetAuthService(ctx, normalizedURL, params.DryRun)
		if err != nil {
			return fmt.Errorf("failed to create auth service for repo %s : %w", params.AppConfigURL, err)
		}
		g.gitClient, err = authsvc.CreateGitClient(ctx, normalizedURL, cname, params.Namespace)
		if err != nil {
			return fmt.Errorf("failed to create git client for repo %s : %w", params.AppConfigURL, err)
		}

	}

	remover, err := app.CloneRepo(g.gitClient, params.AppConfigURL, configBranch, params.DryRun)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()
	// TODO Map (file name, manifest)
	manifests := make(map[string][]byte, 3)
	// TODO use path separator
	clusterPath := ".weave-gitops/clusters/" + cname

	gitsource, sourceName, err := g.genSource(cname, configBranch, params)
	if err != nil {
		return fmt.Errorf("failed to create source manifest: %w", err)
	}
	manifests["flux-source-resource.yaml"] = gitsource
	system, err := g.genKustomize(fmt.Sprintf("%s-system", cname), sourceName, configBranch, "."+clusterPath+"/system", params)
	if err != nil {
		return fmt.Errorf("failed to create system kustomization manifest: %w", err)
	}
	manifests["flux-system-kustomization-resource.yaml"] = system
	user, err := g.genKustomize(fmt.Sprintf("%s-user", cname), sourceName, configBranch, "."+clusterPath+"/user", params)
	if err != nil {
		return fmt.Errorf("failed to create user kustomization manifest: %w", err)
	}
	manifests["flux-user-kustomization-resource.yaml"] = user

	if !params.DryRun {
		// if !params.AutoMerge {
		// 	if err := a.createPullRequestToRepo(info, info.Spec.ConfigURL, appHash, appSpec, appGoat, appSource); err != nil {
		// 		return err
		// 	}
		// } else {
		g.logger.Actionf("Writing manifests to disk")
		if err := g.writeManifestsToGit(filepath.Join(clusterPath, "system"), manifests); err != nil {
			return fmt.Errorf("failed to write manifests: %w", err)
		}
		if err := g.writeManifestsToGit(filepath.Join(clusterPath, "system"), systemManifests); err != nil {
			return fmt.Errorf("failed to write system manifests: %w", err)
		}
		// store a .keep file in the user dir
		userKeep := map[string][]byte{
			".keep": strconv.AppendQuote(nil, "# keep"),
		}
		if err := g.writeManifestsToGit(filepath.Join(clusterPath, "user"), userKeep); err != nil {
			return fmt.Errorf("failed to write user manifests: %w", err)
		}
		// if err := a.writeAppYaml(info, appSpec); err != nil {
		// 	return fmt.Errorf("failed writing app.yaml to disk: %w", err)
		// }

	}

	g.logger.Actionf("Applying manifests to the cluster")
	// only apply the system manfiests as the others will get picked up once flux is running
	if err := g.applyManifestsToK8s(ctx, params.Namespace, manifests); err != nil {
		return fmt.Errorf("failed applying system manifests to cluster %s :%v", cname, err)
	}
	// if err := a.applyToCluster(info, params.DryRun, extRepoMan.source, extRepoMan.target, extRepoMan.appDir); err != nil {
	// 	return fmt.Errorf("could not apply manifests to the cluster: %w", err)
	// }

	return app.CommitAndPush(g.gitClient, "Add GitOps runtime manifests", params.DryRun, g.logger)
}

func (g *Gitops) genSource(cname, branch string, params InstallParams) ([]byte, string, error) {
	secretRef := utils.CreateRepoSecretName(cname, params.AppConfigURL)

	sourceManifest, err := g.flux.CreateSourceGit(secretRef, params.AppConfigURL, branch, secretRef, params.Namespace)
	if err != nil {
		return nil, secretRef, fmt.Errorf("could not create git source for repo %s : %w", params.AppConfigURL, err)
	}

	return sourceManifest, secretRef, nil

}
func (g *Gitops) genKustomize(name, cname, branch, path string, params InstallParams) ([]byte, error) {
	sourceManifest, err := g.flux.CreateKustomization(name, cname, path, params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not create flux kustomization for path %s : %v", path, err)
	}

	return sourceManifest, nil

}
func (g *Gitops) writeManifestsToGit(path string, manifests map[string][]byte) error {
	for k, m := range manifests {
		if err := g.gitClient.Write(filepath.Join(path, k), m); err != nil {
			g.logger.Warningf("failed to write manfiest %s : %v", k, err)
			return err
		}
	}
	return nil
}
func (g *Gitops) applyManifestsToK8s(ctx context.Context, namespace string, manifests map[string][]byte) error {
	for k, manifest := range manifests {
		g.logger.Infow("apply manifests", k, manifest)
		if err := g.kube.Apply(ctx, manifest, namespace); err != nil {
			return fmt.Errorf("could not apply manifest %s : %w", k, err)
		}

	}
	return nil
}
