package helm_test

import (
	"time"

	"github.com/weaveworks/weave-gitops/pkg/helm"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
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
		expectedHelmRelease := &helmv2.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cluster + "-" + name,
				Namespace: ns,
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: helmv2.GroupVersion.Identifier(),
				Kind:       helmv2.HelmReleaseKind,
			},
			Spec: helmv2.HelmReleaseSpec{
				Chart: helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Chart:   name,
						Version: version,
						SourceRef: helmv2.CrossNamespaceObjectReference{
							APIVersion: sourcev1.GroupVersion.Identifier(),
							Kind:       sourcev1.HelmRepositoryKind,
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
	var newRelease *helmv2.HelmRelease

	BeforeEach(func() {
		newRelease = helm.MakeHelmRelease(
			"podinfo", "6.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
	})

	When("the given string is empty", func() {
		It("appends a HelmRelease to it", func() {
			s, err := helm.AppendHelmReleaseToString("", newRelease)
			Expect(err).NotTo(HaveOccurred())
			r, err := yaml.Marshal(newRelease)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(ContainSubstring(string(r)))
		})
	})

	When("the given string is not empty", func() {
		It("appends a HelmRelease to it", func() {
			b, _ := kyaml.Marshal(helm.MakeHelmRelease(
				"another-profile", "7.0.0", "prod", "test-namespace",
				types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
			))
			s, err := helm.AppendHelmReleaseToString(string(b), newRelease)
			Expect(err).NotTo(HaveOccurred())
			r, err := yaml.Marshal(newRelease)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(ContainSubstring(string(r)))
		})
	})
})

var _ = Describe("MarshalHelmRelease", func() {
	It("returns a string with an updated list of HelmReleases", func() {
		release1 := helm.MakeHelmRelease(
			"random-profile", "7.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
		releaseBytes1, _ := kyaml.Marshal(release1)

		release2 := helm.MakeHelmRelease(
			"podinfo", "6.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
		releaseBytes2, _ := kyaml.Marshal(release2)

		patchedContent, err := helm.MarshalHelmReleases([]*helmv2.HelmRelease{release1, release2})
		Expect(err).NotTo(HaveOccurred())
		Expect(cmp.Diff(patchedContent, "---\n"+string(releaseBytes1)+"---\n"+string(releaseBytes2))).To(BeEmpty())
	})
})

var _ = Describe("SplitHelmReleaseYAML", func() {
	When("the resource contains only HelmRelease", func() {
		It("returns a slice of HelmReleases", func() {
			r1 := helm.MakeHelmRelease(
				"podinfo", "6.0.0", "prod", "weave-system",
				types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
			)
			r2 := helm.MakeHelmRelease(
				"profile", "6.0.1", "prod", "test-namespace",
				types.NamespacedName{Name: "helm-repo-name", Namespace: "test-namespace"},
			)
			b1, _ := kyaml.Marshal(r1)
			bytes := append(b1, []byte("\n---\n")...)
			b2, _ := kyaml.Marshal(r2)
			list, err := helm.SplitHelmReleaseYAML(append(bytes, b2...))
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(ContainElements(r1, r2))
		})
	})

	When("the resource contains any resource other than a HelmRelease", func() {
		It("returns an error", func() {
			b, _ := kyaml.Marshal("content")
			_, err := helm.SplitHelmReleaseYAML(b)
			Expect(err).To(MatchError(ContainSubstring("error unmarshaling JSON")))
		})
	})
})
