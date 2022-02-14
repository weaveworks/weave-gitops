package helm_test

import (
	"time"

	"github.com/weaveworks/weave-gitops/pkg/helm"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
	kyaml "sigs.k8s.io/yaml"
)

var _ = Describe("MakeHelmRelease", func() {
	var (
		name                         string
		cluster                      string
		ns                           string
		version                      string
		helmRepositoryNamespacedName types.NamespacedName
	)

	BeforeEach(func() {
		name = "podinfo"
		cluster = "prod"
		ns = "weave-system"
		version = "6.0.0"
		helmRepositoryNamespacedName = types.NamespacedName{Name: name, Namespace: ns}
	})

	It("creates a helm release", func() {
		actualHelmRelease := helm.MakeHelmRelease(name, version, cluster, ns, helmRepositoryNamespacedName)
		expectedHelmRelease := &helmv2beta1.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster + "-" + name,
				Namespace: ns,
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: helmv2beta1.GroupVersion.Identifier(),
				Kind:       helmv2beta1.HelmReleaseKind,
			},
			Spec: helmv2beta1.HelmReleaseSpec{
				Chart: helmv2beta1.HelmChartTemplate{
					Spec: helmv2beta1.HelmChartTemplateSpec{
						Chart:   name,
						Version: version,
						SourceRef: helmv2beta1.CrossNamespaceObjectReference{
							APIVersion: sourcev1beta1.GroupVersion.Identifier(),
							Kind:       sourcev1beta1.HelmRepositoryKind,
							Name:       helmRepositoryNamespacedName.Name,
							Namespace:  helmRepositoryNamespacedName.Namespace,
						},
					},
				},
				Interval: metav1.Duration{Duration: time.Minute},
			},
		}
		Expect(cmp.Diff(&actualHelmRelease, &expectedHelmRelease)).To(BeEmpty())
	})
})

var _ = Describe("AppendHelmReleaseToString", func() {
	var newRelease *helmv2beta1.HelmRelease

	BeforeEach(func() {
		newRelease = helm.MakeHelmRelease(
			"podinfo", "6.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
	})

	When("the file does not exist", func() {
		It("creates one with the new helm release", func() {
			s, err := helm.AppendHelmReleaseToString("", newRelease)
			Expect(err).NotTo(HaveOccurred())
			r, err := yaml.Marshal(newRelease)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(ContainSubstring(string(r)))
		})
	})
})

var _ = Describe("FindReleaseInNamespace", func() {
	var (
		name = "prod-podinfo"
		ns   = "weave-system"
	)

	When("it finds a HelmRelease with a matching name and namespace", func() {
		It("returns its index in the slice of bytes", func() {
			newRelease := helm.MakeHelmRelease(
				"podinfo", "6.0.0", "prod", "weave-system",
				types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
			)
			existingRelease, index, err := helm.FindReleaseInNamespace([]helmv2beta1.HelmRelease{*newRelease}, name, ns)
			Expect(err).NotTo(HaveOccurred())
			Expect(index).To(Equal(0))
			Expect(cmp.Diff(*existingRelease, *newRelease)).To(BeEmpty())
		})
	})

	When("it does not find a HelmRelease with a matching name and namespace", func() {
		It("returns an index of -1", func() {
			_, index, err := helm.FindReleaseInNamespace([]helmv2beta1.HelmRelease{}, name, ns)
			Expect(err).NotTo(HaveOccurred())
			Expect(index).To(Equal(-1))
		})
	})
})

var _ = Describe("PatchHelmReleaseInString", func() {
	var (
		r *v2beta1.HelmRelease
	)

	BeforeEach(func() {
		r = helm.MakeHelmRelease(
			"podinfo", "6.0.1", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
	})

	It("returns a string with an updated list of HelmReleases", func() {
		existingRelease := helm.MakeHelmRelease(
			"podinfo", "6.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)

		expectedContentBytes, _ := kyaml.Marshal(r)
		expectedContent := "---\n" + string(expectedContentBytes)

		patchedContent, err := helm.PatchHelmRelease([]helmv2beta1.HelmRelease{*existingRelease}, *r, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(cmp.Diff(patchedContent, expectedContent)).To(BeEmpty())
	})
})
