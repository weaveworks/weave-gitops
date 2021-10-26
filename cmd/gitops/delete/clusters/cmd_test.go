package clusters_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/weaveworks/weave-gitops/cmd/gitops/delete/clusters"
)

func TestClusterCommand_URL(t *testing.T) {
	e := "http://localhost"
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	cmd := clusters.ClusterCommand(&e, client)
	cmd.SetArgs([]string{"cluster-name"})
	actual := cmd.Execute()
	expected := errors.New("repository url is required")

	if actual == nil || !strings.Contains(actual.Error(), expected.Error()) {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
}
