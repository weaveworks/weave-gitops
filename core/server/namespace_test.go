package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetFluxNamespace(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	coreClient := makeGRPCServer(k8sEnv.Rest, t)

	kClient, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)
	ns.ObjectMeta.Labels = map[string]string{
		types.InstanceLabel: "flux-system",
		types.PartOfLabel:   "flux",
	}

	g.Expect(kClient.Create(ctx, ns)).To(Succeed())

	defer func() {
		// Workaround, somehow it does not get deleted with client.Delete().
		ns.ObjectMeta.Labels = map[string]string{}

		_ = kClient.Update(ctx, ns)
	}()

	res, err := coreClient.GetFluxNamespace(ctx, &pb.GetFluxNamespaceRequest{})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Name).To(Equal(ns.Name))
}

func TestGetFluxNamespace_notFound(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	coreClient := makeGRPCServer(k8sEnv.Rest, t)

	_, err := coreClient.GetFluxNamespace(ctx, &pb.GetFluxNamespaceRequest{})
	g.Expect(err).To(HaveOccurred())
}

func TestListNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	coreClient := makeGRPCServer(k8sEnv.Rest, t)

	kClient, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	namespaces := []*corev1.Namespace{}
	for len(namespaces) < 5 {
		namespaces = append(namespaces, newNamespace(ctx, kClient, g))
	}

	res, err := coreClient.ListNamespaces(ctx, &pb.ListNamespacesRequest{})

	g.Expect(err).NotTo(HaveOccurred())

	// Can't test with HaveLen, because we created a lot of namespaces in previous
	// test cases and we never clean them up. It would be good, but we don't (yet).
	for _, ns := range namespaces {
		g.Expect(namespaceInResponse(res, ns)).To(BeTrue())
	}
}

func namespaceInResponse(list *pb.ListNamespacesResponse, ns *corev1.Namespace) bool {
	for _, item := range list.Namespaces {
		if item.GetName() == ns.GetName() {
			return true
		}
	}

	return false
}
