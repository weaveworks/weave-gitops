package adapters

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/weave-gitops/pkg/capi"
)

const (
	expiredHeaderName = "Entitlement-Expired-Message"
)

// An HTTP client of the cluster service.
type HttpClient struct {
	baseURI *url.URL
	client  *resty.Client
}

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewHttpClient creates a new HTTP client of the cluster service.
// The endpoint is expected to be an absolute HTTP URI.
func NewHttpClient(endpoint string, client *resty.Client, out io.Writer) (*HttpClient, error) {
	u, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}

	client = client.SetHostURL(u.String()).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			if m := r.Header().Get(expiredHeaderName); m != "" {
				fmt.Fprintln(out, m)
			}
			return nil
		})

	return &HttpClient{
		baseURI: u,
		client:  client,
	}, nil
}

// Source returns the endpoint of the cluster service.
func (c *HttpClient) Source() string {
	return c.baseURI.String()
}

// RetrieveTemplates returns the list of all templates from the cluster service.
func (c *HttpClient) RetrieveTemplates() ([]capi.Template, error) {
	endpoint := "v1/templates"

	type ListTemplatesResponse struct {
		Templates []*capi.Template
	}

	var templateList ListTemplatesResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&templateList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET templates from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var ts []capi.Template
	for _, t := range templateList.Templates {
		ts = append(ts, capi.Template{
			Name:        t.Name,
			Description: t.Description,
		})
	}

	return ts, nil
}

// RetrieveTemplateParameters returns the list of all parameters of the
// specified template.
func (c *HttpClient) RetrieveTemplateParameters(name string) ([]capi.TemplateParameter, error) {
	endpoint := "v1/templates/{name}/params"

	type ListTemplateParametersResponse struct {
		Parameters []*capi.TemplateParameter
	}

	var templateParametersList ListTemplateParametersResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetResult(&templateParametersList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET template parameters from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var tps []capi.TemplateParameter
	for _, p := range templateParametersList.Parameters {
		tps = append(tps, capi.TemplateParameter{
			Name:        p.Name,
			Description: p.Description,
			Required:    p.Required,
			Options:     p.Options,
		})
	}

	return tps, nil
}

// RenderTemplateWithParameters returns a YAML representation of the specified
// template populated with the supplied parameters.
func (c *HttpClient) RenderTemplateWithParameters(name string, parameters map[string]string, creds capi.Credentials) (string, error) {
	endpoint := "v1/templates/{name}/render"

	// POST request payload
	type TemplateParameterValuesAndCredentials struct {
		Values      map[string]string `json:"values"`
		Credentials capi.Credentials  `json:"credentials"`
	}

	// POST response payload
	type RenderedTemplate struct {
		Template string `json:"renderedTemplate"`
	}

	var renderedTemplate RenderedTemplate

	var serviceErr *ServiceError

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetBody(TemplateParameterValuesAndCredentials{Values: parameters, Credentials: creds}).
		SetResult(&renderedTemplate).
		SetError(&serviceErr).
		Post(endpoint)

	if serviceErr != nil {
		return "", fmt.Errorf("unable to POST parameters and render template from %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return "", fmt.Errorf("unable to POST parameters and render template from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for POST %q was %d", res.Request.URL, res.StatusCode())
	}

	return renderedTemplate.Template, nil
}

// CreatePullRequestFromTemplate commits the YAML template to the specified
// branch and creates a pull request of that branch.
func (c *HttpClient) CreatePullRequestFromTemplate(params capi.CreatePullRequestFromTemplateParams) (string, error) {
	endpoint := "v1/clusters"

	// POST request payload
	type CreatePullRequestFromTemplateRequest struct {
		RepositoryURL   string            `json:"repositoryUrl"`
		HeadBranch      string            `json:"headBranch"`
		BaseBranch      string            `json:"baseBranch"`
		Title           string            `json:"title"`
		Description     string            `json:"description"`
		TemplateName    string            `json:"templateName"`
		ParameterValues map[string]string `json:"parameter_values"`
		CommitMessage   string            `json:"commitMessage"`
		Credentials     capi.Credentials  `json:"credentials"`
	}

	// POST response payload
	type CreatePullRequestFromTemplateResponse struct {
		WebURL string `json:"webUrl"`
	}

	var result CreatePullRequestFromTemplateResponse

	var serviceErr *ServiceError

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetBody(CreatePullRequestFromTemplateRequest{
			RepositoryURL:   params.RepositoryURL,
			HeadBranch:      params.HeadBranch,
			BaseBranch:      params.BaseBranch,
			Title:           params.Title,
			Description:     params.Description,
			TemplateName:    params.TemplateName,
			ParameterValues: params.ParameterValues,
			CommitMessage:   params.CommitMessage,
			Credentials:     params.Credentials,
		}).
		SetResult(&result).
		SetError(&serviceErr).
		Post(endpoint)

	if serviceErr != nil {
		return "", fmt.Errorf("unable to POST template and create pull request to %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return "", fmt.Errorf("unable to POST template and create pull request to %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for POST %q was %d", res.Request.URL, res.StatusCode())
	}

	return result.WebURL, nil
}

// RetrieveCredentials returns a list of all CAPI credentials.
func (c *HttpClient) RetrieveCredentials() ([]capi.Credentials, error) {
	endpoint := "v1/credentials"

	type ListCredentialsResponse struct {
		Credentials []*capi.Credentials
		Total       int32
	}

	var credentialsList ListCredentialsResponse

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&credentialsList).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET credentials from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var creds []capi.Credentials
	for _, c := range credentialsList.Credentials {
		creds = append(creds, capi.Credentials{
			Group:     c.Group,
			Version:   c.Version,
			Kind:      c.Kind,
			Name:      c.Name,
			Namespace: c.Namespace,
		})
	}

	return creds, nil
}

// RetrieveCredentialsByName returns a specific set of CAPI credentials.
func (c *HttpClient) RetrieveCredentialsByName(name string) (capi.Credentials, error) {
	var creds capi.Credentials

	credsList, err := c.RetrieveCredentials()
	if err != nil {
		return creds, fmt.Errorf("unable to retrieve credentials from %q: %w", c.Source(), err)
	}

	for _, c := range credsList {
		if c.Name == name {
			creds = capi.Credentials{
				Group:     c.Group,
				Version:   c.Version,
				Kind:      c.Kind,
				Name:      c.Name,
				Namespace: c.Namespace,
			}
		}
	}

	return creds, nil
}
