package adapters

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
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
		fmt.Println(err)
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

	type ServiceError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
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
