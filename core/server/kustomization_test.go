package server_test

import (
	"context"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestListKustomizations(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	kust := &kustomizev1.Kustomization{
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	kust.Name = appName
	kust.Namespace = ns.Name

	g.Expect(k.Create(ctx, kust)).To(Succeed())

	res, err := c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Kustomizations).To(HaveLen(1))
	g.Expect(res.Kustomizations[0].Name).To(Equal(appName))
}

func TestGetKustomization(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	kust := &kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev1.KustomizationKind,
			APIVersion: kustomizev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind:       sourcev1.GitRepositoryKind,
				APIVersion: sourcev1.GroupVersion.String(),
			},
		},
	}
	kust.Name = appName
	kust.Namespace = ns.Name

	opt := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("kustomize-controller"),
	}

	g.Expect(k.Patch(ctx, kust, client.Apply, opt...)).To(Succeed())

	st := kustomizev1.KustomizationStatus{
		Inventory: &kustomizev1.ResourceInventory{
			Entries: []kustomizev1.ResourceRef{
				{
					Version: "v1",
					ID:      ns.Name + "_my-deployment_apps_Deployment",
				},
			},
		},
	}

	kust.ManagedFields = nil
	kust.Status = st

	g.Expect(k.Status().Patch(ctx, kust, client.Apply, opt...)).To(Succeed())

	t.Run("gets a kustomization", func(t *testing.T) {
		res, err := c.GetKustomization(ctx, &pb.GetKustomizationRequest{Name: appName, Namespace: ns.Name})
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(len(res.Kustomization.Inventory)).To(Equal(1))
		g.Expect(res.Kustomization.Inventory[0].Group).To(Equal("apps"))
	})
	t.Run("returns not found", func(t *testing.T) {
		_, err = c.GetKustomization(ctx, &pb.GetKustomizationRequest{Name: "somename", Namespace: ns.Name})
		g.Expect(err).To(HaveOccurred())

		status, ok := status.FromError(err)
		if !ok {
			t.Error("could not get status from error")
		}

		g.Expect(status.Code()).To(Equal(codes.NotFound))
	})
}

func newNamespace(ctx context.Context, k client.Client, g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k.Create(ctx, ns)).To(Succeed())

	return ns
}
