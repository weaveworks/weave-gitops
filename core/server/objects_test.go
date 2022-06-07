package server_test

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetObject(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
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
