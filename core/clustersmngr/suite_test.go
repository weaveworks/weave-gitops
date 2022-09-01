package clustersmngr_test

import (
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
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

func makeLeafCluster(t *testing.T, name string) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name:        name,
		Server:      k8sEnv.Rest.Host,
		BearerToken: k8sEnv.Rest.BearerToken,
		TLSConfig:   k8sEnv.Rest.TLSClientConfig,
	}
}

func makeUnreachableLeafCluster(t *testing.T, name string) clustersmngr.Cluster {
	c := makeLeafCluster(t, name)
	// hopefully no k8s server is listening here
	// FIXME: better addresses?
	c.Server = "0.0.0.0:65535"

	return c
}
