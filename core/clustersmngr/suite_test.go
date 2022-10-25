package clustersmngr_test

import (
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/clusters"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"k8s.io/client-go/rest"
)

var k8sEnv *testutils.K8sTestEnv

func TestMain(m *testing.M) {
	var err error
	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})

	if err != nil {
		panic(err)
	}

	code := m.Run()

	k8sEnv.Stop()

	os.Exit(code)
}

func makeLeafCluster(t *testing.T, name string) clusters.Cluster {
	cluster, err := clusters.NewSingleCluster(name, k8sEnv.Rest, nil)
	if err != nil {
		t.Error("Expected err to be nil, got", err)
	}
	return cluster
}

func makeUnreachableLeafCluster(t *testing.T, name string) clusters.Cluster {
	c := rest.CopyConfig(k8sEnv.Rest)

	// hopefully no k8s server is listening here
	// FIXME: better addresses?
	c.Host = "0.0.0.0:65535"

	cluster, err := clusters.NewSingleCluster(name, c, nil)
	if err != nil {
		t.Error("Expected err to be nil, got", err)
	}
	return cluster
}
