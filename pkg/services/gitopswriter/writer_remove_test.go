package gitopswriter

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
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	application wego.Application
	realFlux    flux.Flux
	fluxDir     string
	goatPaths   map[string]bool
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
	realFlux = flux.New(osysClient, cliRunner)
	osysClient.UserHomeDirStub = func() (string, error) {
		return dir, nil
	}

	fluxBin, err := ioutil.ReadFile(filepath.Join("..", "..", "flux", "bin", "flux"))
	if err != nil {
		return err
	}

	binPath, err := realFlux.GetBinPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(binPath, 0777)
	if err != nil {
		return err
	}

	exePath, err := realFlux.GetExePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(exePath, fluxBin, 0777)
	if err != nil {
		return err
	}

	return nil
}

func createRemoveDirWriter() GitOpsDirectoryWriter {
	repoWriter := gitrepo.NewRepoWriter(app.ConfigURL, gitProviders, gitClient, log)
	automationSvc := automation.NewAutomationService(gitProviders, realFlux, log)

	return NewGitOpsDirectoryWriter(automationSvc, repoWriter, osysClient, log)
}

// Run 'gitops add app' and gather the resources we expect to be generated
func runAddAndCollectInfo() error {
	application = automation.AppToWegoApp(app)

	if err := gitOpsDirWriter.AddApplication(context.Background(), app, "test-cluster", true); err != nil {
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
	BeforeEach(func() {
		app = models.Application{
			Name:                "bar",
			Namespace:           wego.DefaultNamespace,
			HelmSourceURL:       "",
			GitSourceURL:        createRepoURL("ssh://git@github.com/foo/bar.git"),
			Branch:              "main",
			Path:                "./kustomize",
			AutomationType:      models.AutomationTypeKustomize,
			SourceType:          models.SourceTypeGit,
			HelmTargetNamespace: "",
		}

		application = automation.AppToWegoApp(app)

		gitProviders.GetDefaultBranchReturns("main", nil)

		Expect(setupFlux()).To(Succeed())

		gitOpsDirWriter = createRemoveDirWriter()

	})

	AfterEach(func() {
		os.RemoveAll(fluxDir)
	})

	Context("Removing resources from cluster", func() {
		var _ = BeforeEach(func() {

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

		Context("Removing resources for helm charts", func() {
			var _ = BeforeEach(func() {
				app = models.Application{
					Name:                "loki",
					Namespace:           wego.DefaultNamespace,
					HelmSourceURL:       "https://charts.kube-ops.io",
					Branch:              "main",
					Path:                "./kustomize",
					AutomationType:      models.AutomationTypeHelm,
					SourceType:          models.SourceTypeHelm,
					HelmTargetNamespace: "",
				}

				goatPaths = map[string]bool{}
			})

			It("removes cluster resources for helm chart from helm repo with configURL = <url>", func() {
				app.ConfigURL = createRepoURL("ssh://git@github.com/user/external.git")
				app.Path = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from git repo with configURL = <url>", func() {
				app.GitSourceURL = createRepoURL("ssh://git@github.com/user/wego-fork-test.git")
				app.ConfigURL = createRepoURL("ssh://git@github.com/user/external.git")
				app.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			Context("Removing resources for non-helm apps", func() {
				var _ = BeforeEach(func() {
					sourceURL := createRepoURL("ssh://git@github.com/user/wego-fork-test.git")

					app = models.Application{
						Name:           "wego-fork-test",
						Namespace:      wego.DefaultNamespace,
						GitSourceURL:   sourceURL,
						ConfigURL:      sourceURL,
						Branch:         "main",
						Path:           "./",
						AutomationType: models.AutomationTypeKustomize,
						SourceType:     models.SourceTypeGit,
					}

					goatPaths = map[string]bool{}
				})

				It("removes cluster resources for non-helm app configURL = ''", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())
				})

				It("commits the manifests with remove message", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
					Expect(checkRemoveResults()).To(Succeed())

					commit, _ := gitClient.CommitArgsForCall(1)
					Expect(commit.Message).To(Equal(RemoveCommitMessage))
				})

				It("removes cluster resources for non-helm app with configURL = <url>", func() {
					app.GitSourceURL = createRepoURL("ssh://git@github.com/user/wego-fork-test.git")
					app.ConfigURL = createRepoURL("ssh://git@github.com/user/external.git")

					Expect(runAddAndCollectInfo()).To(Succeed())
					Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
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
