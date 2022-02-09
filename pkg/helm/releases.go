package helm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

// FindHelmReleaseInString finds all HelmRelease(s) in a given string that can be split into YAML.
func FindHelmReleaseInString(s string, newRelease *v2beta1.HelmRelease) ([]v2beta1.HelmRelease, error) {
	manifestByteSlice, err := splitYAML([]byte(s))
	if err != nil {
		return nil, fmt.Errorf("error splitting into YAML: %w", err)
	}

	found := []v2beta1.HelmRelease{}

	for _, manifestBytes := range manifestByteSlice {
		var r v2beta1.HelmRelease
		if err := kyaml.Unmarshal(manifestBytes, &r); err != nil {
			return nil, fmt.Errorf("error unmarshaling: %w", err)
		}

		if profileIsInstalled(r, *newRelease) {
			found = append(found, r)
		}
	}

	return found, nil
}

// AppendHelmReleaseToString appends a HelmRelease to a string.
func AppendHelmReleaseToString(content string, newRelease *v2beta1.HelmRelease) (string, error) {
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

// splitYAML splits a manifest file that may contain multiple YAML resources separated by '---'
// and validates that each element is YAML.
func splitYAML(resources []byte) ([][]byte, error) {
	var splitResources [][]byte

	decoder := yaml.NewDecoder(bytes.NewReader(resources))

	for {
		var value interface{}
		if err := decoder.Decode(&value); err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}

		splitResources = append(splitResources, valueBytes)
	}

	return splitResources, nil
}

func profileIsInstalled(r, newRelease v2beta1.HelmRelease) bool {
	return r.Name == newRelease.Name && r.Namespace == newRelease.Namespace && r.Spec.Chart.Spec.Version == newRelease.Spec.Chart.Spec.Version
}
