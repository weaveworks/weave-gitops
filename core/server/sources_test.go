package server_test

import (
	"context"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListHelmRepositories(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	appName := "myapp"
	ns := newNamespace()

	hr := &sourcev1.HelmRepository{}
	hr.Name = appName
	hr.Namespace = ns.Name

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(hr, ns).
		Build()

	c := makeGRPCServer(k, t)

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoriesRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmRepositories).To(HaveLen(1))
	g.Expect(res.HelmRepositories[0].Name).To(Equal(appName))
}

func TestListHelmCharts(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	appName := "myapp"
	ns := newNamespace()

	hc := &sourcev1.HelmChart{
		Spec: sourcev1.HelmChartSpec{
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	hc.Name = appName
	hc.Namespace = ns.Name

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(hc, ns).
		Build()

	c := makeGRPCServer(k, t)

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartsRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmCharts).To(HaveLen(1))
	g.Expect(res.HelmCharts[0].Name).To(Equal(appName))
}

func TestListBuckets(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	appName := "myapp"
	ns := newNamespace()

	bucket := &sourcev1.Bucket{
		Spec: sourcev1.BucketSpec{
			SecretRef: &meta.LocalObjectReference{
				Name: "somesecret",
			},
		},
	}
	bucket.Name = appName
	bucket.Namespace = ns.Name

	k := fake.NewClientBuilder().
		WithScheme(kube.CreateScheme()).
		WithRuntimeObjects(bucket, ns).
		Build()

	c := makeGRPCServer(k, t)

	res, err := c.ListBuckets(ctx, &pb.ListBucketRequest{
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Buckets).To(HaveLen(1))
	g.Expect(res.Buckets[0].Name).To(Equal(bucket.Name))
}
