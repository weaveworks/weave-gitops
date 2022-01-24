package helm

import (
	"time"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/pkg/api/profiles"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeHelmRelease(p *profiles.Profile, cluster, namespace, version string) *helmv2beta1.HelmRelease {
	makeHelmReleaseName := func(clusterName, profileName string) string {
		return clusterName + "-" + profileName
	}

	return &helmv2beta1.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makeHelmReleaseName(cluster, p.Name),
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmv2beta1.GroupVersion.Identifier(),
			Kind:       helmv2beta1.HelmReleaseKind,
		},
		Spec: helmv2beta1.HelmReleaseSpec{
			Chart: helmv2beta1.HelmChartTemplate{
				Spec: helmv2beta1.HelmChartTemplateSpec{
					Chart:   p.Name,
					Version: version,
					SourceRef: helmv2beta1.CrossNamespaceObjectReference{
						APIVersion: sourcev1beta1.GroupVersion.Identifier(),
						Kind:       sourcev1beta1.HelmRepositoryKind,
						Name:       p.HelmRepository.Name,
						Namespace:  p.HelmRepository.Namespace,
					},
				},
			},
			Interval: metav1.Duration{Duration: time.Minute},
		},
	}
}
