package server_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/server"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetFeatureFlags(t *testing.T) {
	RegisterFailHandler(Fail)

	featureflags.Set("this is a flag", "you won't find it anywhere else")

	cfg := server.NewCoreConfig(logr.Discard(), &rest.Config{}, "test", &clustersmngrfakes.FakeClientsFactory{})
	k8s := fake.NewClientBuilder().Build()
	fakeClientGetter := kubefakes.NewFakeClientGetter(k8s)
	coreSrv, err := server.NewCoreServer(cfg, server.WithClientGetter(fakeClientGetter))
	Expect(err).NotTo(HaveOccurred())

	resp, err := coreSrv.GetFeatureFlags(context.Background(), &pb.GetFeatureFlagsRequest{})
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.Flags).To(HaveKeyWithValue("this is a flag", "you won't find it anywhere else"))
}
