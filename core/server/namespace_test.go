package server

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestGetFluxNamespace(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	coreClient, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, client, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	// k8sClient := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)
	ns.ObjectMeta.Labels = map[string]string{
		types.InstanceLabel: "flux-system",
		types.PartOfLabel:   "flux",
	}

	g.Expect(client.Create(ctx, ns)).To(Succeed())

	defer func() {
		// Workaround, somehow it does not get deleted with client.Delete().
		ns.ObjectMeta.Labels = map[string]string{}

		_ = client.Update(ctx, ns)
	}()

	res, err := coreClient.GetFluxNamespace(ctx, &pb.GetFluxNamespaceRequest{})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Name).To(Equal(ns.Name))
}

func TestGetFluxNamespace_notFound(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	coreClient, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, _, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	_, err = coreClient.GetFluxNamespace(ctx, &pb.GetFluxNamespaceRequest{})
	g.Expect(err).To(HaveOccurred())
}
