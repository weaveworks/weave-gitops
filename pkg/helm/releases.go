package helm

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
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

// AppendHelmReleaseToFile appends a HelmRelease to a gitProvider CommitFile given a condition, file name and file path.
func AppendHelmReleaseToFile(files []*gitprovider.CommitFile, newRelease *v2beta1.HelmRelease, condition func(r, newRelease v2beta1.HelmRelease) error, fileName, filePath string) (gitprovider.CommitFile, error) {
	var content string

	for _, f := range files {
		if f.Path != nil && *f.Path == filePath {
			if f.Content == nil || *f.Content == "" {
				break
			}

			manifestByteSlice, err := splitYAML([]byte(*f.Content))
			if err != nil {
				return gitprovider.CommitFile{}, fmt.Errorf("error splitting %s: %w", fileName, err)
			}

			for _, manifestBytes := range manifestByteSlice {
				var r v2beta1.HelmRelease
				if err := kyaml.Unmarshal(manifestBytes, &r); err != nil {
					return gitprovider.CommitFile{}, fmt.Errorf("error unmarshaling %s: %w", fileName, err)
				}

				if err := condition(r, *newRelease); err != nil {
					return gitprovider.CommitFile{}, err
				}
			}

			content = *f.Content

			break
		}
	}

	helmReleaseManifest, err := kyaml.Marshal(newRelease)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("failed to marshal new HelmRelease: %w", err)
	}

	content += "\n---\n" + string(helmReleaseManifest)

	return gitprovider.CommitFile{
		Path:    &filePath,
		Content: &content,
	}, nil
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
