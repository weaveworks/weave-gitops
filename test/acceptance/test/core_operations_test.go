// +build !unittest
// +build smoke acceptance

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
	tmpDir      string
	client      gitprovider.Client
	session     *gexec.Session
	tmpPath     string
	err         error
	keyFilePath string
)

var _ = Describe("WEGO Acceptance Tests", func() {

	AfterEach(func() {
		os.RemoveAll(tmpPath)
		Expect(deleteRepos()).Should(Succeed())
	})

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})

		By("Setup test", func() {
			Expect(setupTest()).Should(Succeed())
			Expect(ensureWegoRepoIsAbsent()).Should(Succeed())
			Expect(ensureFluxVersion()).Should(Succeed())
			Expect(installWego()).Should(Succeed())
			Expect(waitForFluxInstall()).Should(Succeed())
			Expect(setUpTestRepo()).Should(Succeed())
		})

	})

	It("Verify add private repo when repo does not already exist", func() {

		By("When i run 'wego add .'", func() {
			dir, err := os.Getwd()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(os.Chdir(tmpDir)).Should(Succeed())
			defer func() {
				Expect(os.Chdir(dir)).Should(Succeed())
			}()
			command := exec.Command(WEGO_BIN_PATH, "add", ".", fmt.Sprintf("--private-key=%v", keyFilePath))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then a private repo with name foo-cluster-wego is created on the remote git", func() {
			access, err := ensureWegoRepoAccess()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(*access).Should(Equal(gitprovider.RepositoryVisibilityPrivate))
		})

		By("kubectl get pods -n wego-system should list the source and kustomize controllers", func() {
			Expect(waitForNginxDeployment()).Should(Succeed())
			Expect(runCommandForGinkgo("kubectl get pods -n wego-system")).Should(Succeed())
			Eventually(session).Should(gbytes.Say("kustomize-controller"))
			Eventually(session).Should(gbytes.Say("source-controller"))
		})
	})

	It("Verify add public repo when repo does not already exist", func() {

		By("When i run 'wego add . --private=false'", func() {
			dir, err := os.Getwd()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(os.Chdir(tmpDir)).Should(Succeed())
			defer func() {
				Expect(os.Chdir(dir)).Should(Succeed())
			}()
			command := exec.Command(WEGO_BIN_PATH, "add", ".", "--private=false", fmt.Sprintf("--private-key=%v", keyFilePath))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then a private repo with name is created on the remote git", func() {
			access, err := ensureWegoRepoAccess()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(*access).Should(Equal(gitprovider.RepositoryVisibilityPublic))
		})

		By("kubectl get pods -n wego-system should list the source and kustomize controllers", func() {
			Expect(waitForNginxDeployment()).Should(Succeed())
			Expect(runCommandForGinkgo("kubectl get pods -n wego-system")).Should(Succeed())
			Eventually(session).Should(gbytes.Say("kustomize-controller"))
			Eventually(session).Should(gbytes.Say("source-controller"))
		})
	})
})

func setupTest() error {
	tmpPath, err = ioutil.TempDir("", "tmp-dir")
	if err != nil {
		return err
	}
	tmpDir = tmpPath

	keyFilePath = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		key := os.Getenv("GITHUB_KEY")
		tmpFile, err := ioutil.TempFile("", "keyfile")
		if err != nil {
			return err
		}
		defer tmpFile.Close()
		err = ioutil.WriteFile(tmpFile.Name(), []byte(key), 600)
		keyFilePath = tmpFile.Name()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(keyFilePath, []byte(key), 700)
		if err != nil {
			return err
		}
	}

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

func ensureWegoRepoAccess() (*gitprovider.RepositoryVisibility, error) {
	ctx := context.Background()
	name, err := getWegoRepoName()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return nil, err
	}

	for i := 1; i < 5; i++ {
		log.Infof("Waiting for wego repo creation... try: %d of 5\n", i)
		repo, err := client.OrgRepositories().Get(ctx, *ref)
		if err == nil {
			return repo.Get().Visibility, nil
		}
		time.Sleep(5 * time.Second)
	}
	return nil, fmt.Errorf("wepo does not exist")
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

func waitForFluxInstall() error {
	for i := 1; i < 11; i++ {
		log.Infof("Waiting for flux... try: %d of 10\n", i)
		if status.GetClusterStatus() == status.FluxInstalled {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("Failed to install flux")
}

func installWego() error {
	flux.SetupFluxBin()
	manifests, err := cmdimpl.Install(cmdimpl.InstallParamSet{Namespace: "wego-system"})
	if err != nil {
		return err
	}
	return utils.CallCommandForEffectWithInputPipeAndDebug("kubectl apply -f -", string(manifests))
}

func getWegoRepoName() (string, error) {
	return fluxops.GetRepoName()
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
	defer func() {
		err = os.Chdir(dir)
	}()

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

	originurl := fmt.Sprintf("ssh://git@github.com/%s/%s", os.Getenv("GITHUB_ORG"), name)
	err = utils.CallCommandForEffectWithDebug(fmt.Sprintf("git remote add origin %s && git pull --rebase origin main && git checkout main && git push --set-upstream origin main", originurl))
	if err != nil {
		return err
	}

	return err
}

func runCommandForGinkgo(cmd string) error {
	command := exec.Command("sh", "-c", utils.Escape(cmd))
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	return err
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
