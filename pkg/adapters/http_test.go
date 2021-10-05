package adapters_test

import (
	"errors"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

const BaseURI = "https://weave.works/api"

func TestRetrieveTemplates(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, templates []templates.Template, err error)
	}{
		{
			name:      "templates returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/templates.json")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.ElementsMatch(t, ts, []templates.Template{
					{
						Name:        "cluster-template",
						Description: "this is test template 1",
					},
					{
						Name:        "cluster-template-2",
						Description: "this is test template 2",
					},
					{
						Name:        "cluster-template-3",
						Description: "this is test template 3",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "unable to GET templates from \"https://weave.works/api/v1/templates\": Get \"https://weave.works/api/v1/templates\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, ts []templates.Template, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/templates", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			ts, err := r.RetrieveTemplates()
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRetrieveTemplateParameters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, templates []templates.TemplateParameter, err error)
	}{
		{
			name:      "template parameters returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/template_parameters.json")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.ElementsMatch(t, ts, []templates.TemplateParameter{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
						Options:     []string{"option1", "option2"},
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "unable to GET template parameters from \"https://weave.works/api/v1/templates/cluster-template/params\": Get \"https://weave.works/api/v1/templates/cluster-template/params\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, ts []templates.TemplateParameter, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/templates/cluster-template/params\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/templates/cluster-template/params", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			ts, err := r.RetrieveTemplateParameters("cluster-template")
			tt.assertFunc(t, ts, err)
		})
	}
}

func TestRenderTemplateWithParameters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "rendered template returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/rendered_template.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, `apiVersion: cluster.x-k8s.io/v1alpha4
kind: Cluster
metadata:
  name: dev
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
    kind: AWSManagedControlPlane
    name: dev-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
    kind: AWSManagedCluster
    name: dev

---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSManagedCluster
metadata:
  name: dev

---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha4
kind: AWSManagedControlPlane
metadata:
  name: dev-control-plane
spec:
  region: us-east-1
  sshKeyName: ssh_key
  version: "1.19"

---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha4
kind: AWSFargateProfile
metadata:
  name: dev-fargate-0
spec:
  clusterName: mb-test-1
  selectors:
  - namespace: default
`)
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("./testdata/service_error.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST parameters and render template from \"https://weave.works/api/v1/templates/cluster-template/render\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to POST parameters and render template from \"https://weave.works/api/v1/templates/cluster-template/render\": Post \"https://weave.works/api/v1/templates/cluster-template/render\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/templates/cluster-template/render\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", BaseURI+"/v1/templates/cluster-template/render", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			result, err := r.RenderTemplateWithParameters("cluster-template", nil, templates.Credentials{})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestCreatePullRequestFromTemplate(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, result string, err error)
	}{
		{
			name:      "pull request created",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/pull_request_created.json")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.Equal(t, result, "https://github.com/org/repo/pull/1")
			},
		},
		{
			name:      "service error",
			responder: httpmock.NewJsonResponderOrPanic(500, httpmock.File("./testdata/service_error.json")),
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
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for POST \"https://weave.works/api/v1/clusters\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("POST", BaseURI+"/v1/clusters", tt.responder)

			c, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			result, err := c.CreatePullRequestFromTemplate(templates.CreatePullRequestFromTemplateParams{})
			tt.assertFunc(t, result, err)
		})
	}
}

func TestRetrieveCredentials(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, credentials []templates.Credentials, err error)
	}{
		{
			name:      "credentials returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/credentials.json")),
			assertFunc: func(t *testing.T, creds []templates.Credentials, err error) {
				assert.ElementsMatch(t, creds, []templates.Credentials{
					{
						Group:     "infrastructure.cluster.x-k8s.io",
						Version:   "v1alpha3",
						Kind:      "AWSClusterStaticIdentity",
						Name:      "aws-creds",
						Namespace: "default",
					},
					{
						Group:     "infrastructure.cluster.x-k8s.io",
						Version:   "v1alpha4",
						Kind:      "AzureClusterIdentity",
						Name:      "azure-creds",
						Namespace: "default",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, creds []templates.Credentials, err error) {
				assert.EqualError(t, err, "unable to GET credentials from \"https://weave.works/api/v1/credentials\": Get \"https://weave.works/api/v1/credentials\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, creds []templates.Credentials, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/credentials\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/credentials", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			creds, err := r.RetrieveCredentials()
			tt.assertFunc(t, creds, err)
		})
	}
}

func TestRetrieveCredentialsByName(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, credentials templates.Credentials, err error)
	}{
		{
			name:      "credentials returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/credentials.json")),
			assertFunc: func(t *testing.T, creds templates.Credentials, err error) {
				assert.Equal(t, creds, templates.Credentials{
					Group:     "infrastructure.cluster.x-k8s.io",
					Version:   "v1alpha3",
					Kind:      "AWSClusterStaticIdentity",
					Name:      "aws-creds",
					Namespace: "default",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/credentials", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			creds, err := r.RetrieveCredentialsByName("aws-creds")
			tt.assertFunc(t, creds, err)
		})
	}
}
