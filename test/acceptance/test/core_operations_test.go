// +build !unittest

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
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
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

// Run core operations and check status
func TestCoreOperations(t *testing.T) {
	tmpPath, err := ioutil.TempDir("", "tmp-dir")
	log.Infof("Using temp directory: %s", tmpPath)

	require.NoError(t, err)
	defer os.RemoveAll(tmpPath)
	tmpDir = tmpPath

	log.Info("Creating GitHub client...")
	token, found := os.LookupEnv("GITHUB_TOKEN")
	require.True(t, found)
	c, err := github.NewClient(github.WithOAuth2Token(token), github.WithDestructiveAPICalls(true))
	require.NoError(t, err)
	client = c
	log.Info("Ensuring flux version is set...")
	ensureFluxVersion(t)
	log.Info("Checking initial status...")
	checkInitialStatus(t)
	log.Info("Install flux...")
	installFlux(t)
	log.Info("Setting up test repository...")
	setUpTestRepo(t) // create repo with simple nginx manifest
	defer deleteRepos(t)
	log.Info("Adding test repository to cluster...")
	require.NoError(t, err)
	err = addRepo(t) // add new repo to cluster
	require.NoError(t, err)
	log.Info("Waiting for workload to start...")
	waitForNginxDeployment(t)
}

func addRepo(t *testing.T) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s add .", wegoBinaryPath(t)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	return cmd.Run()
}

func ensureFluxVersion(t *testing.T) {
	if version.FluxVersion == "undefined" {
		out, err := utils.CallCommandSilently("../../../tools/bin/stoml ../../../tools/dependencies.toml flux.version")
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
	manifests, err := fluxops.QuietInstall("wego-system")
	require.NoError(t, err)
	require.NoError(t, utils.CallCommandForEffectWithInputPipeAndDebug("kubectl apply -f -", string(manifests)))
}

func getWegoRepoName(t *testing.T) string {
	repoName, err := fluxops.GetRepoName()
	require.NoError(t, err)
	return repoName
}

func getRepoName(t *testing.T) string {
	return getWegoRepoName(t) + "-" + filepath.Base(tmpDir)
}

func setUpTestRepo(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		require.NoError(t, os.Chdir(dir))
	}()
	_, err = utils.CallCommand("git init")
	require.NoError(t, err)
	err = ioutil.WriteFile("nginx.yaml", []byte(nginxDeployment), 0666)
	require.NoError(t, err)
	err = utils.CallCommandForEffect("git add nginx.yaml && git commit -m'Added workload'")
	require.NoError(t, err)
	url := fmt.Sprintf("https://github.com/wkp-example-org/%s", getRepoName(t))
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	require.NoError(t, err)
	ctx := context.Background()
	_, err = client.OrgRepositories().Create(ctx, *ref, gitprovider.RepositoryInfo{
		Description: gitprovider.StringVar("test repo"),
	}, &gitprovider.RepositoryCreateOptions{
		AutoInit:        gitprovider.BoolVar(true),
		LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
	})

	require.NoError(t, err)
	err = utils.CallCommandForEffectWithDebug(fmt.Sprintf("git remote add origin %s && git pull --rebase origin main && git push --set-upstream origin main", url))
	require.NoError(t, err)
}

func deleteRepos(t *testing.T) {
	clusterName, err := status.GetClusterName()
	if err == nil {
		ctx := context.Background()
		url := fmt.Sprintf("https://github.com/wkp-example-org/%s", getRepoName(t))
		ref, err := gitprovider.ParseOrgRepositoryURL(url)
		require.NoError(t, err)
		repo, err := client.OrgRepositories().Get(ctx, *ref)
		require.NoError(t, err)
		require.NoError(t, repo.Delete(ctx))
		url = fmt.Sprintf("https://github.com/wkp-example-org/%s", getWegoRepoName(t))
		ref, err = gitprovider.ParseOrgRepositoryURL(url)
		require.NoError(t, err)
		repo, err = client.OrgRepositories().Get(ctx, *ref)
		require.NoError(t, err)
		require.NoError(t, repo.Delete(ctx))
		repoName := clusterName + "-wego"
		os.RemoveAll(fmt.Sprintf("%s/.wego/repositories/%s", os.Getenv("HOME"), repoName))
	} else {
		log.Info("Failed to delete repository")
	}
}

func checkInitialStatus(t *testing.T) {
	require.Equal(t, status.GetClusterStatus(), status.Unmodified)
}

func wegoBinaryPath(t *testing.T) string {
	path, err := filepath.Abs("../../../bin/wego")
	require.NoError(t, err)
	return path
}
