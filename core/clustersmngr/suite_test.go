package clustersmngr_test

import (
	"os"
	"testing"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var k8sEnv *testutils.K8sTestEnv

func TestMain(m *testing.M) {
	os.Setenv("KUBEBUILDER_ASSETS", "../../tools/bin/envtest")

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

func makeLeafCluster(t *testing.T) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name:        "leaf-cluster",
		Server:      k8sEnv.Rest.Host,
		BearerToken: k8sEnv.Rest.BearerToken,
		TLSConfig:   k8sEnv.Rest.TLSClientConfig,
	}
}
