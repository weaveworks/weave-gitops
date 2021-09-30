package manifests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Testing WegoAppDeployment", func() {
	It("should contain the right version", func() {
		v := version.Version
		deploymentYaml, err := GenerateWegoAppDeploymentManifest(v)
		Expect(err).NotTo(HaveOccurred())

		var Deployment appsv1.Deployment
		err = yaml.Unmarshal(deploymentYaml, &Deployment)
		Expect(err).NotTo(HaveOccurred())

		Expect(Deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring(v))
	})
})
