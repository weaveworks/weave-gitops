package manifests

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

var _ = Describe("Testing WegoAppDeployment", func() {
	It("should contain the right version", func() {
		manifests, err := GenerateWegoAppManifests(WegoAppParams{Version: version.Version, Namespace: "my-namespace"})
		Expect(err).NotTo(HaveOccurred())

		for _, m := range manifests {
			if strings.Contains(string(m), "kind: Deployment") {
				Expect(string(m)).To(ContainSubstring("namespace: my-namespace"))
				Expect(string(m)).To(ContainSubstring(version.Version))
			}
		}
	})
})
