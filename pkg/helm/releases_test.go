package helm_test

import (
	"time"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		Expect(cmp.Diff(actualHelmRelease, expectedHelmRelease)).To(BeEmpty())
	})
})
