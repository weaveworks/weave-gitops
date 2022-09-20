package server_test

import (
	"context"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestListKustomizations(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	newKustomization(ctx, appName, ns.Name, k, g)

	res, err := c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Kustomizations).To(HaveLen(1))
	g.Expect(res.Kustomizations[0].Name).To(Equal(appName))
}

func TestListKustomizations_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName1 := "myapp-1"
	ns1 := newNamespace(ctx, k, g)
	newKustomization(ctx, appName1, ns1.Name, k, g)

	appName2 := "myapp-2"
	ns2 := newNamespace(ctx, k, g)
	newKustomization(ctx, appName2, ns2.Name, k, g)

	res, err := c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	resourcesFound := 0

	for _, r := range res.Kustomizations {
		if r.Name == appName1 || r.Name == appName2 {
			resourcesFound++
		}
	}

	g.Expect(resourcesFound).To(Equal(2))
}

func TestListKustomizationPagination(t *testing.T) {
	g := NewGomegaWithT(t)

	testutils.DeleteAllOf(g, &kustomizev1.Kustomization{})

	ctx := context.Background()
	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	ns1 := newNamespace(ctx, k, g)

	for i := 0; i < 2; i++ {
		appName := "myapp-" + strconv.Itoa(i)
		newKustomization(ctx, appName, ns1.Name, k, g)
	}

	ns2 := newNamespace(ctx, k, g)

	for i := 0; i < 2; i++ {
		appName := "myapp-" + strconv.Itoa(i)
		newKustomization(ctx, appName, ns2.Name, k, g)
	}

	res, err := c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{
		Pagination: &pb.Pagination{
			PageSize: 1,
		},
	})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res.Kustomizations).To(HaveLen(2))

	res, err = c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{
		Pagination: &pb.Pagination{
			PageSize:  1,
			PageToken: res.NextPageToken,
		},
	})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res.Kustomizations).To(HaveLen(2))

	res, err = c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{
		Pagination: &pb.Pagination{
			PageSize:  1,
			PageToken: res.NextPageToken,
		},
	})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res.Kustomizations).To(HaveLen(0))
}

func newNamespace(ctx context.Context, k client.Client, g *GomegaWithT) corev1.Namespace {
	ns := corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k.Create(ctx, &ns)).To(Succeed())

	return ns
}

func newKustomization(ctx context.Context, appName, nsName string, k client.Client, g *GomegaWithT) *kustomizev1.Kustomization {
	kust := &kustomizev1.Kustomization{
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	kust.Name = appName
	kust.Namespace = nsName

	g.Expect(k.Create(ctx, kust)).To(Succeed())

	return kust
}
