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

func TestListGitRepositories(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)
	newGitRepo(ctx, appName, ns.Name, k, g)

	res, err := c.ListGitRepositories(ctx, &pb.ListGitRepositoriesRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.GitRepositories).To(HaveLen(1))
	g.Expect(res.GitRepositories[0].Name).To(Equal(appName))
}

func TestListHelmRepositories(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

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

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoriesRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmRepositories).To(HaveLen(1))
	g.Expect(res.HelmRepositories[0].Name).To(Equal(appName))
}

func TestListGitRepositories_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName1 := "myapp1"
	ns1 := newNamespace(ctx, k, g)
	newGitRepo(ctx, appName1, ns1.Name, k, g)

	appName2 := "myapp2"
	ns2 := newNamespace(ctx, k, g)
	newGitRepo(ctx, appName2, ns2.Name, k, g)

	res, err := c.ListGitRepositories(ctx, &pb.ListGitRepositoriesRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	resourcesFound := 0

	for _, r := range res.GitRepositories {
		if r.Name == appName1 || r.Name == appName2 {
			resourcesFound++
		}
	}

	g.Expect(resourcesFound).To(Equal(2))
}

func TestListHelmRepositories_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName1 := "myapp1"
	ns1 := newNamespace(ctx, k, g)
	hr := &sourcev1.HelmRepository{}
	hr.Name = appName1
	hr.Namespace = ns1.Name
	g.Expect(k.Create(ctx, hr)).To(Succeed())

	appName2 := "myapp2"
	ns2 := newNamespace(ctx, k, g)
	hr = &sourcev1.HelmRepository{}
	hr.Name = appName2
	hr.Namespace = ns2.Name
	g.Expect(k.Create(ctx, hr)).To(Succeed())

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoriesRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	resourcesFound := 0

	for _, r := range res.HelmRepositories {
		if r.Name == appName1 || r.Name == appName2 {
			resourcesFound++
		}
	}

	g.Expect(resourcesFound).To(Equal(2))
}

func TestListHelmCharts(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

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

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartsRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmCharts).To(HaveLen(1))
	g.Expect(res.HelmCharts[0].Name).To(Equal(appName))
}

func TestListHelmCharts_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName1 := "myapp1"
	ns1 := newNamespace(ctx, k, g)
	newHelmChart(ctx, appName1, ns1.Name, k, g)

	appName2 := "myapp2"
	ns2 := newNamespace(ctx, k, g)
	newHelmChart(ctx, appName2, ns2.Name, k, g)

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartsRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	resourcesFound := 0

	for _, r := range res.HelmCharts {
		if r.Name == appName1 || r.Name == appName2 {
			resourcesFound++
		}
	}

	g.Expect(resourcesFound).To(Equal(2))
}

func TestListBuckets(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

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

	res, err := c.ListBuckets(ctx, &pb.ListBucketRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Buckets).To(HaveLen(1))
	g.Expect(res.Buckets[0].Name).To(Equal(bucket.Name))
}

func TestListBuckets_inMultipleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, _ := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	appName1 := "myapp1"
	ns1 := newNamespace(ctx, k, g)
	newBucket(ctx, appName1, ns1.Name, k, g)

	appName2 := "myapp2"
	ns2 := newNamespace(ctx, k, g)
	newBucket(ctx, appName2, ns2.Name, k, g)

	res, err := c.ListBuckets(ctx, &pb.ListBucketRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	resourcesFound := 0

	for _, r := range res.Buckets {
		if r.Name == appName1 || r.Name == appName2 {
			resourcesFound++
		}
	}

	g.Expect(resourcesFound).To(Equal(2))
}

func newGitRepo(ctx context.Context, name, namespace string, k client.Client, g *GomegaWithT) *sourcev1.GitRepository {
	repo := &sourcev1.GitRepository{
		Spec: sourcev1.GitRepositorySpec{
			URL:       "https://example.com/repo",
			Reference: &sourcev1.GitRepositoryRef{},
		},
	}
	repo.Name = name
	repo.Namespace = namespace
	g.Expect(k.Create(ctx, repo)).To(Succeed())

	return repo
}

func newHelmChart(ctx context.Context, appName, nsName string, k client.Client, g *GomegaWithT) *sourcev1.HelmChart {
	hc := &sourcev1.HelmChart{
		Spec: sourcev1.HelmChartSpec{
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				Kind: "GitRepository",
			},
		},
	}
	hc.Name = appName
	hc.Namespace = nsName

	g.Expect(k.Create(ctx, hc)).To(Succeed())

	return hc
}

func newBucket(ctx context.Context, appName, nsName string, k client.Client, g *GomegaWithT) *sourcev1.Bucket {
	bucket := &sourcev1.Bucket{
		Spec: sourcev1.BucketSpec{
			SecretRef: &meta.LocalObjectReference{
				Name: "somesecret",
			},
		},
	}
	bucket.Name = appName
	bucket.Namespace = nsName

	g.Expect(k.Create(ctx, bucket)).To(Succeed())

	return bucket
}
