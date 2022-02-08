package helm_test

import (
	"fmt"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/models"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
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

var _ = Describe("AppendHelmReleaseToFile", func() {
	var (
		newRelease   *v2beta1.HelmRelease
		existingFile *gitprovider.CommitFile
		path         string
		content      string
	)

	BeforeEach(func() {
		newRelease = helm.MakeHelmRelease(
			"podinfo", "6.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
		path = git.GetProfilesPath("prod", models.WegoProfilesPath)
	})

	When("the file does not exist", func() {
		It("creates one with the new helm release", func() {
			file, err := helm.AppendHelmReleaseToFile(makeTestFiles(), newRelease, makeTestCondition, "profiles.yaml", path)
			Expect(err).NotTo(HaveOccurred())
			r, err := yaml.Marshal(newRelease)
			Expect(err).NotTo(HaveOccurred())
			Expect(*file.Content).To(ContainSubstring(string(r)))
		})
	})

	When("the file exists", func() {
		When("the given condition succeeds", func() {
			It("appends the release to the manifest", func() {
				existingRelease := helm.MakeHelmRelease(
					"podinfo", "6.0.1", "prod", "weave-system",
					types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
				)
				r, _ := yaml.Marshal(existingRelease)
				content = string(r)
				file, err := helm.AppendHelmReleaseToFile([]*gitprovider.CommitFile{{
					Path:    &path,
					Content: &content,
				}}, newRelease, makeTestCondition, "profiles.yaml", path)
				Expect(err).NotTo(HaveOccurred())
				Expect(*file.Content).To(ContainSubstring(string(r)))
			})
		})

		When("the given condition returns an error", func() {
			It("fails to add the profile", func() {
				existingRelease, _ := yaml.Marshal(newRelease)
				content = string(existingRelease)
				existingFile = &gitprovider.CommitFile{
					Path:    &path,
					Content: &content,
				}
				_, err := helm.AppendHelmReleaseToFile([]*gitprovider.CommitFile{existingFile}, newRelease, makeTestCondition, "profiles.yaml", path)
				Expect(err).To(MatchError("err"))
			})
		})

		It("fails if the manifest contains a resource that is not a HelmRelease", func() {
			content = "content"
			_, err := helm.AppendHelmReleaseToFile([]*gitprovider.CommitFile{{
				Path:    &path,
				Content: &content,
			}}, newRelease, makeTestCondition, "profiles.yaml", path)
			Expect(err).To(MatchError("error unmarshaling profiles.yaml: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v2beta1.HelmRelease"))
		})
	})
})

func makeTestCondition(r, newRelease v2beta1.HelmRelease) error {
	if r.Spec.Chart.Spec.Version == newRelease.Spec.Chart.Spec.Version {
		return fmt.Errorf("err")
	}

	return nil
}

func makeTestFiles() []*gitprovider.CommitFile {
	path0 := ".weave-gitops/clusters/prod/system/wego-system.yaml"
	content0 := "machine1 yaml content"
	path1 := ".weave-gitops/clusters/prod/system/podinfo-helm-release.yaml"
	content1 := "machine2 yaml content"

	files := []gitprovider.CommitFile{
		{
			Path:    &path0,
			Content: &content0,
		},
		{
			Path:    &path1,
			Content: &content1,
		},
	}

	commitFiles := make([]*gitprovider.CommitFile, 0)
	for _, file := range files {
		commitFiles = append(commitFiles, &gitprovider.CommitFile{
			Path:    file.Path,
			Content: file.Content,
		})
	}

	return commitFiles
}
