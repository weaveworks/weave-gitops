package helm

import (
	"time"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MakeHelmRelease returns a HelmRelease object given a name, version, cluster, namespace, and HelmRepository's name and namespace.
func MakeHelmRelease(name, version, cluster, namespace string, helmRepository types.NamespacedName) *helmv2beta1.HelmRelease {
	return &helmv2beta1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster + "-" + name,
			Namespace: namespace,
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
						Name:       helmRepository.Name,
						Namespace:  helmRepository.Namespace,
					},
				},
			},
			Interval: metav1.Duration{Duration: time.Minute},
		},
	}
}
