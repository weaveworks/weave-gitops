package helm_test

import (
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm"
)

var _ = Describe("MakeHelmRelease", func() {
	var (
		name    string
		cluster string
		ns      string
		profile *profiles.Profile
	)

	BeforeEach(func() {
		name = "podinfo"
		cluster = "prod"
		ns = "weave-system"
		profile = &profiles.Profile{
			Name: name,
			HelmRepository: &profiles.HelmRepository{
				Name:      name,
				Namespace: ns,
			},
			AvailableVersions: []string{"6.0.0", "6.0.1"},
		}
	})

	It("creates a helm release", func() {
		hr := helm.MakeHelmRelease(profile, cluster, ns)
		Expect(hr.Name).To(Equal(cluster + "-" + name))
		Expect(hr.Namespace).To(Equal(ns))
		Expect(hr.TypeMeta.APIVersion).To(Equal(helmv2beta1.GroupVersion.Identifier()))
		Expect(hr.TypeMeta.Kind).To(Equal(helmv2beta1.HelmReleaseKind))
		Expect(hr.Spec.Chart.Spec.Chart).To(Equal(name))
		Expect(hr.Spec.Chart.Spec.Version).To(Equal("6.0.0"))
		Expect(hr.Spec.Chart.Spec.SourceRef.Name).To(Equal(name))
		Expect(hr.Spec.Chart.Spec.SourceRef.Namespace).To(Equal(ns))
	})
})
