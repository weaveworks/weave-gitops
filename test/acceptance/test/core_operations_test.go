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
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/prometheus/common/log"
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
	var tmpPath string

	AfterEach(func() {
		os.RemoveAll(tmpPath)
		Expect(err := deleteRepos()).ShouldNot(HaveOccurred())
	})

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})

		By("Setup test", func() {
			Expect(checkInitialStatus()).ShouldNot(HaveOccurred())
			Expect(setupTest()).ShouldNot(HaveOccurred())
			Expect(ensureWegoRepoIsAbsent()).ShouldNot(HaveOccurred())
			Expect(ensureFluxVersion()).ShouldNot(HaveOccurred())
			Expect(installFlux()).ShouldNot(HaveOccurred())
			Expect(setUpTestRepo()).ShouldNot(HaveOccurred())
		})

	})

	It("Verify add private repo when repo does not already exist", func() {

		By("When i run 'wego add .'", func() {
			dir, err := os.Getwd()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(os.Chdir(tmpDir)).ShouldNot(HaveOccurred())
			Expect(os.Chdir(dir)).ShouldNot(HaveOccurred())
			command := exec.Command(WEGO_BIN_PATH, "add", ".")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then a private repo with name foo-cluster-wego is created on the remote git", func() {
			Expect(ensureWegoRepoExists()).ShouldNot(HaveOccurred())
		})

		By("kubectl get pods -n wego-system should list the source and kustomize controllers", func() {
			Expect(waitForNginxDeployment()).ShouldNot(HaveOccurred())
			command := exec.Command("sh", "-c", utils.Escape("kubectl get pods -n wego-system"))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gbytes.Say("kustomize-controller"))
			Eventually(session).Should(gbytes.Say("source-controller"))
		})
	})

	It("Verify add public repo when repo does not already exist", func() {

		By("When i run 'wego add . --private=false'", func() {
			dir, err := os.Getwd()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(os.Chdir(tmpDir)).ShouldNot(HaveOccurred())
			Expect(os.Chdir(dir)).ShouldNot(HaveOccurred())
			command := exec.Command(WEGO_BIN_PATH, "add", ".", "--private=false")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then a private repo with name is created on the remote git", func() {
			Expect(ensureWegoRepoExists()).ShouldNot(HaveOccurred())
		})

		By("kubectl get pods -n wego-system should list the source and kustomize controllers", func() {
			Expect(waitForNginxDeployment()).ShouldNot(HaveOccurred())
			command := exec.Command("sh", "-c", utils.Escape("kubectl get pods -n wego-system"))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gbytes.Say("kustomize-controller"))
			Eventually(session).Should(gbytes.Say("source-controller"))
		})
	})
})

func setupTest() error {
	WEGO_BIN_PATH = "/usr/local/bin/wego"
	tmpPath, err := ioutil.TempDir("", "tmp-dir")
	if err != nil {
		return err
	}
	tmpDir = tmpPath

	token, _ := os.LookupEnv("GITHUB_TOKEN")
	c, err := github.NewClient(github.WithOAuth2Token(token), github.WithDestructiveAPICalls(true), github.WithConditionalRequests(true))
	if err != nil {
		return err
	}
	client = c
	return nil
}

func ensureWegoRepoIsAbsent() error {
	ctx := context.Background()
	name, err := getWegoRepoName()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
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

func ensureWegoRepoExists() error {
	ctx := context.Background()
	name, err := getRepoName()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return err
	}
	_, err = client.OrgRepositories().Get(ctx, *ref)
	if err != nil {
		return err
	}
	return nil
}

func ensureFluxVersion() error {
	if version.FluxVersion == "undefined" {
		tomlpath, err := filepath.Abs("../../../tools/bin/stoml")
		if err != nil {
			return err
		}
		deppath, err := filepath.Abs("../../../tools/dependencies.toml")
		if err != nil {
			return err
		}
		out, err := utils.CallCommandSilently(fmt.Sprintf("%s %s flux.version", tomlpath, deppath))
		if err != nil {
			return err
		}
		version.FluxVersion = strings.TrimRight(string(out), "\n")
	}
	return nil
}

func waitForNginxDeployment() error {
	for i := 1; i < 61; i++ {
		log.Infof("Waiting for nginx... try: %d of 60\n", i)
		err := utils.CallCommandForEffect("kubectl get deployment nginx -n my-nginx")
		if err == nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("Failed to deploy nginx workload to the cluster")
}

func installWego(t *testing.T) {
	flux.SetupFluxBin()
	manifests, err := cmdimpl.Install(cmdimpl.InstallParamSet{Namespace: "wego-system"})
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

	cloneurl := fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
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

	originurl := fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
	err = utils.CallCommandForEffectWithDebug(fmt.Sprintf("git remote add origin %s && git pull --rebase origin main && git checkout main && git push --set-upstream origin main", originurl))
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
		url := fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
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
		url = fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
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
	//Show all resources
	err := ShowItems("")
	if err != nil {
		log.Infof("Failed to print the pods")
	}
	require.Equal(t, status.Unmodified, status.GetClusterStatus())
}
