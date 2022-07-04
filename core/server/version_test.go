package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testVersion = "some-version"
)

func TestGetVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	c, _ := makeGRPCServer(k8sEnv.Rest, t)
	_, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})

	g.Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	_, err = client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	fluxNs := &v1.Namespace{}
	fluxNs.Name = "flux-ns-test"
	fluxNs.Labels = map[string]string{
		coretypes.PartOfLabel: server.FluxNamespacePartOf,
		flux.VersionLabelKey:  testVersion,
	}
	g.Expect(k.Create(ctx, fluxNs)).To(Succeed())

	resp, err := c.GetVersion(ctx, &pb.GetVersionRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp.Semver).To(Equal("v0.0.0"))
	g.Expect(resp.FluxVersion).To(Equal(testVersion))
}
