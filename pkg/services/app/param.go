package app

import (
	"fmt"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/services/app/internal"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"net/url"
)

type AddParams struct {
	Dir                        string
	Name                       string
	Url                        string
	Path                       string
	Branch                     string
	PrivateKey                 string
	DeploymentType             wego.DeploymentType
	Chart                      string
	SourceType                 wego.SourceType
	AppConfigUrl               string
	Namespace                  string
	DryRun                     bool
	AutoMerge                  bool
	GitProviderToken           string
	HelmReleaseTargetNamespace string
}

func (p AddParams) IsHelmChart() bool {
	return p.Chart != ""
}

func (p AddParams) SetDefaultValues(provider gitproviders.GitProvider) (AddParams, error) {
	p.SourceType = wego.SourceTypeGit

	// making sure the config url is in good format
	if IsExternalConfigUrl(p.AppConfigUrl) {
		p.AppConfigUrl = utils.SanitizeRepoUrl(p.AppConfigUrl)
	}

	if p.IsHelmChart() {
		p.SourceType = wego.SourceTypeHelm
		p.DeploymentType = wego.DeploymentTypeHelm
		p.Path = p.Chart

		if p.Name == "" {
			if internal.NameTooLong(p.Chart) {
				return AddParams{}, fmt.Errorf("chart name %q is too long to use as application name; please specify name with '--name'", p.Chart)
			}

			p.Name = p.Chart
		}

		if p.Url == "" {
			return AddParams{}, fmt.Errorf("--url must be specified for helm repositories")
		}
	} else {
		// making sure url is in the correct format
		_, err := url.Parse(p.Url)
		if err != nil {
			return p, fmt.Errorf("error validating url %w", err)
		}

		p.Url = utils.SanitizeRepoUrl(p.Url)

		// resetting Dir param since Url has priority over it
		p.Dir = ""
	}

	if p.Name == "" {
		repoName := utils.UrlToRepoName(p.Url)
		if internal.NameTooLong(repoName) {
			return p, fmt.Errorf("url base name %q is too long to use as application name; please specify name with '--name'", repoName)
		}

		p.Name = internal.GenerateResourceName(p.Url)
	}

	if p.Path == "" {
		p.Path = DefaultPath
	}

	if p.DeploymentType == "" {
		p.DeploymentType = DefaultDeploymentType
	}

	if p.Branch == "" {
		p.Branch = DefaultBranch

		if p.SourceType == wego.SourceTypeGit {
			branch, err := provider.GetDefaultBranch(p.Url)
			if err != nil {
				return p, err
			} else {
				p.Branch = branch
			}
		}
	}

	if internal.NameTooLong(p.Name) {
		return p, fmt.Errorf("application name too long: %s; must be <= 63 characters", p.Name)
	}

	// Validate namespace argument for helm
	if p.HelmReleaseTargetNamespace != "" {
		if nserr := utils.ValidateNamespace(p.HelmReleaseTargetNamespace); nserr != nil {
			return p, nserr
		}
	}

	return p, nil
}
