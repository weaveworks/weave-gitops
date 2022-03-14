package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Code verifier", func() {

	It("raw value is greater than or equal to 50", func() {
		cv, err := NewCodeVerifier(50, 100)
		Expect(err).To(BeNil())
		Expect(len(cv.RawValue())).To(BeNumerically(">=", 50))
	})

	It("raw value is less than 100", func() {
		cv, err := NewCodeVerifier(50, 100)
		Expect(err).To(BeNil())
		Expect(len(cv.RawValue())).To(BeNumerically("<", 100))
	})

	It("raw value is internal value", func() {
		cv, err := NewCodeVerifier(50, 100)

		Expect(err).To(BeNil())
		Expect(cv.RawValue()).To(Equal(cv.value))
	})

	It("code challenge does not contain + / or =", func() {
		cv, _ := NewCodeVerifier(50, 100)

		c1, err := cv.CodeChallenge()
		Expect(err).To(BeNil())
		Expect(c1).NotTo(ContainSubstring("+"))
		Expect(c1).NotTo(ContainSubstring("/"))
		Expect(c1).NotTo(ContainSubstring("="))
	})

	It("code verifiers are unique", func() {
		cv1, _ := NewCodeVerifier(50, 100)
		cv2, _ := NewCodeVerifier(50, 100)
		cv3, _ := NewCodeVerifier(50, 100)

		Expect(cv1.RawValue()).NotTo(Equal(cv2.RawValue()))
		Expect(cv1.RawValue()).NotTo(Equal(cv3.RawValue()))
		Expect(cv2.RawValue()).NotTo(Equal(cv3.RawValue()))
	})

})
