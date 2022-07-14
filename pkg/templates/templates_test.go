package templates_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestCreatePullRequestFromTemplate_CAPI(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "pull request created",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("../adapters/testdata/pull_request_created.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, "https://github.com/org/repo/pull/1")
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("../adapters/testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/clusters\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/clusters\": Post \"https://weave.works/api/v1/clusters\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/clusters\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", testutils.BaseURI+"/v1/clusters", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)

			result, err := client.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{TemplateKind: templates.CAPITemplateKind.String()})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestCreatePullRequestFromTemplate_Terraform(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "pull request created",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("../adapters/testdata/pull_request_created.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, "https://github.com/org/repo/pull/1")
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("../adapters/testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/tfcontrollers\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST template and create pull request to \"https://weave.works/api/v1/tfcontrollers\": Post \"https://weave.works/api/v1/tfcontrollers\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(http.StatusBadRequest, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/tfcontrollers\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &config.Options{
				Endpoint: testutils.BaseURI,
			}
			client := adapters.NewHTTPClient()
			httpmock.ActivateNonDefault(client.GetBaseClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", testutils.BaseURI+"/v1/tfcontrollers", tt.responder)

			err := client.ConfigureClientWithOptions(opts, os.Stdout)
			assert.NoError(t, err)
			result, err := client.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{TemplateKind: templates.GitOpsTemplateKind.String()})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestGetTemplates(t *testing.T) {
	tests := []struct {
		name             string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No templates were found.\n",
		},
		{
			name: "templates includes just name",
			ts: []templates.Template{
				{
					Name:     "template-a",
					Provider: "aws",
				},
				{
					Name: "template-b",
				},
			},
			expected: "NAME\tPROVIDER\tDESCRIPTION\tERROR\ntemplate-a\taws\t\t\ntemplate-b\t\t\t\n",
		},
		{
			name: "templates include all fields",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Description: "a desc",
					Provider:    "azure",
					Error:       "",
				},
				{
					Name:        "template-b",
					Description: "b desc",
					Error:       "something went wrong",
				},
			},
			expected: "NAME\tPROVIDER\tDESCRIPTION\tERROR\ntemplate-a\tazure\ta desc\t\ntemplate-b\t\tb desc\tsomething went wrong\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(tt.ts, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplates(templates.CAPITemplateKind, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name             string
		tmplName         string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			tmplName: "",
			expected: "No templates were found.\n",
		},
		{
			name:     "templates includes just name",
			tmplName: "template-a",
			ts: []templates.Template{
				{
					Name:     "template-a",
					Provider: "aws",
				},
			},
			expected: "NAME\tPROVIDER\tDESCRIPTION\tERROR\ntemplate-a\taws\t\t\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(tt.ts, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplates(templates.CAPITemplateKind, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplatesByProvider(t *testing.T) {
	tests := []struct {
		name             string
		provider         string
		ts               []templates.Template
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			provider: "aws",
			expected: "No templates were found for provider \"aws\".\n",
		},
		{
			name:     "templates includes just name",
			provider: "aws",
			ts: []templates.Template{
				{
					Name:     "template-a",
					Provider: "aws",
				},
				{
					Name:     "template-b",
					Provider: "aws",
				},
			},
			expected: "NAME\tPROVIDER\tDESCRIPTION\tERROR\ntemplate-a\taws\t\t\ntemplate-b\taws\t\t\n",
		},
		{
			name:     "templates include all fields",
			provider: "azure",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Provider:    "azure",
					Description: "a desc",
					Error:       "",
				},
				{
					Name:        "template-b",
					Provider:    "azure",
					Description: "b desc",
					Error:       "something went wrong",
				},
			},
			expected: "NAME\tPROVIDER\tDESCRIPTION\tERROR\ntemplate-a\tazure\ta desc\t\ntemplate-b\tazure\tb desc\tsomething went wrong\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(tt.ts, nil, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplatesByProvider(templates.CAPITemplateKind, tt.provider, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplateParameters(t *testing.T) {
	tests := []struct {
		name             string
		tps              []templates.TemplateParameter
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no templates",
			expected: "No template parameters were found.\n",
		},
		{
			name: "template parameters include just name",
			tps: []templates.TemplateParameter{
				{
					Name:     "template-param-a",
					Required: true,
				},
				{
					Name: "template-param-b",
				},
			},
			expected: "NAME\tREQUIRED\tDESCRIPTION\tOPTIONS\ntemplate-param-a\ttrue\ntemplate-param-b\tfalse\n",
		},
		{
			name: "templates include all fields",
			tps: []templates.TemplateParameter{
				{
					Name:        "template-param-a",
					Required:    true,
					Description: "a desc",
					Options:     []string{"op-1", "op-2"},
				},
				{
					Name:        "template-param-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tREQUIRED\tDESCRIPTION\tOPTIONS\ntemplate-param-a\ttrue\ta desc\top-1, op-2\ntemplate-param-b\tfalse\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve parameters for template \"foo\" from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, tt.tps, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplateParameters(templates.CAPITemplateKind, "foo", c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no result returned",
			expected: "No template was found.\n",
		},
		{
			name:             "error returned",
			err:              errors.New("expected param CLUSTER_NAME to be passed"),
			expectedErrorStr: "unable to render template \"foo\": expected param CLUSTER_NAME to be passed",
		},
		{
			name: "result is rendered to output",
			result: `apiVersion: cluster.x-k8s.io/v1alpha3
				kind: Cluster
				metadata:
					name: foo
				spec:
					clusterNetwork:
					pods:
						cidrBlocks:
						- 192.168.0.0/16
					controlPlaneRef:
					apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
					kind: KubeadmControlPlane
					name: foo-control-plane
					infrastructureRef:
					apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
					kind: AWSCluster
					name: foo`,
			expected: `apiVersion: cluster.x-k8s.io/v1alpha3
				kind: Cluster
				metadata:
					name: foo
				spec:
					clusterNetwork:
					pods:
						cidrBlocks:
						- 192.168.0.0/16
					controlPlaneRef:
					apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
					kind: KubeadmControlPlane
					name: foo-control-plane
					infrastructureRef:
					apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
					kind: AWSCluster
					name: foo`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.RenderTemplateWithParameters(templates.CAPITemplateKind, "foo", nil, templates.Credentials{}, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestCreatePullRequest(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:             "error returned",
			err:              errors.New("something went wrong"),
			expectedErrorStr: "unable to create pull request: something went wrong",
		},
		{
			name:     "pull request created",
			result:   "https://github.com/org/repo/pull/1",
			expected: "Created pull request: https://github.com/org/repo/pull/1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{}, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	tests := []struct {
		name             string
		creds            []templates.Credentials
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no credentials",
			expected: "No credentials were found.\n",
		},
		{
			name: "credentials found",
			creds: []templates.Credentials{
				{
					Name: "creds-a",
					Kind: "AWSCluster",
				},
				{
					Name: "creds-b",
					Kind: "AzureCluster",
				},
			},
			expected: "NAME\tINFRASTRUCTURE PROVIDER\ncreds-a\tAWS\ncreds-b\tAzure\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve credentials from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, tt.creds, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetCredentials(c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetTemplateProfiles(t *testing.T) {
	tests := []struct {
		name             string
		fs               []templates.Profile
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no profiles",
			expected: "No template profiles were found.\n",
		},
		{
			name: "profiles includes just name",
			fs: []templates.Profile{
				{
					Name:              "profile-a",
					AvailableVersions: []string{"v0.0.15"},
				},
				{
					Name: "profile-b",
				},
			},
			expected: "NAME\tLATEST_VERSIONS\nprofile-a\tv0.0.15\nprofile-b\t\n",
		},
		{
			name: "profiles include more than 5 versions",
			fs: []templates.Profile{
				{
					Name:              "profile-a",
					AvailableVersions: []string{"v0.0.9", "v0.0.10", "v0.0.11", "v0.0.12", "v0.0.13", "v0.0.14", "v0.0.15"},
				},
				{
					Name:              "profile-b",
					AvailableVersions: []string{"v0.0.13", "v0.0.14", "v0.0.15"},
				},
			},
			expected: "NAME\tLATEST_VERSIONS\nprofile-a\tv0.0.11, v0.0.12, v0.0.13, v0.0.14, v0.0.15\nprofile-b\tv0.0.13, v0.0.14, v0.0.15\n",
		},
		{
			name:             "error retrieving profiles",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve profiles for template \"profile-b\" from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeClient(nil, nil, nil, tt.fs, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplateProfiles("profile-b", c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type fakeClient struct {
	ts  []templates.Template
	ps  []templates.TemplateParameter
	cs  []templates.Credentials
	fs  []templates.Profile
	s   string
	err error
}

func newFakeClient(ts []templates.Template, ps []templates.TemplateParameter, cs []templates.Credentials, fs []templates.Profile, s string, err error) *fakeClient {
	return &fakeClient{
		ts:  ts,
		ps:  ps,
		cs:  cs,
		fs:  fs,
		s:   s,
		err: err,
	}
}

func (c *fakeClient) Source() string {
	return "In-memory fake"
}

func (c *fakeClient) RetrieveTemplates(kind templates.TemplateKind) ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *fakeClient) RetrieveTemplate(name string, kind templates.TemplateKind) (*templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	if c.ts[0].Name == name {
		return &c.ts[0], nil
	}

	return nil, errors.New("not found")
}

func (c *fakeClient) RetrieveTemplatesByProvider(kind templates.TemplateKind, provider string) ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *fakeClient) RetrieveTemplateParameters(kind templates.TemplateKind, name string) ([]templates.TemplateParameter, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ps, nil
}

func (c *fakeClient) RenderTemplateWithParameters(kind templates.TemplateKind, name string, parameters map[string]string, creds templates.Credentials) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *fakeClient) CreatePullRequestFromTemplate(params templates.CreatePullRequestFromTemplateParams) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *fakeClient) RetrieveCredentials() ([]templates.Credentials, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.cs, nil
}

func (c *fakeClient) RetrieveTemplateProfiles(name string) ([]templates.Profile, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.fs, nil
}
