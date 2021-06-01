/**
* All common util functions and golbal constants will go here.
**/
package acceptance

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	log "github.com/sirupsen/logrus"
)

const EVENTUALLY_DEFAULT_TIME_OUT time.Duration = 60 * time.Second
const INSTALL_RESET_TIMEOUT time.Duration = 300 * time.Second
const INSTALL_PODS_READY_TIMEOUT time.Duration = 180 * time.Second
const WEGO_DEFAULT_NAMESPACE = "wego-system"

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
	if itemType != "" {
		return runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	}
	return runCommandPassThrough([]string{}, "kubectl", "get", "all", "--all-namespaces", "-o", "wide")
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func ResetOrCreateCluster(namespace string) (string, error) {

	supportedProviders := []string{"kind", "eks", "gke"}
	supportedK8SVersions := []string{"1.19.1", "1.20.2", "1.21.1"}
	clusterName := ""

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
		return clusterName, errors.New("Unsupported provider")
	}

	if !contains(supportedK8SVersions, k8sVersion) {
		log.Errorf("Kubernetes version %s is not supported for testing", k8sVersion)
		return clusterName, errors.New("Unsupported kubernetes version")
	}

	//For eks and gke, we will try to reset the namespace only as re-creating cluster is expensive
	if namespace != "" && (provider == "eks" || provider == "gke") {
		err := runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("%s install --namespace %s| kubectl --ignore-not-found=true delete -f -", WEGO_BIN_PATH, namespace))
		if err != nil {
			log.Infof("Failed to reset the namespace %s", namespace)
			return clusterName, err
		}
	}

	if provider == "kind" {
		clusterName = provider + "-" + RandString(6)
		log.Infof("Creating a kind cluster %s", clusterName)
		err := runCommandPassThrough([]string{}, "./scripts/kind-cluster.sh", clusterName, "kindest/node:v"+k8sVersion)
		if err != nil {
			log.Infof("Failed to create kind cluster")
			log.Fatal(err)
			return clusterName, err
		}
	}

	log.Info("Wait for the cluster to be ready")
	err := runCommandPassThrough([]string{}, "kubectl", "wait", "--for=condition=Ready", "--timeout=300s", "-n", "kube-system", "--all", "pods")
	if err != nil {
		log.Infof("Cluster system pods are not ready after waiting for 5 minutes, This can cause tests failures.")
		return clusterName, err
	}

	return clusterName, nil
}

func waitForResource(resourceType string, resourceName string, namespace string, timeout time.Duration) error {
	pollInterval := 5
	if timeout < 5*time.Second {
		timeout = 5 * time.Second
	}

	timeoutInSeconds := int(timeout.Seconds())
	for i := pollInterval; i < timeoutInSeconds; i += pollInterval {
		log.Infof("Waiting for %s in namespace: %s... : %d second(s) passed of %d seconds timeout", resourceType+"/"+resourceName, namespace, i, timeoutInSeconds)
		err := runCommandPassThroughWithoutOutput([]string{}, "sh", "-c", fmt.Sprintf("kubectl get %s %s -n %s", resourceType, resourceName, namespace))
		if err == nil {
			log.Infof("%s are available in cluster", resourceType+"/"+resourceName)
			command := exec.Command("sh", "-c", fmt.Sprintf("kubectl get %s %s -n %s", resourceType, resourceName, namespace))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			noResourcesFoundMessage := fmt.Sprintf("No resources found in %s namespace", namespace)
			if strings.Contains(string(session.Wait().Out.Contents()), noResourcesFoundMessage) {
				log.Infof("Got message => {" + noResourcesFoundMessage + "} Continue looking for resource(s)")
				continue
			}
			return nil
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
	return fmt.Errorf("Error: Failed to find the resource, timeout reached")
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

func runCommandPassThroughWithoutOutput(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	return cmd.Run()
}
