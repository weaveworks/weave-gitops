package check

import (
	"github.com/Masterminds/semver/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check kubernetes version", func() {
	It("should show kuberetes is a valid version", func() {
		version, err := semver.NewVersion("1.21.1")
		Expect(err).ShouldNot(HaveOccurred())

		output, err := checkKubernetesVersion(version)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(output).To(Equal("✔ Kubernetes 1.21.1 >=1.20.6-0"))
	})

	It("should fail with version does not match", func() {
		version, err := semver.NewVersion("1.19.1")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = checkKubernetesVersion(version)
		Expect(err.Error()).Should(Equal("✗ kubernetes version 1.19.1 does not match >=1.20.6-0"))
	})
})

var _ = Describe("parse version", func() {
	It("should parse version", func() {
		expectedVersion, err := semver.NewVersion("1.21.1")
		Expect(err).ShouldNot(HaveOccurred())

		output, err := parseVersion("Server Version: v1.21.1")
		Expect(err).ShouldNot(HaveOccurred())

		Expect(output).To(Equal(expectedVersion))
	})
})
