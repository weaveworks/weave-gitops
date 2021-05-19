/**
* All common util functions and golbal constants will go here.
**/
package acceptance

import (
	"errors"
	"math/rand"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const EVENTUALLY_DEFAULT_TIME_OUT time.Duration = 60 * time.Second
const INSTALL_RESET_TIMEOUT time.Duration = 300 * time.Second

var WEGO_BIN_PATH string

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return StringWithCharset(length, charset)
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// showItems displays the current set of a specified object type in tabular format
func ShowItems(itemType string) error {
	return runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func ResetOrCreateCluster() (string, error) {
	supportedProviders := []string{"kind", "eks", "gke"}
	supportedK8SVersions := []string{"1.19.1", "1.20.2", "1.21.1"}

	clusterName := RandString(6)
	provider, found := os.LookupEnv("CLUSTER_PROVIDER")
	if !found {
		provider = "kind"
	}

	k8sVersion, found := os.LookupEnv("K8S_VERSION")
	if !found {
		k8sVersion = "1.20.2"
	}

	if !contains(supportedProviders, provider) {
		log.Errorf("Cluster provider %s is not supported for testing", provider)
		return "", errors.New("Unsupported provider")
	}

	if !contains(supportedK8SVersions, k8sVersion) {
		log.Errorf("Kubernetes version %s is not supported for testing", k8sVersion)
		return "", errors.New("Unsupported kubernetes version")
	}

	clusterName = provider + "-" + clusterName
	if provider == "kind" {
		log.Infof("Creating a kind cluster %s", clusterName)
		err := runCommandPassThrough([]string{}, "./scripts/create-kind-cluster.sh", clusterName, "kindest/node:v"+k8sVersion)
		if err != nil {
			log.Fatal(err)
		}
	}

	return clusterName, nil
}

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
