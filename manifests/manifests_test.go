package manifests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
)

var _ = Describe("Testing WegoAppDeployment", func() {

	var localDeploymentManifest []byte
	BeforeEach(func() {
		localDeploymentManifest = WegoAppDeployment
	})

	It("should return the right version", func() {
		deploymentYaml, err := GenerateWegoAppDeploymentManifest(localDeploymentManifest)
		Expect(err).NotTo(HaveOccurred())

		Expect(deploymentYaml).To(ContainSubstring(version.Version))
	})

	It("should fail trying to parse the template", func() {

		localDeploymentManifest = []byte("{{.wrongField}}")

		_, err := GenerateWegoAppDeploymentManifest(localDeploymentManifest)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring(errInjectingValuesToTemplate.Error()))
		Expect(err.Error()).Should(ContainSubstring("wrongField"))
	})

})
