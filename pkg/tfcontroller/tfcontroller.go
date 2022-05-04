package tfcontroller

import (
	"fmt"
	"io"
)

type CreatePullRequestFromTemplateParams struct {
	GitProviderToken string
	TemplateName     string
	ParameterValues  map[string]string
	RepositoryURL    string
	HeadBranch       string
	BaseBranch       string
	Title            string
	Description      string
	CommitMessage    string
}

// TemplatePullRequester defines the interface that adapters
// need to implement in order to create a pull request from
// a Terraform template. Implementers should return the web URI of
// the pull request.
type TemplatePullRequester interface {
	CreatePullRequestFromTFControllerTemplate(params CreatePullRequestFromTemplateParams) (string, error)
}

// TODO: extract this to a template pkg
// CreatePullRequestFromTFControllerTemplate uses a TemplatePullRequester
// adapter to create a pull request from a CAPI template.
func CreatePullRequestFromTFControllerTemplate(params CreatePullRequestFromTemplateParams, r TemplatePullRequester, w io.Writer) error {
	res, err := r.CreatePullRequestFromTFControllerTemplate(params)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	fmt.Fprintf(w, "Created pull request: %s\n", res)

	return nil
}
