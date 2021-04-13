package test

// Runs basic WeGO operations against a kind cluster.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/version"
	"sigs.k8s.io/yaml"
)

const nginxDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: my-nginx
  labels:
    name: nginx-deployment
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
	log.Info("Creating temp directory...")
	tmpPath, err := ioutil.TempDir("", "tmp_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tmpPath)
	tmpDir = tmpPath
	log.Info("Ensuring flux version is set...")
	ensureFluxVersion(t)
	log.Info("Checking initial status...")
	checkSimpleStatuses(t)
	log.Info("Bootstrapping flux...")
	bootstrapFlux(t)
	log.Info("Setting up test repository...")
	setUpTestRepo(t) // create repo with simple nginx manifest
	defer deleteRepo(t)
	log.Info("Adding test repository to cluster...")
	err = callWegoForEffect("add .") // add new repo to cluster
	require.NoError(t, err)
	log.Info("Waiting for workload to start...")
	waitForNginxDeployment(t)
}

func ensureFluxVersion(t *testing.T) {
	if version.FluxVersion == "undefined" {
		out, err := fluxops.CallCommand("../../../tools/bin/stoml ../../../tools/dependencies.toml flux.version")
		require.NoError(t, err)
		version.FluxVersion = string(out)
	}
}

func waitForNginxDeployment(t *testing.T) {
	for i := 1; i < 61; i++ {
		_, err := fluxops.CallCommand("kubectl get deployment nginx -n my-nginx")
		if err == nil {
			return
		}
		time.Sleep(5 * time.Second)
	}
	require.FailNow(t, "Failed to deploy nginx workload to the cluster")
}

func bootstrapFlux(t *testing.T) {
	owner := getOwner(t)
	repoName := getRepoName(t)
	if isOrganization(t, owner) {
		_, err := fluxops.CallFlux(fmt.Sprintf("bootstrap github --owner=%s --repository=%s", owner, repoName))
		require.NoError(t, err)
	} else {
		_, err := fluxops.CallFlux(fmt.Sprintf("bootstrap github --owner=%s --repository=%s --branch=main --private=false --personal=true", owner, repoName))
		require.NoError(t, err)
	}
	require.Equal(t, status.GetClusterStatus(), status.FluxInstalled)
}

func getRepoName(t *testing.T) string {
	clusterName, err := status.GetClusterName()
	require.NoError(t, err)
	return clusterName + "-wego"
}

func setUpTestRepo(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	_, err = fluxops.CallCommand("git init .")
	require.NoError(t, err)
	err = ioutil.WriteFile("nginx.yaml", []byte(nginxDeployment), 0666)
	require.NoError(t, err)
	err = fluxops.CallCommandForEffect("git add nginx.yaml && git commit -m'Added workload'")
	require.NoError(t, err)
	err = fluxops.CallCommandForEffect(fmt.Sprintf("hub create %s/%s", getOwner(t), getRepoName(t)))
	require.NoError(t, err)
	err = os.Chdir(dir)
	require.NoError(t, err)
}

func deleteRepo(t *testing.T) {
	org := getOwner(t)
	clusterName, err := status.GetClusterName()
	if err == nil {
		repoName := clusterName + "-wego"
		cmdstr := fmt.Sprintf("hub delete -y %s/%s", org, repoName)
		_ = fluxops.CallCommandForEffect(cmdstr) // there's nothing we can do with the error
	} else {
		log.Info("Failed to delete repository")
	}
}

func isOrganization(t *testing.T, owner string) bool {
	token := os.Getenv("GITHUB_TOKEN")
	response, _, err := fluxops.CallCommandSeparatingOutputStreams(fmt.Sprintf("curl -u %s:%s https://api.github.com/orgs/%s", owner, token, owner))
	fmt.Printf("RESP: %s\n", response)
	require.NoError(t, err)
	var data map[string]interface{}
	err = json.Unmarshal(response, &data)
	require.NoError(t, err)
	return data["message"] != "Not Found"
}

func getOwner(t *testing.T) string {
	// check for github username
	user, okUser := os.LookupEnv("GITHUB_ORG")
	if okUser {
		return user
	}

	return getUserFromHubCredentials(t)
}

func getUserFromHubCredentials(t *testing.T) string {
	homeDir, err := os.UserHomeDir()
	fmt.Printf("HD: %s\n", homeDir)
	require.NoError(t, err)

	// check for existing ~/.config/hub
	config, err := ioutil.ReadFile(filepath.Join(homeDir, ".config", "hub"))
	fmt.Printf("CONF: %s\n", config)
	require.NoError(t, err)

	data := map[string]interface{}{}
	err = yaml.Unmarshal(config, &data)
	require.NoError(t, err)
	fmt.Printf("DATA: %#v\n", data)
	return data["github.com"].([]interface{})[0].(map[string]interface{})["user"].(string)
}

func checkSimpleStatuses(t *testing.T) {
	savedHome := os.Getenv("HOME")

	err := os.Setenv("HOME", "/iewojfoiwejfoiwjfwoijfewj")
	require.NoError(t, err)
	require.Equal(t, status.GetClusterStatus(), status.Unknown)

	err = os.Setenv("HOME", savedHome)
	require.NoError(t, err)
	require.Equal(t, status.GetClusterStatus(), status.Unmodified)
}

func getTestDir(t *testing.T) string {
	testDir, err := os.Getwd()
	require.NoError(t, err)
	return testDir
}

func callWegoForEffect(wegoCmd string) error {
	cmdstr := fmt.Sprintf("%s %s", wegoBinaryPath(), wegoCmd)
	cmd := exec.Command("sh", "-c", fluxops.Escape(cmdstr))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wegoBinaryPath() string {
	return "../../../bin/wego"
}
