// +build !unittest

package acceptance

// Runs basic WeGO operations against a kind cluster.

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

var tmpDir string

// Run core operations and check status
func TestCoreOperations(t *testing.T) {
	tmpPath, err := ioutil.TempDir("", "tmp-dir")
	log.Infof("Using temp directory: %s", tmpPath)

	require.NoError(t, err)
	defer os.RemoveAll(tmpPath)
	tmpDir = tmpPath
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
	manifests, err := fluxops.Install("wego-system")
	require.NoError(t, err)
	require.NoError(t, utils.CallCommandForEffectWithInputPipeAndDebug("kubectl apply -f -", string(manifests)))
}

func getWegoRepoName(t *testing.T) string {
	repoName, err := fluxops.GetRepoName()
	require.NoError(t, err)
	return repoName
}

func getRepoName(t *testing.T) string {
	return getWegoRepoName(t) + "-" + tmpDir
}

func setUpTestRepo(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(dir)
	_, err = utils.CallCommand("git init")
	require.NoError(t, err)
	err = ioutil.WriteFile("nginx.yaml", []byte(nginxDeployment), 0666)
	require.NoError(t, err)
	err = utils.CallCommandForEffect("git add nginx.yaml && git commit -m'Added workload'")
	require.NoError(t, err)
	_, err = utils.CallCommand(fmt.Sprintf("hub create %s/%s", getOwner(t), getRepoName(t)))
	require.NoError(t, err)
	err = utils.CallCommandForEffect("git push -u origin main")
	require.NoError(t, err)
}

func deleteRepos(t *testing.T) {
	org := getOwner(t)
	clusterName, err := status.GetClusterName()
	if err == nil {
		repoName := clusterName + "-wego"
		cmdstr := fmt.Sprintf("hub delete -y %s/%s", org, repoName)
		_ = utils.CallCommandForEffect(cmdstr) // there's nothing we can do with the error
		os.RemoveAll(fmt.Sprintf("%s/.wego/repositories/%s", os.Getenv("HOME"), repoName))
		_ = utils.CallCommandForEffect(fmt.Sprintf("hub delete -y %s/%s", getOwner(t), getRepoName(t)))
		_ = utils.CallCommandForEffect(fmt.Sprintf("hub delete -y %s/%s", getOwner(t), getWegoRepoName(t)))
	} else {
		log.Info("Failed to delete repository")
	}
}

func getOwner(t *testing.T) string {
	owner, err := fluxops.GetOwnerFromEnv()
	require.NoError(t, err)
	return owner
}

func checkInitialStatus(t *testing.T) {
	require.Equal(t, status.GetClusterStatus(), status.Unmodified)
}

func getTestDir(t *testing.T) string {
	testDir, err := os.Getwd()
	require.NoError(t, err)
	return testDir
}

func wegoBinaryPath(t *testing.T) string {
	path, err := filepath.Abs("../../../bin/wego")
	require.NoError(t, err)
	return path
}
