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
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega/gbytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	log "github.com/sirupsen/logrus"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

const APP_REMOVAL_TIMEOUT time.Duration = 30 * time.Second
const EVENTUALLY_DEFAULT_TIME_OUT time.Duration = 60 * time.Second
const TIMEOUT_TWO_MINUTES time.Duration = 120 * time.Second
const INSTALL_RESET_TIMEOUT time.Duration = 300 * time.Second
const NAMESPACE_TERMINATE_TIMEOUT time.Duration = 600 * time.Second
const INSTALL_PODS_READY_TIMEOUT time.Duration = 3 * time.Minute
const WEGO_DEFAULT_NAMESPACE = wego.DefaultNamespace

var DEFAULT_SSH_KEY_PATH string
var GITHUB_ORG string
var WEGO_BIN_PATH string

type TestInputs struct {
	appRepoName         string
	appManifestFilePath string
	workloadName        string
	workloadNamespace   string
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

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

func getClusterName() string {
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

func generateTestInputs() TestInputs {
	var inputs TestInputs
	uniqueSuffix := RandString(6)
	inputs.appRepoName = "wego-test-app-" + RandString(8)
	inputs.appManifestFilePath = getUniqueWorkload("xxyyzz", uniqueSuffix)
	inputs.workloadName = "nginx-" + uniqueSuffix
	inputs.workloadNamespace = "my-nginx-" + uniqueSuffix
	return inputs
}

func getUniqueWorkload(placeHolderSuffix string, uniqueSuffix string) string {
	workloadTemplateFilePath := "./data/nginx-template.yaml"
	absWorkloadManifestFilePath := "/tmp/nginx-" + uniqueSuffix + ".yaml"
	_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("sed 's/%s/%s/g' %s > %s", placeHolderSuffix, uniqueSuffix, workloadTemplateFilePath, absWorkloadManifestFilePath))
	return absWorkloadManifestFilePath
}

func setupSSHKey(sshKeyPath string) {
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		command := exec.Command("sh", "-c", fmt.Sprintf(`
	                       echo "%s" >> %s &&
	                       chmod 0600 %s &&
	                       ls -la %s`, os.Getenv("GITHUB_KEY"), sshKeyPath, sshKeyPath, sshKeyPath))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	}
}

func ResetOrCreateCluster(namespace string, deleteWegoRuntime bool) (string, error) {
	return ResetOrCreateClusterWithName(namespace, deleteWegoRuntime, "")
}

func ResetOrCreateClusterWithName(namespace string, deleteWegoRuntime bool, clusterName string) (string, error) {
	supportedProviders := []string{"kind", "kubectl"}
	supportedK8SVersions := []string{"1.19.1", "1.20.2", "1.21.1"}

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

	//For kubectl, point to a valid cluster, we will try to reset the namespace only
	if namespace != "" && provider == "kubectl" {
		err := runCommandPassThrough([]string{}, "./scripts/reset-wego.sh", namespace)
		if err != nil {
			log.Infof("Failed to reset the wego runtime in namespace %s", namespace)
		}

		if deleteWegoRuntime {
			uninstallWegoRuntime(namespace)
		}
	}

	if provider == "kind" {
		if clusterName == "" {
			clusterName = provider + "-" + RandString(6)
		}
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

func initAndCreateEmptyRepo(appRepoName string, IsPrivateRepo bool) string {
	repoAbsolutePath := "/tmp/" + appRepoName
	privateRepo := ""
	if IsPrivateRepo {
		privateRepo = "-p"
	}

	// We need this step in case running a single test case locally
	err := os.RemoveAll(repoAbsolutePath)
	Expect(err).ShouldNot(HaveOccurred())

	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            mkdir %s &&
                            cd %s &&
                            git init &&
                            git checkout -b main`, repoAbsolutePath, repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, 10, 1).Should(gexec.Exit())
	Expect(string(session.Out.Contents())).Should(MatchRegexp(fmt.Sprintf(`Initialized empty Git repository in (/private)?/tmp/%s/.git/`, appRepoName)))

	randStr := RandString(10)
	fmt.Println("RANDOM-STR", randStr)

	fmt.Fprintf(GinkgoWriter, "%s wainting for creation", randStr)
	err = utils.WaitUntil(GinkgoWriter, time.Second*2, time.Second*20, func() error {
		command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
                            hub create %s %s`, repoAbsolutePath, GITHUB_ORG+"/"+appRepoName, privateRepo))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		if err != nil {
			return fmt.Errorf("%s error running command by ginkgo %w", randStr, err)
		}
		Eventually(session, 10, 1).Should(gexec.Exit())
		fmt.Fprintf(GinkgoWriter, "%s session.Out[%s]", randStr, session.Out.Contents())
		fmt.Fprintf(GinkgoWriter, "%s session.Err[%s]", randStr, session.Err.Contents())
		if session.ExitCode() != 0 && !bytes.Contains(session.Err.Contents(), []byte("Repository not found")) {
			return fmt.Errorf("%s expecting exit code 0, got %d, err %s", randStr, session.ExitCode(), session.Err.Contents())
		}
		Expect(session.Out).Should(gbytes.Say("Updating origin"))
		return nil
	})
	Expect(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(GinkgoWriter, "%s wainting for confirmation", randStr)
	Expect(utils.WaitUntil(GinkgoWriter, time.Second, 20*time.Second, func() error {
		cmd := fmt.Sprintf(`hub api repos/%s/%s`, GITHUB_ORG, appRepoName)
		command := exec.Command("sh", "-c", cmd)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Eventually(session, 10, 1).Should(gexec.Exit())
		return err
	})).ShouldNot(HaveOccurred())
	fmt.Fprintf(GinkgoWriter, "%s after confirmation", randStr)

	return repoAbsolutePath
}

func createSubDir(subDirName string, repoAbsolutePath string) string {
	subDirAbsolutePath := fmt.Sprintf(`%s/%s`, repoAbsolutePath, subDirName)
	command := exec.Command("sh", "-c", fmt.Sprintf(`mkdir -p %s`, subDirAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return subDirAbsolutePath
}

func createGitRepoBranch(repoAbsolutePath string, branchName string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && git checkout -b %s && git push --set-upstream origin %s", repoAbsolutePath, branchName, branchName))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return string(session.Wait().Out.Contents())
}

func getRepoVisibility(org string, repo string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("hub api --flat repos/%s/%s|grep -i private|cut -f2", org, repo))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	visibilityStr := strings.TrimSpace(string(session.Wait().Out.Contents()))
	log.Infof("Repo visibility private=%s", visibilityStr)
	return visibilityStr
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

func waitForNamespaceToTerminate(namespace string, timeout time.Duration) error {
	//check if the namespace exist before cleaning up
	pollInterval := 5
	if timeout < 5*time.Second {
		timeout = 5 * time.Second
	}
	timeoutInSeconds := int(timeout.Seconds())

	err := runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl get ns %s", namespace))

	if err != nil {
		log.Infof("Namespace %s doesn't exist, nothing to clean — skipping...", namespace)
		return nil
	}

	for i := pollInterval; i < timeoutInSeconds; i += pollInterval {
		log.Infof("Waiting for namespace: %s to terminate : %d second(s) passed of %d seconds timeout", namespace, i, timeoutInSeconds)

		out, _ := runCommandAndReturnStringOutput(fmt.Sprintf("kubectl get ns %s --ignore-not-found=true | grep -i terminating", namespace))
		out = strings.TrimSpace(out)
		if out == "" {
			return nil
		}
		if i > timeoutInSeconds/2 && i%10 == 0 {
			//Patch the finalizer
			log.Infof("Patch the finalizer to unstuck the terminating namespace %s", namespace)
			_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl patch ns %s -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge", namespace))
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
	return fmt.Errorf("Error: Failed to terminate the namespace %s", namespace)
}

func VerifyControllersInCluster(namespace string) {
	Expect(waitForResource("deploy", "helm-controller", namespace, INSTALL_PODS_READY_TIMEOUT))
	Expect(waitForResource("deploy", "kustomize-controller", namespace, INSTALL_PODS_READY_TIMEOUT))
	Expect(waitForResource("deploy", "notification-controller", namespace, INSTALL_PODS_READY_TIMEOUT))
	Expect(waitForResource("deploy", "source-controller", namespace, INSTALL_PODS_READY_TIMEOUT))
	Expect(waitForResource("deploy", "image-automation-controller", namespace, INSTALL_PODS_READY_TIMEOUT))
	Expect(waitForResource("deploy", "image-reflector-controller", namespace, INSTALL_PODS_READY_TIMEOUT))
	Expect(waitForResource("pods", "", namespace, INSTALL_PODS_READY_TIMEOUT))

	By("And I wait for the wego controllers to be ready", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=%s -n %s --all pod", "120s", namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	})
}

func installAndVerifyWego(wegoNamespace string) {
	By("And I run 'wego install' command with namespace "+wegoNamespace, func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s gitops install --namespace=%s", WEGO_BIN_PATH, wegoNamespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, TIMEOUT_TWO_MINUTES).Should(gexec.Exit())
		VerifyControllersInCluster(wegoNamespace)
	})
}

func uninstallWegoRuntime(namespace string) {
	log.Infof("About to delete WeGO runtime from namespace: %s", namespace)
	err := runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("%s flux uninstall --namespace %s --silent", WEGO_BIN_PATH, namespace))
	if err != nil {
		log.Infof("Failed to uninstall the wego runtime %s", namespace)
	}

	err = runCommandPassThrough([]string{}, "sh", "-c", "kubectl delete crd apps.wego.weave.works")
	if err != nil {
		log.Infof("Failed to delete crd apps.wego.weave.works")
	}
	Expect(waitForNamespaceToTerminate(namespace, NAMESPACE_TERMINATE_TIMEOUT)).To(Succeed())
}

func deleteNamespace(namespace string) {
	log.Infof("Deleting namespace: " + namespace)
	command := exec.Command("kubectl", "delete", "ns", namespace)
	session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Eventually(session).Should(gexec.Exit())
}

func deleteRepo(appRepoName string) {
	log.Infof("Delete application repo: %s", GITHUB_ORG+"/"+appRepoName)
	_ = runCommandPassThrough([]string{}, "hub", "delete", "-y", GITHUB_ORG+"/"+appRepoName)
}

func deleteWorkload(workloadName string, workloadNamespace string) {
	log.Infof("Delete the namespace %s along with workload %s", workloadNamespace, workloadName)
	_ = runCommandPassThrough([]string{}, "kubectl", "delete", "ns", workloadNamespace)
	_ = waitForNamespaceToTerminate(workloadNamespace, INSTALL_RESET_TIMEOUT)
}

func deletePersistingHelmApp(namespace string, workloadName string, timeout time.Duration) {
	//check if application exists before cleaning up
	err := runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl get -n %s pod/%s", namespace, workloadName))
	if err != nil {
		fmt.Println("No workloads exist under the namespace: " + namespace + ", nothing to clean — skipping...")
	} else {
		log.Infof("Found persisting helm workload under the namespace: " + namespace + ", cleaning up...")
		_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl delete -n %s helmreleases.helm.toolkit.fluxcd.io --all", namespace))
		_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl delete -n %s helmcharts.source.toolkit.fluxcd.io --all", namespace))
		_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl delete -n %s helmrepositories.source.toolkit.fluxcd.io --all", namespace))
		_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl delete apps -n %s --all", namespace))
		_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("kubectl wait --for=delete pod/%s -n %s --timeout=%s", workloadName, namespace, timeout))
	}
}

func createAppReplicas(repoAbsolutePath string, appManifestFilePath string, replicasSetValue int, workloadName string) string {
	log.Infof("Editing app-manifest file in git repo to create replicas of workload: %s", workloadName)
	appManifestFile := repoAbsolutePath + "/" + appManifestFilePath
	_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("sed -ie 's/replicas: 1/replicas: %d/g' %s", replicasSetValue, appManifestFile))
	changedValue, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cat %s", appManifestFile))
	return changedValue
}

func waitForReplicaCreation(namespace string, replicasSetValue int, timeout time.Duration) error {

	replica := strconv.Itoa(replicasSetValue)
	pollInterval := time.Second * 5
	timeoutInSeconds := int(timeout.Seconds())

	_ = utils.WaitUntil(os.Stdout, pollInterval, timeout, func() error {
		log.Infof("Waiting for replicas to be created under namespace: %s || timeout: %d second(s)", namespace, timeoutInSeconds)

		out, _ := runCommandAndReturnStringOutput(fmt.Sprintf("kubectl get pods -n %s --field-selector=status.phase=Running --no-headers=true | wc -l", namespace))
		out = strings.TrimSpace(out)
		if out == replica {
			return nil
		}
		return fmt.Errorf(": Replica(s) not created, waiting...")
	})
	return fmt.Errorf("Timeout reached, failed to create replicas")
}

func waitForAppRemoval(appName string, timeout time.Duration) error {

	pollInterval := time.Second * 5

	_ = utils.WaitUntil(os.Stdout, pollInterval, timeout, func() error {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s app list", WEGO_BIN_PATH))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())

		if strings.Contains(string(session.Wait().Out.Contents()), appName) {
			return fmt.Errorf(": Waiting for app: %s to delete", appName)
		}
		return nil
	})
	return fmt.Errorf("Failed to delete app")
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
	return string(session.Wait().Out.Contents()), string(session.Wait().Err.Contents())
}

func runCommandAndReturnSessionOutput(commandToRun string) *gexec.Session {
	command := exec.Command("sh", "-c", commandToRun)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	return session
}

func runWegoAddCommand(repoAbsolutePath string, addCommand string, wegoNamespace string) {
	log.Infof("Add command to run: %s in namespace %s from dir %s", addCommand, wegoNamespace, repoAbsolutePath)
	_, _ = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, wegoNamespace)
}

func runWegoAddCommandWithOutput(repoAbsolutePath string, addCommand string, wegoNamespace string) (string, string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, WEGO_BIN_PATH, addCommand))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return string(session.Wait().Out.Contents()), string(session.Wait().Err.Contents())
}

func verifyWegoAddCommand(appName string, wegoNamespace string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(" kubectl wait --for=condition=Ready --timeout=60s -n %s GitRepositories --all", wegoNamespace))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	Expect(waitForResource("GitRepositories", appName, wegoNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
}

func verifyWegoHelmAddCommand(appName string, wegoNamespace string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=60s -n %s HelmRepositories --all", wegoNamespace))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	Expect(waitForResource("HelmRepositories", appName, wegoNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
}

func verifyWegoAddCommandWithDryRun(appRepoName string, wegoNamespace string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=30s -n %s GitRepositories --all", wegoNamespace))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	Expect(waitForResource("GitRepositories", appRepoName, wegoNamespace, 30*time.Second)).ToNot(Succeed())
}

func verifyWorkloadIsDeployed(workloadName string, workloadNamespace string) {
	Expect(waitForResource("deploy", workloadName, workloadNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	Expect(waitForResource("pods", "", workloadNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=60s -n %s --all pods", workloadNamespace))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
}

func verifyHelmPodWorkloadIsDeployed(workloadName string, workloadNamespace string) {
	Expect(waitForResource("pods", workloadName, workloadNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	c := fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=360s -n %s --all pods --selector='app!=wego-app'", workloadNamespace)
	command := exec.Command("sh", "-c", c)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
}

func gitAddCommitPush(repoAbsolutePath string, appManifestFilePath string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            (cd %s && git pull origin main || true) &&
                            cp -r %s %s &&
                            cd %s &&
                            git add . &&
                            git commit -m 'add workload manifest' &&
                            git push -u origin main`,
		repoAbsolutePath, appManifestFilePath, repoAbsolutePath, repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, 10, 1).Should(gexec.Exit())
}

func gitUpdateCommitPush(repoAbsolutePath string) {
	log.Infof("Pushing changes made to file(s) in repo: %s", repoAbsolutePath)
	_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("cd %s && git add -u && git commit -m 'edit repo file' && git pull --rebase && git push -f", repoAbsolutePath))
}

func pullBranch(repoAbsolutePath string, branch string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
                            cd %s &&
                            git pull origin %s`, repoAbsolutePath, branch))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func pullGitRepo(repoAbsolutePath string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && git pull", repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func verifyPRCreated(repoAbsolutePath, appName string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && hub pr list", repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	output := string(session.Wait().Out.Contents())
	Expect(output).To(ContainSubstring(fmt.Sprintf("wego add %s", appName)))
}

func mergePR(repoAbsolutePath, prLink string) {
	command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && hub merge %s", repoAbsolutePath, prLink))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	command = exec.Command("sh", "-c", fmt.Sprintf("cd %s && git push", repoAbsolutePath))
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}
