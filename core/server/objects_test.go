package server_test

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetObject(t *testing.T) {
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
	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(Equal("Default"))
	g.Expect(res.Object.Payload).NotTo(BeEmpty())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        helmv2.HelmReleaseKind,
		ClusterName: "Default",
	})
	g.Expect(err).To(HaveOccurred())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Other",
	})
	g.Expect(err).To(HaveOccurred())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        "other name",
		Namespace:   ns.Name,
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Defauult",
	})
	g.Expect(err).To(HaveOccurred())

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   "other ns",
		Kind:        kustomizev1.KustomizationKind,
		ClusterName: "Defauult",
	})
	g.Expect(err).To(HaveOccurred())
}

func TestGetObjectOtherKinds(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	appName := "myapp"
	dep := newDeployment(appName, ns.Name)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&ns, dep).Build()
	cfg := makeServerConfig(client, t)

	c := makeServer(cfg, t)

	_, err = c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        "deployment",
		ClusterName: "Default",
	})
	g.Expect(err).To(HaveOccurred())

	err = cfg.PrimaryKinds.Add("deployment", schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	g.Expect(err).NotTo(HaveOccurred())

	c = makeServer(cfg, t)

	res, err := c.GetObject(ctx, &pb.GetObjectRequest{
		Name:        appName,
		Namespace:   ns.Name,
		Kind:        "deployment",
		ClusterName: "Default",
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Object.ClusterName).To(Equal("Default"))
	g.Expect(res.Object.Payload).NotTo(BeEmpty())
}
