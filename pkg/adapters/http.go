package adapters

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
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
func (c *HttpClient) RetrieveTemplates() ([]templates.Template, error) {
	endpoint := "v1/templates"

	type ListTemplatesResponse struct {
		Templates []*templates.Template
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

	var ts []templates.Template
	for _, t := range templateList.Templates {
		ts = append(ts, templates.Template{
			Name:        t.Name,
			Description: t.Description,
		})
	}

	return ts, nil
}

// RetrieveTemplateParameters returns the list of all parameters of the
// specified template.
func (c *HttpClient) RetrieveTemplateParameters(name string) ([]templates.TemplateParameter, error) {
	endpoint := "v1/templates/{name}/params"

	type ListTemplateParametersResponse struct {
		Parameters []*templates.TemplateParameter
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

	var tps []templates.TemplateParameter
	for _, p := range templateParametersList.Parameters {
		tps = append(tps, templates.TemplateParameter{
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
func (c *HttpClient) RenderTemplateWithParameters(name string, parameters map[string]string, creds templates.Credentials) (string, error) {
	endpoint := "v1/templates/{name}/render"

	// POST request payload
	type TemplateParameterValuesAndCredentials struct {
		Values      map[string]string     `json:"values"`
		Credentials templates.Credentials `json:"credentials"`
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
func (c *HttpClient) CreatePullRequestFromTemplate(params templates.CreatePullRequestFromTemplateParams) (string, error) {
	endpoint := "v1/clusters"

	// POST request payload
	type CreatePullRequestFromTemplateRequest struct {
		RepositoryURL   string                `json:"repositoryUrl"`
		HeadBranch      string                `json:"headBranch"`
		BaseBranch      string                `json:"baseBranch"`
		Title           string                `json:"title"`
		Description     string                `json:"description"`
		TemplateName    string                `json:"templateName"`
		ParameterValues map[string]string     `json:"parameter_values"`
		CommitMessage   string                `json:"commitMessage"`
		Credentials     templates.Credentials `json:"credentials"`
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
func (c *HttpClient) RetrieveCredentials() ([]templates.Credentials, error) {
	endpoint := "v1/credentials"

	type ListCredentialsResponse struct {
		Credentials []*templates.Credentials
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

	var creds []templates.Credentials
	for _, c := range credentialsList.Credentials {
		creds = append(creds, templates.Credentials{
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
func (c *HttpClient) RetrieveCredentialsByName(name string) (templates.Credentials, error) {
	var creds templates.Credentials

	credsList, err := c.RetrieveCredentials()
	if err != nil {
		return creds, fmt.Errorf("unable to retrieve credentials from %q: %w", c.Source(), err)
	}

	for _, c := range credsList {
		if c.Name == name {
			creds = templates.Credentials{
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

// RetrieveClusters returns the list of all clusters from the cluster service.
func (c *HttpClient) RetrieveClusters() ([]clusters.Cluster, error) {
	endpoint := "gitops/api/clusters"

	type ClusterView struct {
		Name            string `json:"name"`
		Status          string `json:"status"`
		PullRequestType string `json:"pr-type"`
	}

	type ClustersResponse struct {
		Clusters []ClusterView `json:"clusters"`
	}

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetDoNotParseResponse(true).
		Get(endpoint)

	if err != nil {
		return nil, fmt.Errorf("unable to GET clusters from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	var clusterList ClustersResponse

	err = json.NewDecoder(res.RawBody()).Decode(&clusterList)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response as a cluster list: %w", err)
	}

	var cs []clusters.Cluster
	for _, c := range clusterList.Clusters {
		cs = append(cs, clusters.Cluster{
			Name:            c.Name,
			Status:          c.Status,
			PullRequestType: c.PullRequestType,
		})
	}

	return cs, nil
}

type GetKubeconfigResponse struct {
	Kubeconfig string
}

func (c *HttpClient) GetClusterKubeconfig(name string) (string, error) {
	endpoint := "v1/clusters/{name}/kubeconfig"

	var result GetKubeconfigResponse
	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetPathParams(map[string]string{
			"name": name,
		}).
		SetResult(&result).
		Get(endpoint)

	if err != nil {
		return "", fmt.Errorf("unable to GET cluster kubeconfig from %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for GET %q was %d", res.Request.URL, res.StatusCode())
	}

	b, err := base64.StdEncoding.DecodeString(result.Kubeconfig)
	if err != nil {
		return "", fmt.Errorf("unable to base64 decode the cluster kubeconfig: %w", err)
	}

	return string(b), nil
}

// DeleteClusters deletes CAPI cluster using its name
func (c *HttpClient) DeleteClusters(params clusters.DeleteClustersParams) (string, error) {
	endpoint := "v1/clusters"

	type DeleteClustersResponse struct {
		WebURL string `json:"webUrl"`
	}

	var result DeleteClustersResponse

	var serviceErr *ServiceError

	res, err := c.client.R().
		SetHeader("Accept", "application/json").
		SetBody(DeleteClustersPullRequestRequest{
			HeadBranch:    params.HeadBranch,
			BaseBranch:    params.BaseBranch,
			Title:         params.Title,
			Description:   params.Description,
			ClusterNames:  params.ClustersNames,
			CommitMessage: params.CommitMessage,
		}).
		SetResult(&result).
		SetError(&serviceErr).
		Delete(endpoint)

	if serviceErr != nil {
		return "", fmt.Errorf("unable to Delete cluster and create pull request to %q: %s", res.Request.URL, serviceErr.Message)
	}

	if err != nil {
		return "", fmt.Errorf("unable to Delete cluster and create pull request to %q: %w", res.Request.URL, err)
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("response status for Delete %q was %d", res.Request.URL, res.StatusCode())
	}

	return result.WebURL, nil
}

type Credential struct {
	Group     string
	Version   string
	Kind      string
	Name      string
	Namespace string
}

type DeleteClustersPullRequestRequest struct {
	RepositoryUrl string
	HeadBranch    string
	BaseBranch    string
	Title         string
	Description   string
	ClusterNames  []string
	CommitMessage string
	Credentials   Credential
}
