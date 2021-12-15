package gitopswriter

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var (
	application                  wego.Application
	clusterUserKustomization     []byte
	clusterUserKustomizationPath string
	fluxDir                      string
	goatPaths                    map[string]bool
	realFlux                     flux.Flux
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
func removeGOATPath(path string) error {
	if !goatPaths[path] {
		return fmt.Errorf("goat path: %s not found in repository", path)
	}

	delete(goatPaths, path)

	return nil
}

func createRemoveDirWriter() GitOpsDirectoryWriter {
	repoWriter := gitrepo.NewRepoWriter(app.ConfigRepo, gitProviders, gitClient, log)
	automationSvc := automation.NewAutomationGenerator(gitProviders, realFlux, log)

	return NewGitOpsDirectoryWriter(automationSvc, repoWriter, osysClient, log)
}

// Run 'gitops add app' using cluster name of test-cluster and gathers the resources
// we expect to be generated
func runAddAndCollectInfo() error {
	return runAddAndCollectInfoWithClusterName("test-cluster")
}

func runAddAndCollectInfoWithClusterName(clusterName string) error {
	if err := gitOpsDirWriter.AddApplication(context.Background(), app, clusterName, true); err != nil {
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

// See if the cluster kustomization has a reference to the app
func checkClusterKustomizationForApp(name string) error {
	if clusterUserKustomization == nil {
		return errors.New("No cluster user kustomization")
	}

	manifestMap := map[string]interface{}{}

	Expect(yaml.Unmarshal(clusterUserKustomization, &manifestMap)).Should(Succeed())
	r := manifestMap["resources"]

	if r != nil {
		l := manifestMap["resources"].([]interface{})
		for _, a := range l {
			if strings.Contains(a.(string), name) {
				return nil
			}
		}
	}

	return errors.New("Not Found")
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

		var err error

		realFlux, fluxDir, err = testutils.SetupFlux()
		Expect(err).To(BeNil())

		gitOpsDirWriter = createRemoveDirWriter()

	})

	AfterEach(func() {
		Expect(os.RemoveAll(fluxDir)).To(Succeed())
	})

	Context("Removing resources from cluster", func() {
		var _ = BeforeEach(func() {

			gitClient.WriteStub = func(path string, manifest []byte) error {
				storeGOATPath(path)
				if strings.HasSuffix(path, "user/kustomization.yaml") {
					clusterUserKustomization = manifest
					clusterUserKustomizationPath = path
				}
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

			Context("Errors removing an app", func() {

				customError := errors.New("some error")
				BeforeEach(func() {
					app.ConfigRepo = createRepoURL("ssh://git@github.com/user/external.git")
					app.Path = "loki"
				})

				It("fails getting default branch", func() {
					gitProviders.GetDefaultBranchReturns("", customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails cloning config repo", func() {
					gitClient.CloneReturns(false, customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails reading directory", func() {
					osysClient.ReadDirReturns(nil, customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails checking out branch", func() {
					gitClient.CheckoutReturns(customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails removing files using git", func() {
					gitClient.RemoveReturns(customError)

					Expect(runAddAndCollectInfo()).To(Succeed())
					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails writing updates using git", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())

					gitClient.WriteReturns(customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails committing files using git", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())

					gitClient.CommitReturns("", customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})

				It("fails creating pull request", func() {
					Expect(runAddAndCollectInfo()).To(Succeed())

					gitProviders.CreatePullRequestReturns(nil, customError)

					err := gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)
					Expect(err.Error()).To(ContainSubstring(customError.Error()))
				})
			})

			It("removes cluster resources for helm chart from helm repo with configRepo = <url> when auto-merge is true", func() {
				app.ConfigRepo = createRepoURL("ssh://git@github.com/user/external.git")
				app.Path = "loki"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from git repo with configRepo = <url>  when auto-merge is true", func() {
				app.GitSourceURL = createRepoURL("ssh://git@github.com/user/wego-fork-test.git")
				app.ConfigRepo = createRepoURL("ssh://git@github.com/user/external.git")
				app.Path = "./"

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

			It("removes cluster resources for helm chart from helm repo with configURL = <url>  when auto-merge is false", func() {
				app.ConfigRepo = createRepoURL("ssh://git@github.com/user/external.git")
				app.Path = "loki"

				gitProviders.CreatePullRequestCalls(func(ctx2 context.Context, url gitproviders.RepoURL, info gitproviders.PullRequestInfo) (gitprovider.PullRequest, error) {
					Expect(info.SkipAddingFilesOnCreation).To(Equal(true))
					Expect(len(info.Files)).To(Equal(0))
					return NewFakePullRequest(app.ConfigRepo.Owner(), app.ConfigRepo.RepositoryName(), 1), nil
				})

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", false)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})

		})

		Context("Removing resources for non-helm apps when", func() {
			var _ = BeforeEach(func() {
				sourceURL := createRepoURL("ssh://git@github.com/user/wego-fork-test.git")

				app = models.Application{
					Name:           "wego-fork-test",
					Namespace:      wego.DefaultNamespace,
					GitSourceURL:   sourceURL,
					ConfigRepo:     sourceURL,
					Branch:         "main",
					Path:           "./",
					AutomationType: models.AutomationTypeKustomize,
					SourceType:     models.SourceTypeGit,
				}

				goatPaths = map[string]bool{}
			})

			It("removes cluster resources for non-helm app configRepo = ''", func() {
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

			It("removes cluster resources for non-helm app with configRepo = <url>", func() {
				app.GitSourceURL = createRepoURL("ssh://git@github.com/user/wego-fork-test.git")
				app.ConfigRepo = createRepoURL("ssh://git@github.com/user/external.git")

				Expect(runAddAndCollectInfo()).To(Succeed())
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, "test-cluster", true)).To(Succeed())
				Expect(checkRemoveResults()).To(Succeed())
			})
			It("removes cluster resources for non-helm app with configRepo = <url> and eksctl cluster name", func() {

				app.GitSourceURL = createRepoURL("ssh://git@github.com/user/wego-fork-test.git")
				app.ConfigRepo = createRepoURL("ssh://git@github.com/user/external.git")

				cname := "arn:aws:eks:us-west-2:01234567890:cluster/default-my-wego-control-plan"
				Expect(runAddAndCollectInfoWithClusterName(cname)).To(Succeed())
				// Check that a resource was added
				Expect(checkClusterKustomizationForApp(app.Name)).To(Succeed())
				gitClient.CloneStub = func(arg1 context.Context, dir string, arg3 string, arg4 string) (bool, error) {
					if clusterUserKustomization != nil {
						p := filepath.Join(dir, filepath.Dir(clusterUserKustomizationPath))
						Expect(os.MkdirAll(p, 0700)).To(Succeed(), "failed to create git dir")
						Expect(os.WriteFile(filepath.Join(p, "kustomization.yaml"), clusterUserKustomization, 0666)).To(Succeed())
					}

					return false, nil
				}
				Expect(gitOpsDirWriter.RemoveApplication(context.Background(), app, cname, true)).To(Succeed())
				Expect(checkClusterKustomizationForApp(app.Name)).ToNot(Succeed())
			})
		})
	})
})

type fakePullRequest struct {
	pullRequestInfo gitprovider.PullRequestInfo
}

func (fc *fakePullRequest) APIObject() interface{} {
	return &fc.pullRequestInfo
}

func (fc *fakePullRequest) Get() gitprovider.PullRequestInfo {
	return fc.pullRequestInfo
}

func NewFakePullRequest(owner string, repoName string, prNumber int) gitprovider.PullRequest {
	return &fakePullRequest{
		pullRequestInfo: gitprovider.PullRequestInfo{
			WebURL: fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repoName, prNumber),
			Merged: false,
			Number: 1,
		},
	}
}
