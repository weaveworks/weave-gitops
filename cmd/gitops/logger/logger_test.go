package logger

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

var _ = Describe("Logger", func() {
	var (
		log    logger.Logger
		output *bytes.Buffer
	)

	BeforeEach(func() {
		output = &bytes.Buffer{}
		log = NewCLILogger(output)
	})

	It("prints adding newline", func() {
		log.Println("foo")
		Expect(output.String()).To(Equal("foo\n"))
	})

	It("prints as is", func() {
		log.Printf("foo")
		Expect(output.String()).To(Equal("foo"))
	})

	It("prints actions", func() {
		log.Actionf("foo")
		Expect(output.String()).To(Equal("► foo\n"))
	})

	It("prints generate", func() {
		log.Generatef("foo")
		Expect(output.String()).To(Equal("✚ foo\n"))
	})

	It("prints waiting", func() {
		log.Waitingf("foo")
		Expect(output.String()).To(Equal("◎ foo\n"))
	})

	It("prints success", func() {
		log.Successf("foo")
		Expect(output.String()).To(Equal("✔ foo\n"))
	})

	It("prints warning", func() {
		log.Warningf("foo")
		Expect(output.String()).To(Equal("⚠️ foo\n"))
	})

	It("prints failure", func() {
		log.Failuref("foo")
		Expect(output.String()).To(Equal("✗ foo\n"))
	})
})
