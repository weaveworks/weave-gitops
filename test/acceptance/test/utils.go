/**
* All common util functions and golbal constants will go here.
**/
package acceptance

import (
	"bytes"
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

const (
	THIRTY_SECOND_TIMEOUT       time.Duration = 30 * time.Second
	EVENTUALLY_DEFAULT_TIMEOUT  time.Duration = 60 * time.Second
	INSTALL_RESET_TIMEOUT       time.Duration = 300 * time.Second
	NAMESPACE_TERMINATE_TIMEOUT time.Duration = 600 * time.Second
	INSTALL_SUCCESSFUL_TIMEOUT  time.Duration = 3 * time.Minute
	INSTALL_PODS_READY_TIMEOUT  time.Duration = 3 * time.Minute
	WEGO_UI_URL                               = "http://localhost:9001"
	SELENIUM_SERVICE_URL                      = "http://localhost:4444/wd/hub"
	SCREENSHOTS_DIR             string        = "screenshots/"
	DEFAULT_BRANCH_NAME                       = "main"
	WEGO_DASHBOARD_TITLE        string        = "Weave GitOps"
	APP_PAGE_HEADER             string        = "Applications"
	charset                                   = "abcdefghijklmnopqrstuvwxyz0123456789"
	DEFAULT_K8S_VERSION         string        = "1.21.1"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func getClusterContext() string {
	out, err := exec.Command("kubectl", "config", "current-context").Output()
	Expect(err).ShouldNot(HaveOccurred())

	return string(bytes.TrimSuffix(out, []byte("\n")))
}

// showItems displays the current set of a specified object type in tabular format
func ShowItems(itemType string) error {
	if itemType != "" {
		return runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	}

	return runCommandPassThrough([]string{}, "kubectl", "get", "all", "--all-namespaces", "-o", "wide")
}

func ShowWegoControllerLogs(ns string) {
	controllers := []string{"helm", "kustomize", "notification", "source"}

	for _, c := range controllers {
		label := c + "-controller"
		log.Infof("Logs for controller: %s", label)
		_ = runCommandPassThrough([]string{}, "kubectl", "logs", "-l", "app="+label, "-n", ns)
	}
}

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

func ResetOrCreateCluster(namespace string, deleteWegoRuntime bool) (string, string, error) {
	return ResetOrCreateClusterWithName(namespace, deleteWegoRuntime, "", false)
}

func getK8sVersion() string {
	k8sVersion, found := os.LookupEnv("K8S_VERSION")
	if found {
		return k8sVersion
	}

	return DEFAULT_K8S_VERSION
}

func ResetOrCreateClusterWithName(namespace string, deleteWegoRuntime bool, clusterName string, keepExistingClusters bool) (string, string, error) {
	supportedProviders := []string{"kind", "kubectl"}
	supportedK8SVersions := []string{"1.19.1", "1.20.2", "1.21.1"}

	provider, found := os.LookupEnv("CLUSTER_PROVIDER")
	if keepExistingClusters || !found {
		provider = "kind"
	}

	k8sVersion := getK8sVersion()

	if !contains(supportedProviders, provider) {
		log.Errorf("Cluster provider %s is not supported for testing", provider)
		return clusterName, "", errors.New("Unsupported provider")
	}

	if !contains(supportedK8SVersions, k8sVersion) {
		log.Errorf("Kubernetes version %s is not supported for testing", k8sVersion)
		return clusterName, "", errors.New("Unsupported kubernetes version")
	}

	//For kubectl, point to a valid cluster, we will try to reset the namespace only
	if namespace != "" && provider == "kubectl" {
		err := runCommandPassThrough([]string{}, "./scripts/reset-wego.sh", namespace)
		if err != nil {
			log.Infof("Failed to reset the wego runtime in namespace %s", namespace)
		}

		name, _ := runCommandAndReturnStringOutput("kubectl config current-context")
		clusterName = strings.TrimSuffix(name, "\n")
	}

	if provider == "kind" {
		var kindCluster string
		if clusterName == "" {
			kindCluster = RandString(6)
		}

		clusterName = provider + "-" + kindCluster

		log.Infof("Creating a kind cluster %s", kindCluster)

		var err error
		if keepExistingClusters {
			err = runCommandPassThrough([]string{}, "./scripts/kind-multi-cluster.sh", kindCluster, "kindest/node:v"+k8sVersion)
		} else {
			err = runCommandPassThrough([]string{}, "./scripts/kind-cluster.sh", kindCluster, "kindest/node:v"+k8sVersion)
		}

		if err != nil {
			log.Infof("Failed to create kind cluster")
			log.Fatal(err)

			return clusterName, "", err
		}
	}

	log.Info("Wait for the cluster to be ready")

	err := runCommandPassThrough([]string{}, "kubectl", "wait", "--for=condition=Ready", "--timeout=300s", "-n", "kube-system", "--all", "pods")

	if err != nil {
		log.Infof("Cluster system pods are not ready after waiting for 5 minutes, This can cause tests failures.")
		return clusterName, "", err
	}

	return clusterName, getClusterContext(), nil
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
			log.Infof("%s is available in cluster", resourceType+"/"+resourceName)
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

	return fmt.Errorf("Error: Failed to find the resource %s of type %s, timeout reached", resourceName, resourceType)
}

func VerifyControllersInCluster(namespace string) {
	Expect(waitForResource("deploy", "helm-controller", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("deploy", "kustomize-controller", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("deploy", "notification-controller", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("deploy", "source-controller", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("deploy", "image-automation-controller", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("deploy", "image-reflector-controller", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("deploy", "wego-app", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("pods", "", namespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())

	By("And I wait for the gitops controllers to be ready", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=%s -n %s --all pod --selector='app!=wego-app'", "120s", namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	})
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

func runCommandAndReturnStringOutput(commandToRun string) (stdOut string, stdErr string) {
	command := exec.Command("sh", "-c", commandToRun)
	session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Eventually(session).Should(gexec.Exit())

	outContent := session.Wait().Out.Contents()
	errContent := session.Wait().Err.Contents()

	return string(outContent), string(errContent)
}

func runCommandAndReturnSessionOutput(commandToRun string) *gexec.Session {
	command := exec.Command("sh", "-c", commandToRun)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())

	return session
}
