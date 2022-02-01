package server

import (
	"context"
	"fmt"
	"testing"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateHelmRepository(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmRepositoryReq{
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "someapp",
			Url:       "someurl",
			Interval:  &pb.Interval{Minutes: 1},
		}

		res, err := c.AddHelmRepository(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmRepository{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmRepository(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmRepository{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmRepositoryReq{
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "",
			Url:       "",
			Interval:  nil,
		}

		res, err := c.AddHelmRepository(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmRepository{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmRepository(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmRepository{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListHelmRepositories(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddHelmRepositoryReq{
		Name:      "myhelmrepository",
		Namespace: ns.Name,
		AppName:   appName,
		Url:       "someurl",
		Interval:  &pb.Interval{Minutes: 1},
	}

	addRes, err := c.AddHelmRepository(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedKustomizationReq := &pb.AddHelmRepositoryReq{
		Name:      "otherhelmrepository",
		Namespace: ns.Name,
		AppName:   "",
		Url:       "someurl",
		Interval:  &pb.Interval{Minutes: 1},
	}

	_, err = c.AddHelmRepository(ctx, unAssociatedKustomizationReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoryReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmRepositories).To(HaveLen(1))
	g.Expect(res.HelmRepositories[0].Name).To(Equal(r.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoryReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.HelmRepositories).To(HaveLen(2))
}

func TestCreateHelmChart(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmChartReq{
			AppName:   "someapp",
			Namespace: ns.Name,
			HelmChart: &pb.HelmChart{
				Name:      "myhelmchart",
				Namespace: ns.Name,
				SourceRef: &pb.SourceRef{
					Kind: pb.SourceRef_HelmRepository,
					Name: "myhelmrepository",
				},
				Chart:    "mychart",
				Version:  "v0.0.0",
				Interval: &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddHelmChart(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmChart{}

		fmt.Println("r.HelmChart.Name", r.HelmChart.Name)
		fmt.Println("ns.Name", ns.Name)

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.HelmChart.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmChart(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmChart{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmChartReq{
			AppName:   "",
			Namespace: ns.Name,
			HelmChart: &pb.HelmChart{
				Name:      "myhelmchart",
				Namespace: ns.Name,
				SourceRef: &pb.SourceRef{
					Kind: pb.SourceRef_HelmRepository,
					Name: "myhelmrepository",
				},
				Chart:    "mychart",
				Version:  "v0.0.0",
				Interval: &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddHelmChart(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmChart{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.HelmChart.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmChart(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmChart{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListHelmCharts(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddHelmChartReq{
		AppName:   appName,
		Namespace: ns.Name,
		HelmChart: &pb.HelmChart{
			Name:      "myhelmchart",
			Namespace: ns.Name,
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_HelmRepository,
				Name: "myhelmrepository",
			},
			Chart:    "mychart",
			Version:  "v0.0.0",
			Interval: &pb.Interval{Minutes: 1},
		},
	}

	addRes, err := c.AddHelmChart(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedHelmChartReq := &pb.AddHelmChartReq{
		AppName:   "",
		Namespace: ns.Name,
		HelmChart: &pb.HelmChart{
			Name:      "otherhelmrepository",
			Namespace: ns.Name,
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_HelmRepository,
				Name: "myhelmrepository",
			},
			Chart:    "mychart",
			Version:  "v0.0.0",
			Interval: &pb.Interval{Minutes: 1},
		},
	}

	_, err = c.AddHelmChart(ctx, unAssociatedHelmChartReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmCharts).To(HaveLen(1))
	g.Expect(res.HelmCharts[0].Name).To(Equal(r.HelmChart.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListHelmCharts(ctx, &pb.ListHelmChartReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.HelmCharts).To(HaveLen(2))
}

func TestCreateBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddBucketReq{
			AppName:   "someapp",
			Namespace: ns.Name,
			Bucket: &pb.Bucket{
				Name:          "mybucket",
				Namespace:     ns.Name,
				Endpoint:      "endpoint",
				Insecure:      true,
				Provider:      pb.Bucket_AWS,
				Region:        "myregion",
				SecretRefName: "secret_name",
				Interval:      &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddBucket(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.Bucket{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Bucket.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToBucket(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.Bucket{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddBucketReq{
			AppName:   "",
			Namespace: ns.Name,
			Bucket: &pb.Bucket{
				Name:          "mybucket",
				Namespace:     ns.Name,
				Endpoint:      "endpoint",
				Insecure:      true,
				Provider:      pb.Bucket_AWS,
				Region:        "myregion",
				SecretRefName: "secret_name",
				Interval:      &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddBucket(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.Bucket{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Bucket.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToBucket(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.Bucket{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListBuckets(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddBucketReq{
		AppName:   appName,
		Namespace: ns.Name,
		Bucket: &pb.Bucket{
			Name:          "mybucket",
			Namespace:     ns.Name,
			Endpoint:      "endpoint",
			Insecure:      true,
			Provider:      pb.Bucket_AWS,
			Region:        "myregion",
			SecretRefName: "secret_name",
			Interval:      &pb.Interval{Minutes: 1},
		},
	}

	addRes, err := c.AddBucket(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedBucketReq := &pb.AddBucketReq{
		AppName:   "",
		Namespace: ns.Name,
		Bucket: &pb.Bucket{
			Name:          "othermybucket",
			Namespace:     ns.Name,
			Endpoint:      "endpoint",
			Insecure:      true,
			Provider:      pb.Bucket_AWS,
			Region:        "myregion",
			SecretRefName: "secret_name",
			Interval:      &pb.Interval{Minutes: 1},
		},
	}

	_, err = c.AddBucket(ctx, unAssociatedBucketReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListBuckets(ctx, &pb.ListBucketReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Buckets).To(HaveLen(1))
	g.Expect(res.Buckets[0].Name).To(Equal(r.Bucket.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListBuckets(ctx, &pb.ListBucketReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.Buckets).To(HaveLen(2))
}
