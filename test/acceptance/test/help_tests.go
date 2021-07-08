package acceptance

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("WEGO Help Tests", func() {

	var sessionOutput *gexec.Session
	var err error

	BeforeEach(func() {

		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify that wego displays error message when provided with the wrong flag", func() {

		By("When I run 'wego foo'", func() {
			command := exec.Command(WEGO_BIN_PATH, "foo")
			sessionOutput, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see wego error message", func() {
			Eventually(sessionOutput.Err).Should(gbytes.Say("Error: unknown command \"foo\" for \"wego\""))
			Eventually(sessionOutput.Err).Should(gbytes.Say("Run 'wego --help' for usage."))
		})
	})

	It("Verify that wego help flag prints the help text", func() {

		By("When I run the command 'wego --help' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " --help")
		})

		By("Then I should see help message printed for wego", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`Weave GitOps\n*Usage:\n\s*wego \[command]\n*Available Commands:\n\s*app\s*\n\s*flux\s*Use flux commands\n\s*gitops\s*Manages your wego installation\n\s*help\s*Help about any command\n\s*version\s*Display wego version\n*Flags:\n\s*-h, --help\s*help for wego\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output\n*Use "wego \[command] --help" for more information about a command.`))
		})
	})

	It("Verify that wego app help flag prints the help text", func() {

		By("When I run the command 'wego app -h' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app -h")
		})

		By("Then I should see help message printed for wego app", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`Usage:\n\s*wego app \[command]\n*Available Commands:\n\s*add\s*Add a workload repository to a wego cluster\n\s*list\s*List applications\n\s*status\s*Get status of an app\n*Flags:\n\s*-h, --help\s*help for app\n*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output\n*Use "wego app \[command] --help" for more information about a command.`))
		})
	})

	It("Verify that wego app add help flag prints the help text", func() {

		By("When I run the command 'wego app add -h' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app add -h")
		})

		By("Then I should see help message printed for wego app add", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`Associates an additional application in a git repository with a wego cluster so that its contents may be managed via GitOps\n*Usage:\n\s*wego app add \[--name <name>] \[--url <url>] \[--branch <branch>] \[--path <path within repository>] \[--private-key <keyfile>] <repository directory> \[flags]\n*Examples:\nwego app add .\n*Flags:\n\s*--app-config-url string\s*URL of external repository \(if any\) which will hold automation manifests; NONE to store only in the cluster\n\s*--auto-merge\s*If set, 'wego add' will merge automatically into the set --branch\n\s*--branch string\s*Branch to watch within git repository \(default "main"\)\n\s*--chart string\s*Specify chart for helm source\n\s*--deployment-type string\s*deployment type \[kustomize, helm] \(default "kustomize"\)\n\s*--dry-run\s*If set, 'wego add' will not make any changes to the system; it will just display the actions that would have been taken\n\s*-h, --help\s*help for add\n\s*--name string\s*Name of remote git repository\n\s*--path string\s*Path of files within git repository \(default "\.\/"\)\n\s*--private-key string\s*Private key to access git repository over ssh\n\s*--url string\s*URL of remote repository\n*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that wego app status help flag prints the help text", func() {

		By("When I run the command 'wego app add -h' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app status -h")
		})

		By("Then I should see help message printed for wego app status", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`Get status of an app\n*Usage:\n\s*wego app status <app-name> \[flags]\n*Examples:\nwego app status podinfo\n*Flags:\n\s*-h, --help\s*help for status\n*\s*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that wego app list help flag prints the help text", func() {

		By("When I run the command 'wego app list -h' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app list -h")
		})

		By("Then I should see help message printed for wego app list", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`List applications\n*Usage:\n\s*wego app list \[flags]\n*Examples:\nwego app list\n*Flags:\n\s*-h, --help\s*help for list\n*\s*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})
})
