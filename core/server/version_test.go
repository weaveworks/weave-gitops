package server_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestGetVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := t.Context()

	c := makeGRPCServer(ctx, t, k8sEnv.Rest)
	logf.SetLogger(logr.Discard())

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	_, err = client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	md := metadata.Pairs(MetadataUserKey, "anne", MetadataGroupsKey, "system:masters")
	outgoingCtx := metadata.NewOutgoingContext(ctx, md)
	resp, err := c.GetVersion(outgoingCtx, &pb.GetVersionRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp.Semver).To(Equal("v0.0.0"))
}
