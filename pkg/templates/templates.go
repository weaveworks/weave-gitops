package templates

import (
	"fmt"
	"io"

	"github.com/weaveworks/weave-gitops/pkg/capi"
)

const (
	// CAPI template
	CAPITemplateKind = "CAPITemplate"

	// TF template
	GitOpsTemplateKind = "GitOpsTemplate"
)

type CreatePullRequestFromTemplateParams struct {
	GitProviderToken string
	TemplateName     string
	TemplateKind     string
	ParameterValues  map[string]string
	RepositoryURL    string
	HeadBranch       string
	BaseBranch       string
	Title            string
	Description      string
	CommitMessage    string
	Credentials      capi.Credentials
	ProfileValues    []capi.ProfileValues
}

// TemplatePullRequester defines the interface that adapters
// need to implement in order to create a pull request from
// a template (e.g. CAPI template, TF-Controller template).
// Implementers should return the web URI of the pull request.
type TemplatePullRequester interface {
	CreatePullRequestFromTemplate(params CreatePullRequestFromTemplateParams) (string, error)
}

// CreatePullRequestFromTemplate uses a TemplatePullRequester
// adapter to create a pull request from a template.
func CreatePullRequestFromTemplate(params CreatePullRequestFromTemplateParams, r TemplatePullRequester, w io.Writer) error {
	res, err := r.CreatePullRequestFromTemplate(params)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	fmt.Fprintf(w, "Created pull request: %s\n", res)

	return nil
}
