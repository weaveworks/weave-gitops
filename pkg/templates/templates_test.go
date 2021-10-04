package templates_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

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
			expected: "No templates found.\n",
		},
		{
			name: "templates includes just name",
			ts: []templates.Template{
				{
					Name: "template-a",
				},
				{
					Name: "template-b",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-a\ntemplate-b\n",
		},
		{
			name: "templates include all fields",
			ts: []templates.Template{
				{
					Name:        "template-a",
					Description: "a desc",
				},
				{
					Name:        "template-b",
					Description: "b desc",
				},
			},
			expected: "NAME\tDESCRIPTION\ntemplate-a\ta desc\ntemplate-b\tb desc\n",
		},
		{
			name:             "error retrieving templates",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve templates from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(tt.ts, nil, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplates(c, w)
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
			expected: "No template parameters were found.",
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
			c := NewFakeClient(nil, tt.tps, nil, "", tt.err)
			w := new(bytes.Buffer)
			err := templates.GetTemplateParameters("foo", c, w)
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
			expected: "No template found.",
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
			c := NewFakeClient(nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.RenderTemplateWithParameters("foo", nil, templates.Credentials{}, c, w)
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
			c := NewFakeClient(nil, nil, nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := templates.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{}, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeClient struct {
	ts  []templates.Template
	ps  []templates.TemplateParameter
	cs  []templates.Credentials
	s   string
	err error
}

func NewFakeClient(ts []templates.Template, ps []templates.TemplateParameter, cs []templates.Credentials, s string, err error) *FakeClient {
	return &FakeClient{
		ts:  ts,
		ps:  ps,
		cs:  cs,
		s:   s,
		err: err,
	}
}

func (c *FakeClient) Source() string {
	return "In-memory fake"
}

func (c *FakeClient) RetrieveTemplates() ([]templates.Template, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ts, nil
}

func (c *FakeClient) RetrieveTemplateParameters(name string) ([]templates.TemplateParameter, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.ps, nil
}

func (c *FakeClient) RenderTemplateWithParameters(name string, parameters map[string]string, creds templates.Credentials) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *FakeClient) CreatePullRequestFromTemplate(params templates.CreatePullRequestFromTemplateParams) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}
