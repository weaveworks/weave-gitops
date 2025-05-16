package server_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/server"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestGetFeatureFlags(t *testing.T) {
	RegisterFailHandler(Fail)

	ctx := t.Context()

	featureflags.Set("this is a flag", "you won't find it anywhere else")

	scheme, err := kube.CreateScheme()
	if err != nil {
		t.Fatal(err)
	}

	cluster, err := cluster.NewSingleCluster("Default", k8sEnv.Rest, scheme, kube.UserPrefixes{})
	if err != nil {
		t.Fatal(err)
	}

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{
		fetcher.NewSingleClusterFetcher(cluster),
	}, &nsChecker, logr.Discard())

	hc := health.NewHealthChecker()

	cfg, err := server.NewCoreConfig(logr.Discard(), &rest.Config{}, "test", clustersManager, hc)
	Expect(err).NotTo(HaveOccurred())
	coreSrv, err := server.NewCoreServer(ctx, cfg)
	Expect(err).NotTo(HaveOccurred())

	resp, err := coreSrv.GetFeatureFlags(t.Context(), &pb.GetFeatureFlagsRequest{})
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.Flags).To(HaveKeyWithValue("this is a flag", "you won't find it anywhere else"))
}
