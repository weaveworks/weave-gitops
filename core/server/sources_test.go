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
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "someapp",
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_HelmRepository,
				Name: "myhelmrepository",
			},
			Chart:    "mychart",
			Version:  "v0.0.0",
			Interval: &pb.Interval{Minutes: 1},
		}

		res, err := c.AddHelmChart(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmChart{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

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
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "",
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_HelmRepository,
				Name: "myhelmrepository",
			},
			Chart:    "mychart",
			Version:  "v0.0.0",
			Interval: nil,
		}

		res, err := c.AddHelmChart(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmChart{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

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
		Name:      "myhelmrepository",
		Namespace: ns.Name,
		AppName:   appName,
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_HelmRepository,
			Name: "myhelmrepository",
		},
		Chart:    "mychart",
		Version:  "v0.0.0",
		Interval: &pb.Interval{Minutes: 1},
	}

	addRes, err := c.AddHelmChart(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedHelmChartReq := &pb.AddHelmChartReq{
		Name:      "otherhelmrepository",
		Namespace: ns.Name,
		AppName:   "",
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_HelmRepository,
			Name: "myhelmrepository",
		},
		Chart:    "mychart",
		Version:  "v0.0.0",
		Interval: &pb.Interval{Minutes: 1},
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
	g.Expect(res.HelmCharts[0].Name).To(Equal(r.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListHelmCharts(ctx, &pb.ListHelmChartReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.HelmCharts).To(HaveLen(2))
}
