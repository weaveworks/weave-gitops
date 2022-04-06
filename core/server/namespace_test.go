package server_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetFluxNamespace(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)
	ns.ObjectMeta.Labels = map[string]string{
		types.InstanceLabel: "flux-system",
		types.PartOfLabel:   "flux",
	}

	kClient := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(ns).
		Build()

	coreClient := makeGRPCServer(kClient, t)

	res, err := coreClient.GetFluxNamespace(ctx, &pb.GetFluxNamespaceRequest{})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Name).To(Equal(ns.Name))
}

func TestGetFluxNamespace_notFound(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	kClient := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		Build()

	coreClient := makeGRPCServer(kClient, t)

	_, err := coreClient.GetFluxNamespace(ctx, &pb.GetFluxNamespaceRequest{})
	g.Expect(err).To(HaveOccurred())
}

func TestListNamespaces(t *testing.T) {
	ctx := context.Background()

	namespaces := []corev1.Namespace{}
	for len(namespaces) < 5 {
		namespaces = append(namespaces, *newNamespace())
	}

	objs := []runtime.Object{}
	for i := range namespaces {
		objs = append(objs, &namespaces[i])
	}

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(objs...).
		Build()

	coreClient := makeGRPCServer(k, t)

	t.Run("returns a list of namespaces", func(t *testing.T) {
		g := NewGomegaWithT(t)

		res, err := coreClient.ListNamespaces(ctx, &pb.ListNamespacesRequest{})

		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Namespaces).To(HaveLen(5))
		for _, ns := range namespaces {
			g.Expect(namespaceInResponse(res, ns)).To(BeTrue())
		}

	})

	t.Run("returns filtered namespaces", func(t *testing.T) {
		g := NewGomegaWithT(t)
		filtered := namespaces[:3]
		nsChecker.FilterAccessibleNamespacesReturns(filtered, nil)

		res, err := coreClient.ListNamespaces(ctx, &pb.ListNamespacesRequest{})
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Namespaces).To(HaveLen(len(filtered)))
	})
}

func namespaceInResponse(list *pb.ListNamespacesResponse, ns corev1.Namespace) bool {
	for _, item := range list.Namespaces {
		if item.GetName() == ns.GetName() {
			return true
		}
	}

	return false
}
