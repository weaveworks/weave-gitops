package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Code verifier", func() {
	It("raw value is greater than or equal to 50", func() {
		cv, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(cv.RawValue())).To(BeNumerically(">=", 50))
	})

	It("raw value is less than 100", func() {
		cv, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(cv.RawValue())).To(BeNumerically("<", 100))
	})

	It("raw value is internal value", func() {
		cv, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())
		Expect(cv.RawValue()).To(Equal(cv.value))
	})

	It("code challenge does not contain + / or =", func() {
		cv, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())

		c1, err := cv.CodeChallenge()
		Expect(err).NotTo(HaveOccurred())
		Expect(c1).NotTo(ContainSubstring("+"))
		Expect(c1).NotTo(ContainSubstring("/"))
		Expect(c1).NotTo(ContainSubstring("="))
	})

	It("code verifiers are unique", func() {
		cv1, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())
		cv2, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())
		cv3, err := NewCodeVerifier(50, 100)
		Expect(err).NotTo(HaveOccurred())

		Expect(cv1.RawValue()).NotTo(Equal(cv2.RawValue()))
		Expect(cv1.RawValue()).NotTo(Equal(cv3.RawValue()))
		Expect(cv2.RawValue()).NotTo(Equal(cv3.RawValue()))
	})
})
