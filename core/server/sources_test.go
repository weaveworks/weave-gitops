package server_test

import (
	"context"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListHelmRepositories(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	hr := &sourcev1.HelmRepository{}
	hr.Name = appName
	hr.Namespace = ns.Name

	g.Expect(k.Create(ctx, hr)).To(Succeed())

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoriesRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmRepositories).To(HaveLen(1))
	g.Expect(res.HelmRepositories[0].Name).To(Equal(appName))
}

func TestListHelmRepositories_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	existingHelmRepositoriesNo := func() int {
		res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoriesRequest{})
		g.Expect(err).NotTo(HaveOccurred())

		return len(res.HelmRepositories)
	}()

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	hr := &sourcev1.HelmRepository{}
	hr.Name = appName
	hr.Namespace = ns.Name

	g.Expect(k.Create(ctx, hr)).To(Succeed())

	updateNamespaceCache()

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoriesRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmRepositories).To(HaveLen(existingHelmRepositoriesNo + 1))
}

func TestListHelmCharts(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	hc := &sourcev1.HelmChart{
		Spec: sourcev1.HelmChartSpec{
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	hc.Name = appName
	hc.Namespace = ns.Name

	g.Expect(k.Create(ctx, hc)).To(Succeed())

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartsRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmCharts).To(HaveLen(1))
	g.Expect(res.HelmCharts[0].Name).To(Equal(appName))
}

func TestListHelmCharts_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	existingHelmChartNo := func() int {
		res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartsRequest{})
		g.Expect(err).NotTo(HaveOccurred())

		return len(res.HelmCharts)
	}()

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	hc := &sourcev1.HelmChart{
		Spec: sourcev1.HelmChartSpec{
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	hc.Name = appName
	hc.Namespace = ns.Name

	g.Expect(k.Create(ctx, hc)).To(Succeed())

	updateNamespaceCache()

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartsRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmCharts).To(HaveLen(existingHelmChartNo + 1))
}

func TestListBuckets(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	bucket := &sourcev1.Bucket{
		Spec: sourcev1.BucketSpec{
			SecretRef: &meta.LocalObjectReference{
				Name: "somesecret",
			},
		},
	}
	bucket.Name = appName
	bucket.Namespace = ns.Name

	g.Expect(k.Create(ctx, bucket)).To(Succeed())

	res, err := c.ListBuckets(ctx, &pb.ListBucketRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Buckets).To(HaveLen(1))
	g.Expect(res.Buckets[0].Name).To(Equal(bucket.Name))
}

func TestListBuckets_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	existingBucketNo := func() int {
		res, err := c.ListBuckets(ctx, &pb.ListBucketRequest{})
		g.Expect(err).NotTo(HaveOccurred())

		return len(res.Buckets)
	}()

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	bucket := &sourcev1.Bucket{
		Spec: sourcev1.BucketSpec{
			SecretRef: &meta.LocalObjectReference{
				Name: "somesecret",
			},
		},
	}
	bucket.Name = appName
	bucket.Namespace = ns.Name

	g.Expect(k.Create(ctx, bucket)).To(Succeed())

	updateNamespaceCache()

	res, err := c.ListBuckets(ctx, &pb.ListBucketRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Buckets).To(HaveLen(existingBucketNo + 1))
}
