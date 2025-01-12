package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseConfig", func() {
	It("parses config from data", func() {
		data := []byte(`{
	"analytics": true,
	"userId": "Hph7bg5SiK"
}`)
		config := &GitopsCLIConfig{}

		err := parseConfig(data, config)

		Expect(err).NotTo(HaveOccurred())
		Expect(config.Analytics).To(BeTrue())
		Expect(config.UserID).To(Equal("Hph7bg5SiK"))
	})
})

var _ = Describe("GenerateUserID", func() {
	It("generates user ID", func() {
		userID := GenerateUserID(10, 1024)

		Expect(userID).To(Equal("ULhi8C5Ti1"))
	})
})
