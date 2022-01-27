package server

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestCreateKustomization(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddKustomizationReq{
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "someapp",
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_GitRepository,
				Name: "othersource",
			},
		}

		res, err := c.AddKustomization(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &kustomizev1.Kustomization{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToKustomization(r)

		opt := cmpopts.IgnoreFields(kustomizev1.Kustomization{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddKustomizationReq{
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "",
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_GitRepository,
				Name: "othersource",
			},
		}

		res, err := c.AddKustomization(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &kustomizev1.Kustomization{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToKustomization(r)

		opt := cmpopts.IgnoreFields(kustomizev1.Kustomization{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListKustomizations(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddKustomizationReq{
		Name:      "mykustomization",
		Namespace: ns.Name,
		AppName:   appName,
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_GitRepository,
			Name: "othersource",
		},
	}

	addRes, err := c.AddKustomization(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedKustomizationReq := &pb.AddKustomizationReq{
		Name:      "otherkustomization",
		Namespace: ns.Name,
		AppName:   "",
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_GitRepository,
			Name: "othersource",
		},
	}

	_, err = c.AddKustomization(ctx, unAssociatedKustomizationReq)
	g.Expect(err).NotTo(HaveOccurred())

	res, err := c.ListKustomizations(ctx, &pb.ListKustomizationsReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Kustomizations).To(HaveLen(1))
	g.Expect(res.Kustomizations[0].Name).To(Equal(r.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListKustomizations(ctx, &pb.ListKustomizationsReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.Kustomizations).To(HaveLen(2))

}

func TestRemoveKustomization(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddKustomizationReq{
		Name:      "mykustomization",
		Namespace: ns.Name,
		AppName:   appName,
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_GitRepository,
			Name: "othersource",
		},
	}

	addRes, err := c.AddKustomization(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	name := types.NamespacedName{Name: r.Name, Namespace: r.Namespace}
	// Make sure the kustomization actually got created
	g.Expect(k.Get(ctx, name, &kustomizev1.Kustomization{})).To(Succeed())

	res, err := c.RemoveKustomizations(ctx, &pb.RemoveKustomizationReq{
		KustomizationName: r.Name,
		Namespace:         r.Namespace,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Success).To(BeTrue())

	err = k.Get(ctx, name, &kustomizev1.Kustomization{})
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "expected a NotFound error after removal")
}

func newNamespace(ctx context.Context, k client.Client, g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k.Create(ctx, ns)).To(Succeed())

	return ns
}
