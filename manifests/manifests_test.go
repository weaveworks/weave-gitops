package manifests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Testing WegoAppDeployment", func() {

	var localDeploymentManifest []byte
	BeforeEach(func() {
		localDeploymentManifest = WegoAppDeployment
	})

	It("should return the right version", func() {
		deploymentYaml, err := GenerateWegoAppDeploymentManifest(localDeploymentManifest)
		Expect(err).NotTo(HaveOccurred())

		var Deployment appsv1.Deployment
		err = yaml.Unmarshal(deploymentYaml, &Deployment)
		Expect(err).NotTo(HaveOccurred())

		Expect(Deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring(version.Version))
	})

	It("should fail trying to parse the template", func() {

		localDeploymentManifest = []byte("{{.wrongField}}")

		_, err := GenerateWegoAppDeploymentManifest(localDeploymentManifest)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring(errInjectingValuesToTemplate.Error()))
		Expect(err.Error()).Should(ContainSubstring("wrongField"))
	})

})
