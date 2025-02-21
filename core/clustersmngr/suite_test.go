package clustersmngr_test

import (
	"os"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var k8sEnv *testutils.K8sTestEnv

func TestMain(m *testing.M) {
	var err error
	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../tools/testcrds",
	})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	k8sEnv.Stop()

	os.Exit(code)
}

func makeLeafCluster(t *testing.T, name string) cluster.Cluster {
	t.Helper()
	cluster, err := cluster.NewSingleCluster(name, k8sEnv.Rest, nil, kube.UserPrefixes{})
	if err != nil {
		t.Error("Expected err to be nil, got", err)
	}

	return cluster
}

func makeUnreachableLeafCluster(t *testing.T, name string) cluster.Cluster {
	t.Helper()
	c := rest.CopyConfig(k8sEnv.Rest)

	// hopefully no k8s server is listening here
	// FIXME: better addresses?
	c.Host = "0.0.0.0:65535"

	cluster, err := cluster.NewSingleCluster(name, c, nil, kube.UserPrefixes{})
	if err != nil {
		t.Error("Expected err to be nil, got", err)
	}

	return cluster
}
