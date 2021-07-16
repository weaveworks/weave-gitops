package acceptance

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

//Disabling WEGO Help Tests Suite Until Further Notice...
var _ = XDescribe("WEGO Help Tests", func() {

	var sessionOutput *gexec.Session
	var stringOutput string
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
			stringOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " --help")
		})

		By("Then I should see help message printed for wego", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Weave GitOps\n*Usage:\n\s*wego \[command]\n*Available Commands:`))
			Eventually(stringOutput).Should(MatchRegexp(`app`))
			Eventually(stringOutput).Should(MatchRegexp(`gitops\s*Manages your wego installation`))
			Eventually(stringOutput).Should(MatchRegexp(`flux\s*Use flux commands`))
			Eventually(stringOutput).Should(MatchRegexp(`version\s*Display wego version`))
			Eventually(stringOutput).Should(MatchRegexp(`help\s*Help about any command`))
			Eventually(stringOutput).Should(MatchRegexp(`Flags:\n\s*-h, --help\s*help for wego\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output\n*Use "wego \[command] --help" for more information about a command.`))
		})
	})

	It("Verify that wego app help flag prints the help text", func() {

		By("When I run the command 'wego app -h' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app -h")
		})

		By("Then I should see help message printed for wego app", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Usage:\n\s*wego app \[command]\n*Available Commands:`))
			Eventually(stringOutput).Should(MatchRegexp(`add\s*Add a workload repository to a wego cluster`))
			Eventually(stringOutput).Should(MatchRegexp(`list\s*List applications`))
			Eventually(stringOutput).Should(MatchRegexp(`status\s*Get status of an app`))
			Eventually(stringOutput).Should(MatchRegexp(`Flags:\n\s*-h, --help\s*help for app\n*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output\n*Use "wego app \[command] --help" for more information about a command.`))
		})
	})

	It("Verify that wego app add help flag prints the help text", func() {

		By("When I run the command 'wego app add -h' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app add -h")
		})

		By("Then I should see help message printed for wego app add", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Associates an additional application in a git repository with a wego cluster so that its contents may be managed via GitOps\n*Usage:`))
			Eventually(stringOutput).Should(MatchRegexp(`wego app add \[--name <name>] \[--url <url>] \[--branch <branch>] \[--path <path within repository>] \[--private-key <keyfile>] <repository directory> \[flags]`))
			Eventually(stringOutput).Should(MatchRegexp(`Examples:\nwego app add .\n*Flags:`))
			Eventually(stringOutput).Should(MatchRegexp(`--app-config-url string\s*URL of external repository \(if any\) which will hold automation manifests; NONE to store only in the cluster`))
			Eventually(stringOutput).Should(MatchRegexp(`--auto-merge\s*If set, 'wego add' will merge automatically into the set`))
			Eventually(stringOutput).Should(MatchRegexp(`--branch\n\s*--branch string\s*Branch to watch within git repository \(default "main"\)`))
			Eventually(stringOutput).Should(MatchRegexp(`--chart string\s*Specify chart for helm source`))
			Eventually(stringOutput).Should(MatchRegexp(`--deployment-type string\s*deployment type \[kustomize, helm] \(default "kustomize"\)`))
			Eventually(stringOutput).Should(MatchRegexp(`--dry-run\s*If set, 'wego add' will not make any changes to the system; it will just display the actions that would have been taken`))
			Eventually(stringOutput).Should(MatchRegexp(`-h, --help\s*help for add`))
			Eventually(stringOutput).Should(MatchRegexp(`--name string\s*Name of remote git repository`))
			Eventually(stringOutput).Should(MatchRegexp(`--path string\s*Path of files within git repository \(default "\.\/"\)`))
			Eventually(stringOutput).Should(MatchRegexp(`--private-key string\s*Private key to access git repository over ssh`))
			Eventually(stringOutput).Should(MatchRegexp(`--url string\s*URL of remote repository`))
			Eventually(stringOutput).Should(MatchRegexp(`Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that wego app status help flag prints the help text", func() {

		By("When I run the command 'wego app status -h' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app status -h")
		})

		By("Then I should see help message printed for wego app status", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`Get status of an app\n*Usage:\n\s*wego app status <app-name> \[flags]\n*Examples:\nwego app status podinfo\n*Flags:\n\s*-h, --help\s*help for status\n*\s*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that wego app list help flag prints the help text", func() {

		By("When I run the command 'wego app list -h' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app list -h")
		})

		By("Then I should see help message printed for wego app list", func() {
			Eventually(stringOutput).Should(MatchRegexp(`List applications\n*Usage:\n\s*wego app list \[flags]`))
			Eventually(stringOutput).Should(MatchRegexp(`Examples:\nwego app list`))
			Eventually(stringOutput).Should(MatchRegexp(`Flags:\n\s*-h, --help\s*help for list\n*\s*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})
})
