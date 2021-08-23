/**
* All tests related to 'wego install' will go into this file
 */

package acceptance

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Install Tests", func() {

	var sessionOutput *gexec.Session

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Validate that wego displays help text for 'install' command", func() {

		By("When I run the command 'wego gitops install -h'", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " gitops install -h")
		})

		By("Then I should see wego help text displayed for 'install' command", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`The install command deploys Wego in the specified namespace.\nIf a previous version is installed, then an in-place upgrade will be performed.\n*Usage:\n\s*wego gitops install \[flags]\n*Examples:\n\s*# Install wego in the wego-system namespace\n\s*wego gitops install\n*Flags:\n\s*-h, --help\s*help for install\n*Global Flags:\n\s*--dry-run\s*outputs all the manifests that would be installed\n\s*-n, --namespace string\s*the namespace scope for this operation \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Validate that wego displays help text for 'uninstall' command", func() {

		By("When I run the command 'wego gitops uninstall -h'", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " gitops uninstall -h")
		})

		By("Then I should see wego help text displayed for 'uninstall' command", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`The uninstall command removes Wego components from the cluster.\n*Usage:\n\s*wego gitops uninstall \[flags]\n*Examples:\n\s*# Uninstall wego in the wego-system namespace\n\s*wego uninstall\n*Flags:\n\s*-h, --help\s*help for uninstall\n*Global Flags:\n\s*--dry-run\s*outputs all the manifests that would be installed\n\s*-n, --namespace string \s*the namespace scope for this operation \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that wego quits if flux-system namespace is present", func() {
		var errOutput string
		namespace := "flux-system"

		defer deleteNamespace(namespace)

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, true)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("When I create a '"+namespace+"' namespace", func() {
			namespaceCreatedMsg := runCommandAndReturnSessionOutput("kubectl create ns " + namespace)
			Eventually(namespaceCreatedMsg).Should(gbytes.Say("namespace/" + namespace + " created"))
		})

		By("And I run 'wego gitops install' command", func() {
			_, errOutput = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " gitops install")
		})

		By("Then I should see a quitting message", func() {
			Eventually(errOutput).Should(MatchRegexp(
				`Error: Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n\s*. flux uninstall`))
		})
	})

	It("Verify that wego can install & uninstall wego components under a user-specified namespace", func() {

		namespace := "test-namespace"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(namespace, true)
			Expect(err).ShouldNot(HaveOccurred())
		})

		installAndVerifyWego(namespace)

		By("When I run 'wego gitops uninstall' command", func() {
			_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("%s gitops uninstall --namespace %s", WEGO_BIN_PATH, namespace))
		})

		_ = waitForNamespaceToTerminate(namespace, NAMESPACE_TERMINATE_TIMEOUT)

		By("Then I should not see any wego components", func() {
			_, errOutput := runCommandAndReturnStringOutput("kubectl get ns " + namespace)
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + namespace + `" not found`))
		})
	})

	It("Verify that wego can: install wego components, uninstall wego components, and work in dry-run mode", func() {

		var errorOutput string
		var installDryRunOutput string
		var uninstallDryRunOutput string

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, true)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("When I try to uninstall wego in dry-run mode without installing components first", func() {
			_, errorOutput = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " gitops uninstall --dry-run")
		})

		By("Then I should see an error output", func() {
			Eventually(errorOutput).Should(ContainSubstring("Error: Wego is not installed... exiting"))
		})

		By("When I try to install wego in dry-run mode", func() {
			installDryRunOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " gitops install --dry-run")
		})

		By("Then I should see install dry-run output in the console", func() {
			Eventually(installDryRunOutput).Should(ContainSubstring("# Flux version: "))
			Eventually(installDryRunOutput).Should(ContainSubstring("# Components: source-controller,kustomize-controller,helm-controller,notification-controller,image-reflector-controller,image-automation-controller"))
			Eventually(installDryRunOutput).Should(ContainSubstring("name: " + WEGO_DEFAULT_NAMESPACE))
		})

		By("And wego components should be absent from the cluster", func() {
			_, err := runCommandAndReturnStringOutput("kubectl get ns " + WEGO_DEFAULT_NAMESPACE)
			Eventually(err).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + WEGO_DEFAULT_NAMESPACE + `" not found`))
		})

		installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)

		By("When I try to uninstall wego in dry-run mode", func() {
			uninstallDryRunOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " gitops uninstall --dry-run")
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

		By("And wego components should be present in the cluster", func() {
			VerifyControllersInCluster(WEGO_DEFAULT_NAMESPACE)
		})

		By("When I run 'wego gitops uninstall' command", func() {
			_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("%s gitops uninstall --namespace %s", WEGO_BIN_PATH, WEGO_DEFAULT_NAMESPACE))
		})

		_ = waitForNamespaceToTerminate(WEGO_DEFAULT_NAMESPACE, NAMESPACE_TERMINATE_TIMEOUT)

		By("Then I should not see any wego components", func() {
			_, errOutput := runCommandAndReturnStringOutput("kubectl get ns " + WEGO_DEFAULT_NAMESPACE)
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + WEGO_DEFAULT_NAMESPACE + `" not found`))
		})
	})
})
