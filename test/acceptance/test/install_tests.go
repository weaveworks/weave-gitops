/**
* All tests related to 'gitops install' will go into this file
 */

package acceptance

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Install Tests", func() {

	var sessionOutput *gexec.Session

	BeforeEach(func() {
		By("Given I have a gitops binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Validate that gitops displays help text for 'install' command", func() {

		By("When I run the command 'gitops install -h'", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH+" install -h", "")
		})

		By("Then I should see gitops help text displayed for 'install' command", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`The install command deploys GitOps in the specified namespace.\nIf a previous version is installed, then an in-place upgrade will be performed.\n*Usage:\n\s*gitops install \[flags]\n*Examples:\n\s*# Install GitOps in the wego-system namespace\n\s*gitops install\n*Flags:\n\s*--dry-run\s*outputs all the manifests that would be installed\n\s*-h, --help\s*help for install\n*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Validate that gitops displays help text for 'uninstall' command", func() {

		By("When I run the command 'gitops uninstall -h'", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH+" uninstall -h", "")
		})

		By("Then I should see gitops help text displayed for 'uninstall' command", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`The uninstall command removes GitOps components from the cluster.\n*Usage:\n\s*gitops uninstall \[flags]\n*Examples:\n\s*# Uninstall GitOps from the wego-system namespace\n\s*gitops uninstall\n*Flags:\n\s*--dry-run\s*outputs all the manifests that would be uninstalled\n\s*-h, --help\s*help for uninstall\n*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that gitops quits if flux-system namespace is present", func() {
		var errOutput string
		namespace := "flux-system"

		defer deleteNamespace(namespace)

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, true, "")
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("When I create a '"+namespace+"' namespace", func() {
			namespaceCreatedMsg := runCommandAndReturnSessionOutput("kubectl create ns "+namespace, "")
			Eventually(namespaceCreatedMsg).Should(gbytes.Say("namespace/" + namespace + " created"))
		})

		By("And I run 'gitops install' command", func() {
			_, errOutput = runCommandAndReturnStringOutput(WEGO_BIN_PATH+" install", "")
		})

		By("Then I should see a quitting message", func() {
			Eventually(errOutput).Should(MatchRegexp(
				`Error: Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n\s*. flux uninstall`))
		})
	})

	It("Verify that gitops can install & uninstall gitops components under a user-specified namespace", func() {

		namespace := "test-namespace"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(namespace, true, "")
			Expect(err).ShouldNot(HaveOccurred())
		})

		installAndVerifyWego(namespace, "")

		By("When I run 'gitops uninstall' command", func() {
			_ = runCommandPassThrough([]string{}, "", "sh", "-c", fmt.Sprintf("%s uninstall --namespace %s", WEGO_BIN_PATH, namespace))
		})

		_ = waitForNamespaceToTerminate(namespace, NAMESPACE_TERMINATE_TIMEOUT, "")

		By("Then I should not see any gitops components", func() {
			_, errOutput := runCommandAndReturnStringOutput("kubectl get ns "+namespace, "")
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + namespace + `" not found`))
		})
	})

	It("Verify that gitops can uninstall flux if gitops was not fully installed", func() {

		namespace := "test-namespace"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(namespace, true, "")
			Expect(err).ShouldNot(HaveOccurred())
		})

		installAndVerifyWego(namespace, "")

		ctx := context.Background()

		kubeClient, _, kubeErr := kube.NewKubeHTTPClient()
		Expect(kubeErr).ShouldNot(HaveOccurred())

		crdErr := kubeClient.Delete(ctx, manifests.AppCRD)
		Expect(crdErr).ShouldNot(HaveOccurred())

		By("When I run 'gitops uninstall' command", func() {
			runErr := runCommandPassThrough([]string{}, "", "sh", "-c", fmt.Sprintf("%s uninstall --namespace %s", WEGO_BIN_PATH, namespace))
			Expect(runErr).ShouldNot(HaveOccurred())
		})

		_ = waitForNamespaceToTerminate(namespace, NAMESPACE_TERMINATE_TIMEOUT, "")

		By("Then I should not see any gitops components", func() {
			_, errOutput := runCommandAndReturnStringOutput("kubectl get ns "+namespace, "")
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + namespace + `" not found`))
		})
	})

	It("Verify that gitops can: install gitops components, uninstall gitops components, and work in dry-run mode", func() {

		var installDryRunOutput string
		var uninstallDryRunOutput string

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, true, "")
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("When I try to install gitops in dry-run mode", func() {
			installDryRunOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH+" install --dry-run", "")
		})

		By("Then I should see install dry-run output in the console", func() {
			Eventually(installDryRunOutput).Should(ContainSubstring("# Flux version: "))
			Eventually(installDryRunOutput).Should(ContainSubstring("# Components: source-controller,kustomize-controller,helm-controller,notification-controller,image-reflector-controller,image-automation-controller"))
			Eventually(installDryRunOutput).Should(ContainSubstring("name: " + WEGO_DEFAULT_NAMESPACE))
		})

		By("And gitops components should be absent from the cluster", func() {
			_, err := runCommandAndReturnStringOutput("kubectl get ns "+WEGO_DEFAULT_NAMESPACE, "")
			Eventually(err).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + WEGO_DEFAULT_NAMESPACE + `" not found`))
		})

		installAndVerifyWego(WEGO_DEFAULT_NAMESPACE, "")

		By("When I try to uninstall gitops in dry-run mode", func() {
			uninstallDryRunOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH+" uninstall --dry-run", "")
		})

		By("Then I should see uninstall dry-run output in the console", func() {
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("► deleting components in " + WEGO_DEFAULT_NAMESPACE + " namespace"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Deployment/wego-system/helm-controller deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Deployment/wego-system/image-automation-controller deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Deployment/wego-system/image-reflector-controller deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Deployment/wego-system/kustomize-controller deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Deployment/wego-system/notification-controller deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Deployment/wego-system/source-controller deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ Namespace/wego-system deleted (dry run)"))
			Eventually(uninstallDryRunOutput).Should(ContainSubstring("✔ uninstall finished"))
		})

		By("And gitops components should be present in the cluster", func() {
			VerifyControllersInCluster(WEGO_DEFAULT_NAMESPACE, "")
		})

		By("When I run 'gitops uninstall' command", func() {
			_ = runCommandPassThrough([]string{}, "", "sh", "-c", fmt.Sprintf("%s uninstall --namespace %s", WEGO_BIN_PATH, WEGO_DEFAULT_NAMESPACE))
		})

		_ = waitForNamespaceToTerminate(WEGO_DEFAULT_NAMESPACE, NAMESPACE_TERMINATE_TIMEOUT, "")

		By("Then I should not see any gitops components", func() {
			_, errOutput := runCommandAndReturnStringOutput("kubectl get ns "+WEGO_DEFAULT_NAMESPACE, "")
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + WEGO_DEFAULT_NAMESPACE + `" not found`))
		})
	})
})
