package profiles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AddOptions struct {
	Name       string
	Cluster    string
	ConfigRepo string
	Version    string
	Port       string
	Namespace  string
	AutoMerge  bool
	Logger     logger.Logger
}

// Add adds a new profile to the cluster
func (s *ProfilesSvc) Add(ctx context.Context, gitProvider gitproviders.GitProvider, opts AddOptions) error {
	validatedOps, err := validateAddOptions(opts)
	if err != nil {
		return err
	}

	configRepoUrl, err := gitproviders.NewRepoURL(opts.ConfigRepo)
	if err != nil {
		return err
	}

	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoUrl)
	if err != nil {
		return err
	} else if !repoExists {
		return fmt.Errorf("repository '%v' could not be found", configRepoUrl.String())
	}

	_, err = gitProvider.GetRepoFiles(ctx, configRepoUrl, git.GetSystemPath(opts.Cluster), "")
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' for config repository '%s': %s", git.GetSystemPath(opts.Cluster), configRepoUrl.String(), err)
	}

	profilesList, err := doKubeProfilesGetRequest(ctx, opts.Namespace, wegoServiceName, opts.Port, getProfilesPath, s.ClientSet)
	if err != nil {
		return err
	}

	_, err = getAvailableProfile(profilesList, opts)
	if err != nil {
		return err
	}

	// makeHelmReleaseName := func(clusterName, installName string) string {
	// 	return clusterName + "-" + installName
	// }

	_ = &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			ReleaseName:     opts.Name,
			TargetNamespace: opts.Namespace,
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:       "https://my-chart",
					Version:     "v1.2.3",
					ValuesFiles: []string{"file-1.yaml"},
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: opts.Name,
					},
				},
			},
		},
	}

	printAddSummary(validatedOps)
	return nil
}

func validateAddOptions(opts AddOptions) (AddOptions, error) {
	if opts.ConfigRepo == "" {
		return opts, errors.New("--config-repo should be provided")
	}

	if opts.Name == "" {
		return opts, errors.New("--name should be provided")
	}

	if opts.Cluster == "" {
		return opts, errors.New("--cluster should be provided")
	}

	if automation.ApplicationNameTooLong(opts.Name) {
		return opts, fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			opts.Name, automation.MaxKubernetesResourceNameLength)
	}

	if strings.HasPrefix(opts.Name, "wego") {
		return opts, fmt.Errorf("the prefix 'wego' is used by weave gitops and is not allowed for a profile name")
	}

	return opts, nil
}

func printAddSummary(opts AddOptions) {
	opts.Logger.Println("Adding profile:\n")
	opts.Logger.Println("Name: %s", opts.Name)
	opts.Logger.Println("Version: %s", opts.Version)
	opts.Logger.Printf("Cluster: %s", opts.Cluster)

	opts.Logger.Println("")
}

func getAvailableProfile(profilesList *profiles.GetProfilesResponse, opts AddOptions) (*profiles.Profile, error) {
	var availableProfile *profiles.Profile
	for _, p := range profilesList.Profiles {
		if p.Name == opts.Name {
			if len(p.AvailableVersions) == 0 {
				return nil, fmt.Errorf("no available version found for profile '%s' in %s/%s", p.Name, opts.Cluster, opts.Namespace)
			}
			availableProfile = p
			break
		}
	}
	if availableProfile == nil {
		return nil, fmt.Errorf("no available profile '%s' found in %s/%s", opts.Name, opts.Cluster, opts.Namespace)
	}
	return availableProfile, nil
}
