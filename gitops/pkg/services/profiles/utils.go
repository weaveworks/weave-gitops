package profiles

import (
	"bytes"
	"io"
	"sort"

	"github.com/Masterminds/semver"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops/gitops/pkg/helm"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// ConvertStringListToSemanticVersionList converts a slice of strings into a slice of semantic version.
func ConvertStringListToSemanticVersionList(versions []string) ([]*semver.Version, error) {
	var result []*semver.Version

	for _, v := range versions {
		ver, err := semver.NewVersion(v)
		if err != nil {
			return nil, err
		}

		result = append(result, ver)
	}

	return result, nil
}

// SortVersions sorts semver versions in decreasing order.
func SortVersions(versions []*semver.Version) {
	sort.SliceStable(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})
}

// SplitHelmReleaseYAML splits a manifest file that contains one or more Helm Releases that may be separated by '---'.
func SplitHelmReleaseYAML(resources []byte) ([]*helmv2beta1.HelmRelease, error) {
	var helmReleaseList []*helmv2beta1.HelmRelease

	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(resources), helm.DefaultBufferSize)

	for {
		var value helmv2beta1.HelmRelease
		if err := decoder.Decode(&value); err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		helmReleaseList = append(helmReleaseList, &value)
	}

	return helmReleaseList, nil
}
