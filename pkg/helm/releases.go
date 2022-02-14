package helm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	apimachinery "k8s.io/apimachinery/pkg/util/yaml"
	kyaml "sigs.k8s.io/yaml"
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

// FindReleaseInNamespace iterates through a slice of HelmReleases to find one with a given name in a given namespace, and returns it with its index.
func FindReleaseInNamespace(existingReleases []helmv2beta1.HelmRelease, name, ns string) (*helmv2beta1.HelmRelease, int, error) {
	for i, r := range existingReleases {
		if r.Name == name && r.Namespace == ns {
			return &r, i, nil
		}
	}

	return nil, -1, nil
}

// AppendHelmReleaseToString appends a HelmRelease to a string.
func AppendHelmReleaseToString(content string, newRelease *helmv2beta1.HelmRelease) (string, error) {
	var sb strings.Builder
	if content != "" {
		sb.WriteString(content + "\n")
	}

	helmReleaseManifest, err := kyaml.Marshal(newRelease)
	if err != nil {
		return "", fmt.Errorf("failed to marshal HelmRelease: %w", err)
	}

	sb.WriteString("---\n" + string(helmReleaseManifest))

	return sb.String(), nil
}

// SplitHelmReleaseYAML splits a manifest file that contains one or more Helm Releases that may be separated by '---'.
func SplitHelmReleaseYAML(resources []byte) ([]helmv2beta1.HelmRelease, error) {
	var helmReleaseList []helmv2beta1.HelmRelease

	decoder := apimachinery.NewYAMLOrJSONDecoder(bytes.NewReader(resources), 100000000)

	for {
		var value helmv2beta1.HelmRelease
		if err := decoder.Decode(&value); err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		helmReleaseList = append(helmReleaseList, value)
	}

	return helmReleaseList, nil
}

func PatchHelmRelease(existingReleases []helmv2beta1.HelmRelease, patchedHelmRelease helmv2beta1.HelmRelease, index int) (string, error) {
	existingReleases[index] = patchedHelmRelease

	var sb strings.Builder

	for _, r := range existingReleases {
		b, err := kyaml.Marshal(r)
		if err != nil {
			return "", fmt.Errorf("failed to marshal: %w", err)
		}

		sb.WriteString("---\n" + string(b))
	}

	return sb.String(), nil
}
