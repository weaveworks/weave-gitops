package manifests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
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

		var Deployment struct {
			Spec struct {
				Template struct {
					Spec struct {
						Containers []struct {
							Image string `yaml:"image"`
						} `yaml:"containers"`
					} `yaml:"spec"`
				} `yaml:"template"`
			} `yaml:"spec"`
		}

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
