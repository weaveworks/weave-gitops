package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate random string", func() {

	It("does not return the same value on each call", func() {
		value1, _ := GenerateRandomString(10, 20)
		value2, _ := GenerateRandomString(10, 20)
		value3, _ := GenerateRandomString(10, 20)

		Expect(value1).NotTo(Equal(value2))
		Expect(value1).NotTo(Equal(value3))
		Expect(value2).NotTo(Equal(value3))
	})

	It("panics when min is greater than max", func() {
		Expect(func() { _, _ = GenerateRandomString(20, 10) }).Should(Panic())
	})

	It("panics when both numbers are equal", func() {
		Expect(func() { _, _ = GenerateRandomString(5, 5) }).Should(Panic())
	})

	It("returns string greater than or equal to min value", func() {
		value, err := GenerateRandomString(11, 12)
		Expect(err).To(BeNil())
		Expect(len(value)).To(BeNumerically(">=", 11))
	})

	It("returns string less than max value", func() {
		value, err := GenerateRandomString(11, 12)
		Expect(err).To(BeNil())
		Expect(len(value)).To(BeNumerically("<", 12))
	})

})
