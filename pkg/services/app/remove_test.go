package app

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
)

var (
	localAddParams AddParams
	removeParams   RemoveParams
	application    wego.Application
	app            models.Application
	fluxDir        string
	goatPaths      map[string]bool
)

type dummyDirEntry struct {
	name string
}

func (d dummyDirEntry) Name() string {
	return d.name
}

func (d dummyDirEntry) IsDir() bool {
	return false
}

func (d dummyDirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

func (d dummyDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

func makeDirEntries(paths map[string]bool) []os.DirEntry {
	results := []os.DirEntry{}

	for path := range paths {
		results = append(results, dummyDirEntry{name: filepath.Base(path)})
	}

	return results
}

// Store the path of a resource tracked in the repo
func storeGOATPath(path string) {
	if strings.HasSuffix(path, "user/kustomization.yaml") {
		return
	}

	goatPaths[path] = true
}

// Stop tracking a stored path; used to ensure after calling remove
// that all paths have been handled
func removeGOATPath(basepath string) error {
	path := filepath.Join(automation.AppYamlDir(app), basepath)

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

// Ensure that every resource that was written to the repository gets removed
func checkRemoveResults() error {
	if len(goatPaths) > 0 {
		return fmt.Errorf("unexpected paths: %#+v", goatPaths)
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

	Context("Removing resources from cluster", func() {
		var _ = BeforeEach(func() {
			Expect(setupFlux()).To(Succeed())

			gitClient.WriteStub = func(path string, manifest []byte) error {
				storeGOATPath(path)
				return nil
			}

			gitClient.RemoveStub = func(path string) error {
				return removeGOATPath(path)
			}

			osysClient.ReadDirStub = func(dirName string) ([]os.DirEntry, error) {
				return makeDirEntries(goatPaths), nil
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
					AppConfigUrl:   "ssh://git@github.com/user/external.git",
					AutoMerge:      true,
				}

				removeParams = RemoveParams{
					Name: "loki",
				}

				goatPaths = map[string]bool{}
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
						AppConfigUrl:   "",
						AutoMerge:      true,
					}

					removeParams = RemoveParams{
						Name: "thor",
					}

					goatPaths = map[string]bool{}
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
