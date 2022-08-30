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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetObject(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	appName := "myapp"

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust).Build()
	cfg := makeServerConfig(client, t)
	c := makeServer(cfg, t)

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

func TestListObjectSingle(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"toolkit.fluxcd.io/tenant": "Neil",
			},
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-name",
			Namespace: ns.Name,
			UID:       "not a real uid",
		},
		Spec: kustomizev1.KustomizationSpec{},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust).Build()
	cfg := makeServerConfig(client, t)
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace: ns.Name,
		Kind:      kustomizev1.KustomizationKind,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(1))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("kust-name"))
	g.Expect(res.Objects[0].Uid).To(Equal("not a real uid"))
	g.Expect(res.Objects[0].Tenant).To(Equal("Neil"))
}

func TestListObjectMultiple(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-name",
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{},
	}
	helm1 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}
	helm2 := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "second-helm-name",
			Namespace: ns.Name,
		},
		Spec: helmv2.HelmReleaseSpec{},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(ns, kust, helm1, helm2).Build()
	cfg := makeServerConfig(client, t)
	c := makeServer(cfg, t)

	res, err := c.ListObjects(ctx, &pb.ListObjectsRequest{
		Namespace: ns.Name,
		Kind:      helmv2.HelmReleaseKind,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Errors).To(BeEmpty())
	g.Expect(res.Objects).To(HaveLen(2))
	g.Expect(res.Objects[0].ClusterName).To(Equal("Default"))
	g.Expect(res.Objects[0].Payload).To(ContainSubstring("helm-name"))
	g.Expect(res.Objects[1].Payload).To(ContainSubstring("helm-name"))
}
