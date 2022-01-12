package gitopswriter

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/testutils"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	configURL     gitproviders.RepoURL
	cluster       models.Cluster
	namespace     string
	systemPath    string
	userPath      string
	expectedFiles map[string][]byte
)

func systemQualifiedPath(relativePath string) string {
	return filepath.Join(systemPath, relativePath)
}

func userQualifiedPath(relativePath string) string {
	return filepath.Join(userPath, relativePath)
}

var _ = Describe("Associate Cluster", func() {
	var _ = BeforeEach(func() {
		ctx = context.Background()
		cluster = models.Cluster{Name: "cluster-name"}
		configURL = createRepoURL("ssh://git@github.com/foo/bar.git")
		namespace = "test-namespace"
		gitOpsDirWriter = createDirWriter()
		systemPath = filepath.Join(git.WegoRoot, git.WegoClusterDir, cluster.Name, git.WegoClusterOSWorkloadDir)
		userPath = filepath.Join(git.WegoRoot, git.WegoClusterDir, cluster.Name, git.WegoClusterUserWorkloadDir)

		gitProviders.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

		ca, err := gitOpsDirWriter.(*gitOpsDirectoryWriterSvc).Automation.GenerateClusterAutomation(ctx, cluster, configURL, namespace)
		Expect(err).ShouldNot(HaveOccurred())

		configManifest, err := ca.GenerateWegoConfigManifest(cluster.Name, "flux-system", namespace)
		Expect(err).ShouldNot(HaveOccurred())

		expectedFiles = map[string][]byte{
			systemQualifiedPath(automation.AppCRDPath):              ca.AppCRD.Content,
			systemQualifiedPath(automation.RuntimePath):             ca.GitOpsRuntime.Content,
			systemQualifiedPath(automation.SourcePath):              ca.SourceManifest.Content,
			systemQualifiedPath(automation.SystemKustomizationPath): ca.SystemKustomizationManifest.Content,
			systemQualifiedPath(automation.SystemKustResourcePath):  ca.SystemKustResourceManifest.Content,
			systemQualifiedPath(automation.UserKustResourcePath):    ca.UserKustResourceManifest.Content,
			systemQualifiedPath(automation.WegoAppPath):             ca.WegoAppManifest.Content,
			systemQualifiedPath(automation.WegoConfigPath):          configManifest.Content,
			userQualifiedPath(".keep"):                              []byte(strconv.AppendQuote(nil, "# keep")),
		}

		gitClient.OpenStub = func(s string) (*gogit.Repository, error) {
			r, err := gogit.Init(memory.NewStorage(), memfs.New())
			Expect(err).ShouldNot(HaveOccurred())

			_, err = r.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{"git@github.com:foo/bar.git"},
			})
			Expect(err).ShouldNot(HaveOccurred())
			return r, nil
		}
	})

	Describe("Passes in correct files for pull request", func() {
		It("associates a cluster", func() {
			err := gitOpsDirWriter.AssociateCluster(ctx, cluster, configURL, namespace, "flux-system", false)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CloneCallCount()).To(Equal(0)) // PR case
			_, _, info := gitProviders.CreatePullRequestArgsForCall(0)

			Expect(info.CommitMessage).To(Equal(ClusterCommitMessage))
			Expect(info.NewBranch).To(Equal(automation.GetClusterHash(cluster)))
			Expect(len(info.Files)).To(Equal(len(expectedFiles)))

			for _, f := range info.Files {
				Expect(expectedFiles[*f.Path]).To(Equal([]byte(*f.Content)))
			}
		})
	})

	Describe("Passes in correct files for auto-merge", func() {
		It("associates a cluster", func() {
			err := gitOpsDirWriter.AssociateCluster(ctx, cluster, configURL, namespace, "flux-system", true)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CloneCallCount()).To(Equal(1))
			Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(0))
			Expect(gitClient.WriteCallCount()).To(Equal(len(expectedFiles)))

			for i := 0; i < gitClient.WriteCallCount(); i++ {
				path, content := gitClient.WriteArgsForCall(i)

				Expect(expectedFiles[path]).To(Equal(content))
			}
		})
	})
})
