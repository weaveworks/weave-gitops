// +build !unittest
// +build !smoke
// +build !acceptance

package acceptance

// Runs basic WeGO operations against a kind cluster.

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

const nginxDeployment = `apiVersion: v1
kind: Namespace
metadata:
  name: my-nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: my-nginx
  labels:
    name: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      name: nginx
  template:
    metadata:
      namespace: my-nginx
      labels:
        name: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
`

var (
	tmpDir string
	client gitprovider.Client
)

var _ = Describe("WEGO Acceptance Tests", func() {

	var session *gexec.Session
	var err error

	WEGO_BIN_PATH = "/usr/local/bin/wego"
	tmpPath, _ := ioutil.TempDir("", "tmp-dir")
	tmpDir = tmpPath
	token, _ := os.LookupEnv("GITHUB_TOKEN")
	c, _ := github.NewClient(github.WithOAuth2Token(token), github.WithDestructiveAPICalls(true), github.WithConditionalRequests(true))
	client = c

	AfterSuite(func() {
		os.RemoveAll(tmpPath)
		deleteRepos()
	})

	BeforeEach(func() {

		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify add repo when repo does not already exist", func() {

		By("Setup Test Repo", func() {
			err := setUpTestRepo()
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Ensuring wego repo does not exist...", func() {
			err := ensureWegoRepoIsAbsent()
			Expect(err).ShouldNot(HaveOccurred())

		})

		By("When i run 'wego add .'", func() {
			command := exec.Command(WEGO_BIN_PATH, "add .", "--private=false")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())

		})

		By("Then a private repo with name foo-cluster-wego is created on the remote git", func() {
			access, err := getRepoAccess()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(access).To(Equal("Private"))
		})

		By("And branch name", func() {
			Eventually(session).Should(gbytes.Say("Branch: main|HEAD\n"))
		})

		// By()
	})
})

func addRepo(t *testing.T) {
	keyFilePath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		key := os.Getenv("GITHUB_KEY")
		tmpFile, err := ioutil.TempFile("", "keyfile")
		require.NoError(t, err)
		defer tmpFile.Close()
		require.NoError(t, ioutil.WriteFile(tmpFile.Name(), []byte(key), 600))
		keyFilePath = tmpFile.Name()
	}

	dir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		require.NoError(t, os.Chdir(dir))
	}()

	cmdimpl.Add([]string{"."}, cmdimpl.AddParamSet{Name: "", Url: "", Path: "./", Branch: "main", PrivateKey: keyFilePath, IsPrivate: true})
}

func ensureWegoRepoIsAbsent() error {
	ctx := context.Background()
	name, err := getWegoRepoName()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://github.com/wkp-example-org/%s", name)
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return err
	}
	repo, err := client.OrgRepositories().Get(ctx, *ref)
	if err != nil {
		log.Info("Repo already deleted")
	} else {
		return repo.Delete(ctx)
	}
	clusterName, err := status.GetClusterName()
	if err != nil {
		return err
	}
	repoName := clusterName + "-wego"
	os.RemoveAll(fmt.Sprintf("%s/.wego/repositories/%s", os.Getenv("HOME"), repoName))
	return nil
}

func getRepoAccess() (string, error) {
	ctx := context.Background()
	name, err := getWegoRepoName()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://github.com/wkp-example-org/%s", name)
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return "", err
	}
	repo, err := client.OrgRepositories().Get(ctx, *ref)
	if err != nil {
		return "", err
	}
	name, err = getRepoName()
	if err != nil {
		return "", err
	}
	access, err := repo.TeamAccess().Get(ctx, name)
	if err != nil {
		return "", err
	}
	fmt.Println(access.Get().Permission)
	return string(*access.Get().Permission), nil
}

func ensureFluxVersion(t *testing.T) {
	if version.FluxVersion == "undefined" {
		tomlpath, err := filepath.Abs("../../../tools/bin/stoml")
		require.NoError(t, err)
		deppath, err := filepath.Abs("../../../tools/dependencies.toml")
		require.NoError(t, err)
		out, err := utils.CallCommandSilently(fmt.Sprintf("%s %s flux.version", tomlpath, deppath))
		require.NoError(t, err)
		version.FluxVersion = strings.TrimRight(string(out), "\n")
	}
}

func waitForNginxDeployment(t *testing.T) {
	for i := 1; i < 61; i++ {
		log.Infof("Waiting for nginx... try: %d of 60\n", i)
		err := utils.CallCommandForEffect("kubectl get deployment nginx -n my-nginx")
		if err == nil {
			return
		}
		time.Sleep(5 * time.Second)
	}
	require.FailNow(t, "Failed to deploy nginx workload to the cluster")
}

func installFlux(t *testing.T) {
	flux.SetupFluxBin()
	manifests, err := fluxops.QuietInstall("wego-system")
	require.NoError(t, err)
	require.NoError(t, utils.CallCommandForEffectWithInputPipeAndDebug("kubectl apply -f -", string(manifests)))
}

func getWegoRepoName() (string, error) {
	repoName, err := fluxops.GetRepoName()
	return repoName, err
}

func getRepoName() (string, error) {
	name, err := getWegoRepoName()
	return name + "-" + filepath.Base(tmpDir), err
}

func setUpTestRepo() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		return err
	}
	_, err = utils.CallCommand("git init")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("nginx.yaml", []byte(nginxDeployment), 0666)
	if err != nil {
		return err
	}
	err = utils.CallCommandForEffectWithDebug("git add nginx.yaml && git commit -m'Added workload'")
	if err != nil {
		return err
	}
	name, err := getRepoName()
	if err != nil {
		return err
	}
	cloneurl := fmt.Sprintf("https://github.com/wkp-example-org/%s", name)
	ref, err := gitprovider.ParseOrgRepositoryURL(cloneurl)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = client.OrgRepositories().Create(ctx, *ref, gitprovider.RepositoryInfo{
		Description: gitprovider.StringVar("test repo"),
	}, &gitprovider.RepositoryCreateOptions{
		AutoInit:        gitprovider.BoolVar(true),
		LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
	})

	if err != nil {
		return err
	}
	name, err = getRepoName()
	if err != nil {
		return err
	}
	originurl := fmt.Sprintf("ssh://git@github.com/wkp-example-org/%s", name)
	err = utils.CallCommandForEffectWithDebug(fmt.Sprintf("git remote add origin %s && git pull --rebase origin main && git push --set-upstream origin main", originurl))
	if err != nil {
		return err
	}
	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	return nil
}

func deleteRepos() error {
	clusterName, err := status.GetClusterName()
	if err == nil {
		ctx := context.Background()
		name, err := getRepoName()
		if err != nil {
			return err
		}
		url := fmt.Sprintf("https://github.com/wkp-example-org/%s", name)
		ref, err := gitprovider.ParseOrgRepositoryURL(url)
		if err != nil {
			return err
		}
		repo, err := client.OrgRepositories().Get(ctx, *ref)
		if err != nil {
			return err
		}
		err = repo.Delete(ctx)
		if err != nil {
			return err
		}
		name, err = getWegoRepoName()
		if err != nil {
			return err
		}
		url = fmt.Sprintf("https://github.com/wkp-example-org/%s", name)
		ref, err = gitprovider.ParseOrgRepositoryURL(url)
		if err != nil {
			return err
		}
		repo, err = client.OrgRepositories().Get(ctx, *ref)
		if err != nil {
			return err
		}
		err = repo.Delete(ctx)
		if err != nil {
			return err
		}
		repoName := clusterName + "-wego"
		os.RemoveAll(fmt.Sprintf("%s/.wego/repositories/%s", os.Getenv("HOME"), repoName))
	} else {
		log.Info("Failed to delete repository")
	}
	return nil
}

func checkInitialStatus(t *testing.T) {
	require.Equal(t, status.Unmodified, status.GetClusterStatus())
}
