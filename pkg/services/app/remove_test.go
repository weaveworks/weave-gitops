package app

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

var removeParams RemoveParams

var application wego.Application

var info *AppResourceInfo

var appResources []ResourceRef

var fluxDir string

var createdResources map[ResourceKind]map[string]bool

var goatPaths map[string]bool

var manifestsByPath map[string][]byte

func populateAppRepo() (string, error) {
	dir, err := ioutil.TempDir("", "an-app-dir")
	if err != nil {
		return "", err
	}

	workloadPath1 := filepath.Join(dir, "kustomize", "one", "path", "to", "files")
	workloadPath2 := filepath.Join(dir, "kustomize", "another", "path", "to", "more", "files")
	if err := os.MkdirAll(workloadPath1, 0777); err != nil {
		return "", err
	}
	if err := os.MkdirAll(workloadPath2, 0777); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(workloadPath1, "nginx.yaml"), []byte("file1"), 0644); err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filepath.Join(workloadPath2, "nginx.yaml"), []byte("file2"), 0644); err != nil {
		return "", err
	}

	return dir, nil
}

// Track all resources created during a "wego app add" so that they can be looked up by "kind" and "name"
func storeCreatedResource(manifestData []byte) error {
	manifests := bytes.Split(manifestData, []byte("\n---\n"))
	for _, manifest := range manifests {
		manifestMap := map[string]interface{}{}

		if err := yaml.Unmarshal(manifest, &manifestMap); err != nil {
			return err
		}

		metamap := manifestMap["metadata"].(map[string]interface{})
		kind := ResourceKind(manifestMap["kind"].(string))
		name := metamap["name"].(string)

		if createdResources[kind] == nil {
			createdResources[kind] = map[string]bool{}
		}

		createdResources[kind][name] = true
	}
	return nil
}

// Remove all tracking for a resource based on its path in the repository
func removeCreatedResourceByPath(path string) error {
	manifest := manifestsByPath[path]
	delete(manifestsByPath, path)
	return removeCreatedResource(manifest)
}

// Remove tracking for a resource given its manifest
func removeCreatedResource(manifestData []byte) error {
	manifests := bytes.Split(manifestData, []byte("\n---\n"))
	for _, manifest := range manifests {
		manifestMap := map[string]interface{}{}

		if err := yaml.Unmarshal(manifest, &manifestMap); err != nil {
			return err
		}

		metamap := manifestMap["metadata"].(map[string]interface{})
		kind := ResourceKind(manifestMap["kind"].(string))

		if createdResources[kind] == nil {
			return fmt.Errorf("expected %s resources to be present", kind)
		}

		delete(createdResources[kind], metamap["name"].(string))
	}
	return nil
}

// Remove tracking for a resource given its name and kind
func removeCreatedResourceByName(name, kindString string) error {
	if kindString == "app" { // figure out later why Application doesn't work
		kindString = "Application"
	}

	kind := ResourceKind(kindString)
	if createdResources[kind] == nil {
		return fmt.Errorf("expected %s resources to be present", kind)
	}

	delete(createdResources[kind], name)
	return nil
}

// Store the path of a resource tracked in the repo
// and associate its manifest with the path for later lookup
func storeGOATPath(path string, manifest []byte) {
	goatPaths[path] = true
	manifestsByPath[path] = manifest
}

// Stop tracking a stored path; used to ensure after calling remove
// that all paths have been handled
func removeGOATPath(path string) error {
	if !goatPaths[path] {
		return fmt.Errorf("goat path: %s not found in repository", path)
	}

	delete(goatPaths, path)
	return nil
}

// Set up a flux binary in a temp dir that will be used to generate flux manifests
func setupFlux() error {
	dir, err := ioutil.TempDir("", "a-home-dir")
	if err != nil {
		return err
	}

	fluxDir = dir
	cliRunner := &runner.CLIRunner{}
	osysClient := &osysfakes.FakeOsys{}
	fluxClient := flux.New(osysClient, cliRunner)
	osysClient.UserHomeDirStub = func() (string, error) {
		return dir, nil
	}
	appSrv.(*App).flux = fluxClient
	fluxBin, err := ioutil.ReadFile(filepath.Join("..", "..", "flux", "bin", "flux"))
	if err != nil {
		return err
	}

	binPath, err := fluxClient.GetBinPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(binPath, 0777)
	if err != nil {
		return err
	}

	exePath, err := fluxClient.GetExePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(exePath, fluxBin, 0777)
	if err != nil {
		return err
	}

	return nil
}

// Run 'wego app add' and gather the resources we expect to be generated
func runAddAndCollectInfo() error {
	params, err := appSrv.(*App).updateParametersIfNecessary(addParams)
	if err != nil {
		return err
	}

	addParams = params
	application = makeWegoApplication(addParams)
	info = getAppResourceInfo(application, "test-cluster")
	appResources = info.clusterResources()

	if err := appSrv.Add(addParams); err != nil {
		return err
	}

	return nil
}

// Make sure that each of the expected resources was created and the expected files were
// written to the repo
func checkAddResults() error {
	for _, res := range appResources {
		resources := createdResources[res.kind]
		if len(resources) == 0 {
			return fmt.Errorf("expected %s resources to be created", res.kind)
		}
		delete(resources, res.name)
	}

	for kind, leftovers := range createdResources {
		if len(leftovers) > 0 {
			return fmt.Errorf("unexpected %s resources: %#+v", kind, leftovers)
		}
	}

	if len(goatPaths) != len(info.clusterResourcePaths()) {
		return fmt.Errorf("expected %d goat paths, found: %d", len(info.clusterResourcePaths()), len(goatPaths))
	}

	for _, path := range info.clusterResourcePaths() {
		delete(goatPaths, path)
	}

	if len(goatPaths) > 0 {
		return fmt.Errorf("unexpected paths: %#+v", goatPaths)
	}

	return nil
}

// Ensure that every resource that was written to the repository gets removed
func checkRemoveResults() error {
	if len(goatPaths) > 0 {
		return fmt.Errorf("unexpected paths: %#+v", goatPaths)
	}

	for _, res := range appResources {
		if res.repositoryPath != "" || res.kind == ResourceKindKustomization || res.kind == ResourceKindHelmRelease {
			resources := createdResources[res.kind]
			if resources[res.name] {
				return fmt.Errorf("expected %s named %s to be removed from the repository", res.kind, res.name)
			}
		}
	}

	return nil
}

var _ = Describe("Remove", func() {
	var _ = BeforeEach(func() {
		application = makeWegoApplication(AddParams{
			Url:            "https://github.com/foo/bar",
			Path:           "./kustomize",
			Branch:         "main",
			Dir:            ".",
			DeploymentType: "kustomize",
			Namespace:      "wego-system",
			AppConfigUrl:   "NONE",
			AutoMerge:      true,
		})
	})

	It("gives a correct error message when app path not found", func() {
		application.Spec.Path = "./badpath"
		appRepoDir, err := populateAppRepo()
		Expect(err).ShouldNot(HaveOccurred())
		defer os.RemoveAll(appRepoDir)
		_, err = findAppManifests(application, appRepoDir)
		Expect(err).Should(MatchError("application path './badpath' not found"))
	})

	It("locates application manifests", func() {
		appRepoDir, err := populateAppRepo()
		Expect(err).ShouldNot(HaveOccurred())
		defer os.RemoveAll(appRepoDir)
		manifests, err := findAppManifests(application, appRepoDir)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(manifests)).To(Equal(2))
		for _, manifest := range manifests {
			Expect(manifest).To(Or(Equal([]byte("file1")), Equal([]byte("file2"))))
		}
	})

	Context("Collecting resources deployed to cluster", func() {
		var _ = BeforeEach(func() {
			Expect(setupFlux()).To(Succeed())

			// Track the resources added to the cluster via files added to the repository
			gitClient.WriteStub = func(path string, manifest []byte) error {
				storeGOATPath(path, manifest)
				return storeCreatedResource(manifest)
			}

			// Track the resources added directly to the cluster
			kubeClient.ApplyStub = func(manifest []byte, namespace string) ([]byte, error) {
				if err := storeCreatedResource(manifest); err != nil {
					return nil, err
				}
				return []byte(""), nil
			}
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		Context("Collecting resources for helm charts", func() {
			var _ = BeforeEach(func() {
				addParams = AddParams{
					Url:            "https://charts.kube-ops.io",
					Branch:         "main",
					DeploymentType: "helm",
					Namespace:      "wego-system",
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				goatPaths = map[string]bool{}
				createdResources = map[ResourceKind]map[string]bool{}
				manifestsByPath = map[string][]byte{}
			})

			It("collects cluster resources for helm chart from helm repo with configURL = NONE", func() {
				addParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = NONE", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for helm chart from helm repo with configURL = <url>", func() {
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})
		})

		Context("Collecting resources for non-helm apps", func() {
			var _ = BeforeEach(func() {
				addParams = AddParams{
					Url:            "ssh://git@github.com/user/wego-fork-test.git",
					Branch:         "main",
					DeploymentType: "kustomize",
					Namespace:      "wego-system",
					Path:           "./",
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				goatPaths = map[string]bool{}
				createdResources = map[ResourceKind]map[string]bool{}
				manifestsByPath = map[string][]byte{}
			})

			It("collects cluster resources for non-helm app with configURL = NONE", func() {
				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for non-helm app configURL = ''", func() {
				addParams.AppConfigUrl = ""

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for non-helm app with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})
		})
	})

	Context("Removing resources from cluster", func() {
		var _ = BeforeEach(func() {
			Expect(setupFlux()).To(Succeed())

			gitClient.WriteStub = func(path string, manifest []byte) error {
				storeGOATPath(path, manifest)
				return storeCreatedResource(manifest)
			}

			gitClient.RemoveStub = func(path string) error {
				if err := removeGOATPath(path); err != nil {
					return err
				}

				return removeCreatedResourceByPath(path)
			}

			kubeClient.ApplyStub = func(manifest []byte, namespace string) ([]byte, error) {
				if err := storeCreatedResource(manifest); err != nil {
					return nil, err
				}
				return []byte(""), nil
			}

			kubeClient.DeleteByNameStub = func(name, kind, namespace string) ([]byte, error) {
				if err := removeCreatedResourceByName(name, kind); err != nil {
					return nil, err
				}
				return []byte(""), nil
			}

			kubeClient.GetApplicationStub = func(_ context.Context, name types.NamespacedName) (*wego.Application, error) {
				return &application, nil
			}
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		Context("Removing resources for helm charts", func() {
			var _ = BeforeEach(func() {
				addParams = AddParams{
					Url:            "https://charts.kube-ops.io",
					Branch:         "main",
					DeploymentType: "helm",
					Namespace:      "wego-system",
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				removeParams = RemoveParams{
					Name: "loki",
				}

				goatPaths = map[string]bool{}
				createdResources = map[ResourceKind]map[string]bool{}
			})

			It("removes cluster resources for helm chart from helm repo with configURL = NONE", func() {
				addParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from git repo with configURL = NONE", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from helm repo with configURL = <url>", func() {
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from git repo with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			Context("Removing resources for non-helm apps", func() {
				var _ = BeforeEach(func() {
					addParams = AddParams{
						Url:            "ssh://git@github.com/user/wego-fork-test.git",
						Branch:         "main",
						DeploymentType: "kustomize",
						Namespace:      "wego-system",
						Path:           "./",
						AppConfigUrl:   "NONE",
						AutoMerge:      true,
					}

					removeParams = RemoveParams{
						Name: "thor",
					}

					goatPaths = map[string]bool{}
					createdResources = map[ResourceKind]map[string]bool{}
				})

				It("removes cluster resources for non-helm app with configURL = NONE", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})

				It("removes cluster resources for non-helm app configURL = ''", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})

				It("removes cluster resources for non-helm app with configURL = <url>", func() {
					addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
					addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})
			})
		})
	})
})
