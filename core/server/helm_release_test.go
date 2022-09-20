package server_test

import (
	"context"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListHelmReleases(t *testing.T) {
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

	newHelmRelease(ctx, appName, ns.Name, k, g)

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmReleases).To(HaveLen(1))
	g.Expect(res.HelmReleases[0].Name).To(Equal(appName))
}

func TestListHelmReleases_inMultipleNamespaces(t *testing.T) {
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

	newHelmRelease(ctx, appName1, ns1.Name, k, g)

	appName2 := "myapp-2"
	ns2 := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName2, ns2.Name, k, g)

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	releasesFound := 0

	for _, r := range res.HelmReleases {
		if r.Name == appName1 || r.Name == appName2 {
			releasesFound++
		}
	}

	g.Expect(releasesFound).To(Equal(2))
}

/**
 * This test demonstrates the behavior of ListHelmReleases.
 * The third HR created in this test is managed by the Helm Release Controller,
 * but does not have a corresponding secret. Because of this the getHelmReleaseInventory
 * call throws an error. This should not cause the entire list call to error out, and
 * instead skip just the one HR that is in error state.
 **/
func TestListHelmReleases_withInventory(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp" + rand.String(5)
	nsA := newNamespace(ctx, k, g)
	nsB := newNamespace(ctx, k, g)
	nsC := newNamespace(ctx, k, g)

	newHelmRelease(ctx, appName, nsA.Name, k, g)
	newHelmRelease(ctx, appName, nsB.Name, k, g)

	res1, error1 := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{})

	g.Expect(error1).NotTo(HaveOccurred())

	releasesFound := len(res1.HelmReleases)

	// Create helm release without a corresponding secret.
	helmRelease := helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: nsC.Name,
		},
		Status: helmv2.HelmReleaseStatus{
			LastReleaseRevision: 2,
		},
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

	opt := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("helmrelease-controller"),
	}

	g.Expect(k.Patch(ctx, &helmRelease, client.Apply, opt...)).To(Succeed())

	helmRelease.Status.LastReleaseRevision = 1
	helmRelease.ManagedFields = nil

	g.Expect(k.Status().Patch(ctx, &helmRelease, client.Apply, opt...)).To(Succeed())

	res2, error2 := c.ListHelmReleases(ctx, &pb.ListHelmReleasesRequest{})

	g.Expect(error2).NotTo(HaveOccurred())
	g.Expect(res2.HelmReleases).To(HaveLen(releasesFound))
}

func helmReleaseInventoryObjects() string {
	return `
---
apiVersion: v1
kind: Secret
metadata:
  name: test
  namespace: default
immutable: true
stringData:
  key: "private-key"
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: crossplane-provider-aws2
spec:
  package: crossplane/provider-aws:v0.23.0
  controllerConfigRef:
    name: provider-aws
`
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
