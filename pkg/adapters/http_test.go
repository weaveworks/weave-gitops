package adapters_test

import (
	"errors"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
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

func TestRetrieveClusters(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, cs []clusters.Cluster, err error)
	}{
		{
			name:      "clusters returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/clusters.json")),
			assertFunc: func(t *testing.T, cs []clusters.Cluster, err error) {
				assert.ElementsMatch(t, cs, []clusters.Cluster{
					{
						Name:   "cluster-a",
						Status: "pullRequestCreated",
					},
					{
						Name:   "cluster-b",
						Status: "pullRequestCreated",
					},
					{
						Name:   "cluster-c",
						Status: "pullRequestCreated",
					},
				})
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, cs []clusters.Cluster, err error) {
				assert.EqualError(t, err, "unable to GET clusters from \"https://weave.works/api/gitops/api/clusters\": Get \"https://weave.works/api/gitops/api/clusters\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, cs []clusters.Cluster, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/gitops/api/clusters\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/gitops/api/clusters", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			cs, err := r.RetrieveClusters()
			tt.assertFunc(t, cs, err)
		})
	}
}

func TestGetClusterKubeconfig(t *testing.T) {
	tests := []struct {
		name       string
		responder  httpmock.Responder
		assertFunc func(t *testing.T, s string, err error)
	}{
		{
			name:      "kubeconfig returned",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/cluster_kubeconfig.json")),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.YAMLEq(t, s, httpmock.File("./testdata/cluster_kubeconfig.yaml").String())
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.EqualError(t, err, "unable to GET cluster kubeconfig from \"https://weave.works/api/v1/clusters/dev/kubeconfig\": Get \"https://weave.works/api/v1/clusters/dev/kubeconfig\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.EqualError(t, err, "response status for GET \"https://weave.works/api/v1/clusters/dev/kubeconfig\" was 400")
			},
		},
		{
			name:      "base64 decode failure",
			responder: httpmock.NewJsonResponderOrPanic(200, httpmock.File("./testdata/cluster_kubeconfig_decode_failure.json")),
			assertFunc: func(t *testing.T, s string, err error) {
				assert.EqualError(t, err, "unable to base64 decode the cluster kubeconfig: illegal base64 data at input byte 3")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("GET", BaseURI+"/v1/clusters/dev/kubeconfig", tt.responder)

			r, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			k, err := r.GetClusterKubeconfig("dev")
			tt.assertFunc(t, k, err)
		})
	}
}

func TestDeleteClusters(t *testing.T) {
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
				assert.EqualError(t, err, "unable to Delete cluster and create pull request to \"https://weave.works/api/v1/clusters\": something bad happened")
			},
		},
		{
			name:      "error returned",
			responder: httpmock.NewErrorResponder(errors.New("oops")),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "unable to Delete cluster and create pull request to \"https://weave.works/api/v1/clusters\": Delete \"https://weave.works/api/v1/clusters\": oops")
			},
		},
		{
			name:      "unexpected status code",
			responder: httpmock.NewStringResponder(400, ""),
			assertFunc: func(t *testing.T, result string, err error) {
				assert.EqualError(t, err, "response status for Delete \"https://weave.works/api/v1/clusters\" was 400")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder("DELETE", BaseURI+"/v1/clusters", tt.responder)

			c, err := adapters.NewHttpClient(BaseURI, client, os.Stdout)
			assert.NoError(t, err)
			result, err := c.DeleteClusters(clusters.DeleteClustersParams{})
			tt.assertFunc(t, result, err)
		})
	}
}
