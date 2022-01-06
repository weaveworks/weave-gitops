package capi_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/capi"
)

func TestGetTemplates(t *testing.T) {
	tests := []struct {
		name             string
		ts               []capi.Template
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
			ts: []capi.Template{
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
			ts: []capi.Template{
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
			err := capi.GetTemplates(c, w)
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
		ts               []capi.Template
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
			ts: []capi.Template{
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
			ts: []capi.Template{
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
			err := capi.GetTemplatesByProvider(tt.provider, c, w)
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
		tps              []capi.TemplateParameter
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
			tps: []capi.TemplateParameter{
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
			tps: []capi.TemplateParameter{
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
			err := capi.GetTemplateParameters("foo", c, w)
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
			err := capi.RenderTemplateWithParameters("foo", nil, capi.Credentials{}, c, w)
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
			err := capi.CreatePullRequestFromTemplate(capi.CreatePullRequestFromTemplateParams{}, c, w)
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
		creds            []capi.Credentials
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
			creds: []capi.Credentials{
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
			err := capi.GetCredentials(c, w)
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
		fs               []capi.Profile
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
			fs: []capi.Profile{
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
			fs: []capi.Profile{
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
			err := capi.GetTemplateProfiles("profile-b", c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type fakeClient struct {
	ts  []capi.Template
	ps  []capi.TemplateParameter
	cs  []capi.Credentials
	fs  []capi.Profile
	s   string
	err error
}

func newFakeClient(ts []capi.Template, ps []capi.TemplateParameter, cs []capi.Credentials, fs []capi.Profile, s string, err error) *fakeClient {
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

func (c *fakeClient) RetrieveTemplates() ([]capi.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *fakeClient) RetrieveTemplatesByProvider(provider string) ([]capi.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *fakeClient) RetrieveTemplateParameters(name string) ([]capi.TemplateParameter, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ps, nil
}

func (c *fakeClient) RenderTemplateWithParameters(name string, parameters map[string]string, creds capi.Credentials) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *fakeClient) CreatePullRequestFromTemplate(params capi.CreatePullRequestFromTemplateParams) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *fakeClient) RetrieveCredentials() ([]capi.Credentials, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.cs, nil
}

func (c *fakeClient) RetrieveTemplateProfiles(name string) ([]capi.Profile, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.fs, nil
}
