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
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

const clusterName = "test-cluster"

var (
	localAddParams   AddParams
	removeParams     RemoveParams
	application      wego.Application
	app              models.Application
	appResources     []automation.ResourceRef
	fluxDir          string
	createdResources map[automation.ResourceKind]map[string]bool
	goatPaths        map[string]bool
	manifestsByPath  = map[string][]byte{}
)

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

// Track all resources created during a "gitops add app" so that they can be looked up by "kind" and "name"
func storeCreatedResource(manifestData []byte) error {
	manifests := bytes.Split(manifestData, []byte("\n---\n"))
	for _, manifest := range manifests {
		manifestMap := map[string]interface{}{}

		if err := yaml.Unmarshal(manifest, &manifestMap); err != nil {
			return err
		}

		metamap := manifestMap["metadata"].(map[string]interface{})
		kind := automation.ResourceKind(manifestMap["kind"].(string))
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
		kind := automation.ResourceKind(manifestMap["kind"].(string))

		if createdResources[kind] == nil {
			return fmt.Errorf("expected %s resources to be present", kind)
		}

		delete(createdResources[kind], metamap["name"].(string))
	}

	return nil
}

// Remove tracking for a resource given itsq name and kind
func removeCreatedResourceByName(name string, kind automation.ResourceKind) error {
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
	appSrv.(*AppSvc).Flux = fluxClient
	appSrv.(*AppSvc).Automation = automation.NewAutomationService(gitProviders, fluxClient, log)

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

func updateAppFromParams() error {
	params, err := appSrv.(*AppSvc).updateParametersIfNecessary(context.Background(), gitProviders, localAddParams)
	if err != nil {
		return err
	}

	localAddParams = params

	app, err = makeApplication(localAddParams)
	if err != nil {
		return err
	}

	application = automation.AppToWegoApp(app)
	appResources = automation.ClusterResources(app, clusterName)

	return nil
}

// Run 'gitops add app' and gather the resources we expect to be generated
func runAddAndCollectInfo() error {
	if err := updateAppFromParams(); err != nil {
		return err
	}

	if err := appSrv.Add(gitClient, gitProviders, localAddParams); err != nil {
		return err
	}

	return nil
}

// Make sure that each of the expected resources was created and the expected files were
// written to the repo
func checkAddResults() error {
	for _, res := range appResources {
		if res.Kind != automation.ResourceKindSecret {
			resources := createdResources[res.Kind]
			if len(resources) == 0 {
				return fmt.Errorf("expected %s resources to be created", res.Kind)
			}

			delete(resources, res.Name)
		}
	}

	for kind, leftovers := range createdResources {
		if len(leftovers) > 0 {
			return fmt.Errorf("unexpected %s resources: %#+v", kind, leftovers)
		}
	}

	if len(goatPaths) != len(automation.ClusterResourcePaths(app, clusterName)) {
		return fmt.Errorf("expected %d goat paths, found: %d", len(automation.ClusterResourcePaths(app, clusterName)), len(goatPaths))
	}

	for _, path := range automation.ClusterResourcePaths(app, clusterName) {
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
		if res.RepositoryPath != "" || res.Kind == automation.ResourceKindKustomization || res.Kind == automation.ResourceKindHelmRelease {
			resources := createdResources[res.Kind]
			if resources[res.Name] {
				return fmt.Errorf("expected %s named %s to be removed from the repository", res.Kind, res.Name)
			}
		}
	}

	return nil
}

var _ = Describe("Remove", func() {
	var _ = BeforeEach(func() {
		localAddParams = AddParams{
			Url:            "https://github.com/foo/bar",
			Path:           "./kustomize",
			Branch:         "main",
			DeploymentType: "kustomize",
			Namespace:      wego.DefaultNamespace,
			AppConfigUrl:   "NONE",
			AutoMerge:      true,
		}

		var err error

		a, err := makeApplication(localAddParams)
		Expect(err).ShouldNot(HaveOccurred())

		application = automation.AppToWegoApp(a)

		gitProviders.GetDefaultBranchReturns("main", nil)
	})

	Context("Collecting resources deployed to cluster", func() {
		var _ = BeforeEach(func() {
			Expect(setupFlux()).To(Succeed())

			// Track the resources added to the cluster via files added to the repository
			gitClient.WriteStub = func(path string, manifest []byte) error {
				storeGOATPath(path, manifest)
				err := storeCreatedResource(manifest)
				return err
			}

			// Track the resources added directly to the cluster
			kubeClient.ApplyStub = func(ctx context.Context, manifest []byte, namespace string) error {
				err := storeCreatedResource(manifest)
				return err
			}
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		Context("Collecting resources for helm charts", func() {
			var _ = BeforeEach(func() {
				localAddParams = AddParams{
					Url:            "https://charts.kube-ops.io",
					Branch:         "main",
					DeploymentType: "helm",
					Namespace:      wego.DefaultNamespace,
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				goatPaths = map[string]bool{}
				createdResources = map[automation.ResourceKind]map[string]bool{}
				manifestsByPath = map[string][]byte{}
			})

			It("collects cluster resources for helm chart from helm repo with configURL = NONE", func() {
				localAddParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = NONE", func() {
				localAddParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				localAddParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for helm chart from helm repo with configURL = <url>", func() {
				localAddParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				localAddParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = <url>", func() {
				localAddParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				localAddParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				localAddParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})
		})

		Context("Collecting resources for non-helm apps", func() {
			var _ = BeforeEach(func() {
				localAddParams = AddParams{
					Url:            "ssh://git@github.com/user/wego-fork-test.git",
					Branch:         "main",
					DeploymentType: "kustomize",
					Namespace:      wego.DefaultNamespace,
					Path:           "./",
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				goatPaths = map[string]bool{}
				createdResources = map[automation.ResourceKind]map[string]bool{}
				manifestsByPath = map[string][]byte{}
			})

			It("collects cluster resources for non-helm app with configURL = NONE", func() {
				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for non-helm app configURL = ''", func() {
				localAddParams.AppConfigUrl = ""

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(checkAddResults()).To(Succeed())
			})

			It("collects cluster resources for non-helm app with configURL = <url>", func() {
				localAddParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				localAddParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

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

			kubeClient.ApplyStub = func(ctx context.Context, manifest []byte, namespace string) error {
				return storeCreatedResource(manifest)
			}

			kubeClient.DeleteByNameStub = func(ctx context.Context, name string, resource schema.GroupVersionResource, namespace string) error {
				if err := removeCreatedResourceByName(name, GVRToResourceKind(resource)); err != nil {
					return err
				}
				return nil
			}

			kubeClient.GetApplicationReturns(&application, nil)
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		Context("Removing resources for helm charts", func() {
			var _ = BeforeEach(func() {
				localAddParams = AddParams{
					Url:            "https://charts.kube-ops.io",
					Branch:         "main",
					DeploymentType: "helm",
					Namespace:      wego.DefaultNamespace,
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				removeParams = RemoveParams{
					Name: "loki",
				}

				goatPaths = map[string]bool{}
				createdResources = map[automation.ResourceKind]map[string]bool{}
			})

			It("removes cluster resources for helm chart from helm repo with configURL = NONE", func() {
				localAddParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from git repo with configURL = NONE", func() {
				localAddParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				localAddParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from helm repo with configURL = <url>", func() {
				localAddParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				localAddParams.Chart = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from git repo with configURL = <url>", func() {
				localAddParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				localAddParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				localAddParams.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			Context("Removing resources for non-helm apps", func() {
				var _ = BeforeEach(func() {
					localAddParams = AddParams{
						Url:            "ssh://git@github.com/user/wego-fork-test.git",
						Branch:         "main",
						DeploymentType: "kustomize",
						Namespace:      wego.DefaultNamespace,
						Path:           "./",
						AppConfigUrl:   "NONE",
						AutoMerge:      true,
					}

					removeParams = RemoveParams{
						Name: "thor",
					}

					goatPaths = map[string]bool{}
					createdResources = map[automation.ResourceKind]map[string]bool{}
				})

				It("removes cluster resources for non-helm app with configURL = NONE", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})

				It("removes cluster resources for non-helm app configURL = ''", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})

				It("commits the manifests with remove message", func() {
					localAddParams.AppConfigUrl = ""

					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())

					commit, _ := gitClient.CommitArgsForCall(1)
					Expect(commit.Message).To(Equal(gitopswriter.RemoveCommitMessage))
				})

				It("removes cluster resources for non-helm app with configURL = <url>", func() {
					localAddParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
					localAddParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(appSrv.Remove(gitClient, gitProviders, removeParams)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})
			})
		})
	})
})

func GVRToResourceKind(gvr schema.GroupVersionResource) automation.ResourceKind {
	switch gvr {
	case kube.GVRApp:
		return automation.ResourceKindApplication
	case kube.GVRSecret:
		return automation.ResourceKindSecret
	case kube.GVRGitRepository:
		return automation.ResourceKindGitRepository
	case kube.GVRHelmRepository:
		return automation.ResourceKindHelmRepository
	case kube.GVRHelmRelease:
		return automation.ResourceKindHelmRelease
	default:
		return automation.ResourceKindKustomization
	}
}
