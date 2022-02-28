package server_test

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListHelmReleases(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName, ns.Name, k, g)

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmReleases).To(HaveLen(1))
	g.Expect(res.HelmReleases[0].Name).To(Equal(appName))
}

func TestGetHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp" + rand.String(5)
	ns1 := newNamespace(ctx, k, g)
	ns2 := newNamespace(ctx, k, g)
	ns3 := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName, ns1.Name, k, g)
	newHelmRelease(ctx, appName, ns2.Name, k, g)

	// Get app from ns1.
	response, err := c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:      appName,
		Namespace: ns1.Name,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.HelmRelease.Name).To(Equal(appName))
	g.Expect(response.HelmRelease.Namespace).To(Equal(ns1.Name))

	// Get app from ns2.
	response, err = c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:      appName,
		Namespace: ns2.Name,
	})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.HelmRelease.Name).To(Equal(appName))
	g.Expect(response.HelmRelease.Namespace).To(Equal(ns2.Name))

	// Get app from ns3, should fail.
	_, err = c.GetHelmRelease(ctx, &pb.GetHelmReleaseRequest{
		Name:      appName,
		Namespace: ns3.Name,
	})

	g.Expect(err).To(HaveOccurred())
}

func newHelmRelease(
	ctx context.Context,
	appName, nsName string,
	k client.Client,
	g *GomegaWithT,
) helmv2.HelmRelease {
	release := helmv2.HelmRelease{
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "somesource",
					},
				},
			},
		},
	}
	release.Name = appName
	release.Namespace = nsName

	g.Expect(k.Create(ctx, &release)).To(Succeed())

	return release
}
