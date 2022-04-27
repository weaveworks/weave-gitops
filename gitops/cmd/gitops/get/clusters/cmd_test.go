package clusters_test

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/gitops/cmd/gitops/root"
)

func TestGetCluster(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		status    int
		response  interface{}
		args      []string
		result    string
		errString string
	}{
		{
			name:     "cluster kubeconfig",
			url:      "http://localhost:8000/v1/clusters/dev-cluster/kubeconfig",
			status:   http.StatusOK,
			response: httpmock.File("../../../../pkg/adapters/testdata/cluster_kubeconfig.json"),
			args: []string{
				"get", "cluster",
				"dev-cluster",
				"--kubeconfig",
				"--endpoint", "http://localhost:8000",
			},
		},
		{
			name: "http error",
			args: []string{
				"get", "cluster",
				"dev-cluster",
				"--kubeconfig",
				"--endpoint", "not_a_valid_url",
			},
			errString: "parse \"not_a_valid_url\": invalid URI for request",
		},
		{
			name: "no endpoint",
			args: []string{
				"get", "cluster",
				"dev-cluster",
				"--kubeconfig",
			},
			errString: "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()
			httpmock.RegisterResponder(
				http.MethodGet,
				tt.url,
				func(r *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(tt.status, tt.response)
				},
			)

			cmd := root.RootCmd(client)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.errString == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.errString)
			}
		})
	}
}
