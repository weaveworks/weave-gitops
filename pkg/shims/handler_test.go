package shims

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Log Tests")
}

var _ = Describe("Stream Shim Tests", func() {
	It("Verify that replacing os streams works correctly", func() {
		By("Checking that normal streams are returned if left alone", func() {
			Expect(os.Stdin).To(Equal(Stdin()))
			Expect(os.Stdout).To(Equal(Stdout()))
			Expect(os.Stderr).To(Equal(Stderr()))
		})
		By("Replacing the streams and checking the results", func() {
			f, err := ioutil.TempFile("", "stream")
			Expect(err).To(BeNil())
			WithStdin(f, func() {
				WithStdout(f, func() {
					WithStderr(f, func() {
						Expect(Stdin()).To(Equal(f))
						Expect(Stdout()).To(Equal(f))
						Expect(Stderr()).To(Equal(f))
					})
				})
			})
		})
	})
})
