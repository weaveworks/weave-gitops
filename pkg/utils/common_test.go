package utils

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test common utils", func() {
	//It("Verify WaitUntil succeeds at first attempt", func() {
	//
	//	var output bytes.Buffer
	//
	//	start := time.Now()
	//	err := WaitUntil(&output, time.Millisecond, time.Microsecond, func() error {
	//		return nil
	//	})
	//	Expect(err).ShouldNot(HaveOccurred())
	//	Expect(time.Since(start)).Should(match)
	//
	//	Eventually(func() (time.Duration, error) {
	//		start = time.Now()
	//		err := WaitUntil(&output, time.Millisecond, time.Microsecond, func() error {
	//			return nil
	//		})
	//		return time.Since(start), err
	//	}).Should(BeNumerically("<=", time.Millisecond))
	//
	//})

	It("Verify WaitUntil succeeds at first attempt in less than 1 millisecond", func() {

		var output bytes.Buffer
		start := time.Now()

		err := WaitUntil(&output, time.Millisecond, time.Millisecond, func() error {
			return nil
		})
		Expect(time.Since(start)).Should(BeNumerically("<=", time.Millisecond))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(output.String()).To(BeEmpty())

	})

	It("Verify WaitUntil prints out proper messages after succeeding at second attempt in less than 1 millisecond", func() {

		counter := 0

		var output bytes.Buffer
		start := time.Now()
		err := WaitUntil(&output, time.Millisecond, time.Millisecond*10, func() error {
			if counter == 0 {
				counter++
				return fmt.Errorf("some error")
			}
			return nil
		})
		Expect(time.Since(start)).Should(BeNumerically("<=", time.Millisecond*2))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(output.String()).To(Equal("error occurred some error, retrying in 1ms\n"))

	})

	It("Verify WaitUntil prints out proper messages after reaching timeout", func() {

		var output bytes.Buffer
		start := time.Now()
		err := WaitUntil(&output, time.Second, time.Second*2, func() error {
			return fmt.Errorf("some error")
		})
		Expect(time.Since(start)).Should(BeNumerically(">=", time.Second*2))
		Expect(err).Should(MatchError("timeout reached 2s"))
		Expect(output.String()).Should(Equal(`error occurred some error, retrying in 1s
error occurred some error, retrying in 1s
`))

	})
})
