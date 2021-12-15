package acceptance

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

//Disabling WEGO Help Tests Suite Until Further Notice...
var _ = XDescribe("WEGO Help Tests", func() {

	var sessionOutput *gexec.Session
	var stringOutput string
	var err error

	BeforeEach(func() {

		By("Given I have a gitops binary installed on my local machine", func() {
			Expect(FileExists(gitopsBinaryPath)).To(BeTrue())
		})
	})

	It("Verify that gitops displays error message when provided with the wrong flag", func() {

		By("When I run 'gitops foo'", func() {
			command := exec.Command(gitopsBinaryPath, "foo")
			sessionOutput, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see gitops error message", func() {
			Eventually(sessionOutput.Err).Should(gbytes.Say("Error: unknown command \"foo\" for \"gitops\""))
			Eventually(sessionOutput.Err).Should(gbytes.Say("Run 'gitops --help' for usage."))
		})
	})

	It("Verify that gitops help flag prints the help text", func() {

		By("When I run the command 'gitops --help' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(gitopsBinaryPath + " --help")
		})

		By("Then I should see help message printed for gitops", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Weave GitOps\n*Usage:\n\s*gitops \[command]\n*Available Commands:`))
			Eventually(stringOutput).Should(MatchRegexp(`app`))
			Eventually(stringOutput).Should(MatchRegexp(`gitops\s*Manages your gitops installation`))
			Eventually(stringOutput).Should(MatchRegexp(`flux\s*Use flux commands`))
			Eventually(stringOutput).Should(MatchRegexp(`version\s*Display gitops version`))
			Eventually(stringOutput).Should(MatchRegexp(`help\s*Help about any command`))
			Eventually(stringOutput).Should(MatchRegexp(fmt.Sprintf(`Flags:\n\s*-h, --help\s*help for gitops\n\s*--namespace string\s*gitops runtime namespace \(default "%s"\)\n\s*-v, --verbose\s*Enable verbose output\n*Use "gitops \[command] --help" for more information about a command.`, wego.DefaultNamespace)))
		})
	})

	It("Verify that gitops app help flag prints the help text", func() {

		By("When I run the command 'gitops app -h' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(gitopsBinaryPath + " app -h")
		})

		By("Then I should see help message printed for gitops app", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Usage:\n\s*gitops app \[command]\n*Available Commands:`))
			Eventually(stringOutput).Should(MatchRegexp(`add\s*Add a workload repository to a gitops cluster`))
			Eventually(stringOutput).Should(MatchRegexp(`list\s*List applications`))
			Eventually(stringOutput).Should(MatchRegexp(`status\s*Get status of an app`))
			Eventually(stringOutput).Should(MatchRegexp(fmt.Sprintf(`Flags:\n\s*-h, --help\s*help for app\n*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "%s"\)\n\s*-v, --verbose\s*Enable verbose output\n*Use "gitops app \[command] --help" for more information about a command.`, wego.DefaultNamespace)))
		})
	})

	It("Verify that gitops add app help flag prints the help text", func() {

		By("When I run the command 'gitops add app -h' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(gitopsBinaryPath + " add app -h")
		})

		By("Then I should see help message printed for gitops add app", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Associates an additional application in a git repository with a gitops cluster so that its contents may be managed via GitOps\n*Usage:`))
			Eventually(stringOutput).Should(MatchRegexp(`gitops add app \[--name <name>] \[--url <url>] \[--branch <branch>] \[--path <path within repository>] \ <repository directory> \[flags]`))
			Eventually(stringOutput).Should(MatchRegexp(`Examples:\ngitops add app .\n*Flags:`))
			Eventually(stringOutput).Should(MatchRegexp(`--config-repo string\s*URL of external repository \(if any\) which will hold automation manifests`))
			Eventually(stringOutput).Should(MatchRegexp(`--auto-merge\s*If set, 'gitops add app' will merge automatically into the set`))
			Eventually(stringOutput).Should(MatchRegexp(`--branch\n\s*--branch string\s*Branch to watch within git repository \(default "main"\)`))
			Eventually(stringOutput).Should(MatchRegexp(`--chart string\s*Specify chart for helm source`))
			Eventually(stringOutput).Should(MatchRegexp(`--deployment-type string\s*deployment type \[kustomize, helm] \(default "kustomize"\)`))
			Eventually(stringOutput).Should(MatchRegexp(`--dry-run\s*If set, 'gitops add app' will not make any changes to the system; it will just display the actions that would have been taken`))
			Eventually(stringOutput).Should(MatchRegexp(`-h, --help\s*help for add`))
			Eventually(stringOutput).Should(MatchRegexp(`--name string\s*Name of remote git repository`))
			Eventually(stringOutput).Should(MatchRegexp(`--path string\s*Path of files within git repository \(default "\.\/"\)`))
			Eventually(stringOutput).Should(MatchRegexp(`--url string\s*URL of remote repository`))
			Eventually(stringOutput).Should(MatchRegexp(fmt.Sprintf(`Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "%s"\)\n\s*-v, --verbose\s*Enable verbose output`, wego.DefaultNamespace)))
		})
	})

	It("Verify that gitops get app help flag prints the help text", func() {

		By("When I run the command 'gitops get app -h' ", func() {
			sessionOutput = runCommandAndReturnSessionOutput(gitopsBinaryPath + " get app -h")
		})

		By("Then I should see help message printed for gitops get app", func() {
			Eventually(string(sessionOutput.Wait().Out.Contents())).Should(MatchRegexp(
				`Show information about one or all of the applications under gitops control\n*Usage:\n\s*gitops get app [flags]`))
			Eventually(stringOutput).Should(MatchRegexp(`Examples:\ngitops get app <app-name>`))
			Eventually(stringOutput).Should(MatchRegexp(fmt.Sprintf(`Flags:\n\s*-h, --help\s*help for app\n*\s*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "%s"\)\n\s*-v, --verbose\s*Enable verbose output`, wego.DefaultNamespace)))
		})
	})

	It("Verify that gitops get apps help flag prints the help text", func() {

		By("When I run the command 'gitops get apps -h' ", func() {
			stringOutput, _ = runCommandAndReturnStringOutput(gitopsBinaryPath + " get apps -h")
		})

		By("Then I should see help message printed for gitops get apps", func() {
			Eventually(stringOutput).Should(MatchRegexp(`Show information about one or all of the applications under gitops control\n*Usage:\n\s*gitops get app [flags]`))
			Eventually(stringOutput).Should(MatchRegexp(`Examples:\ngitops get apps`))
			Eventually(stringOutput).Should(MatchRegexp(fmt.Sprintf(`Flags:\n\s*-h, --help\s*help for app\n*\s*Global Flags:\n\s*--namespace string\s*gitops runtime namespace \(default "%s"\)\n\s*-v, --verbose\s*Enable verbose output`, wego.DefaultNamespace)))
		})
	})
})
