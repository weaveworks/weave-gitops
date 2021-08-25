package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega/gbytes"

	"github.com/stretchr/testify/require"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestExists(t *testing.T) {

	// Existing file
	tempFile, err := ioutil.TempFile(t.TempDir(), "")
	require.NoError(t, err)
	require.True(t, Exists(tempFile.Name()))

	// Not existing file
	require.NoError(t, os.Remove(tempFile.Name()))
	require.False(t, Exists(tempFile.Name()))

	// Existing file
	tempFolder, err := ioutil.TempDir(t.TempDir(), "")
	require.NoError(t, err)
	require.True(t, Exists(tempFolder))

	// Not existing file
	require.NoError(t, os.Remove(tempFolder))
	require.False(t, Exists(tempFolder))

}

var _ = Describe("Test common utils", func() {

	It("Verify WaitUntil succeeds at  first attempt in less than 1 millisecond", func(done Done) {

		counter := 0
		output := gbytes.NewBuffer()
		go func() {
			err := WaitUntil(output, time.Millisecond, time.Millisecond, func() error {
				counter++
				return nil
			})
			Expect(err).ShouldNot(HaveOccurred())
			close(done)
		}()
		Eventually(output, time.Millisecond*2, time.Millisecond*2).Should(gbytes.Say(""))
		Expect(counter).Should(Equal(1))

	})

	It("Verify WaitUntil prints out proper messages after succeeding at second attempt in less than 1 millisecond", func(done Done) {

		counter := 0
		output := gbytes.NewBuffer()
		go func() {
			err := WaitUntil(output, time.Millisecond, time.Millisecond*2, func() error {
				if counter == 0 {
					counter++
					return fmt.Errorf("some error")
				}
				return nil
			})
			Expect(err).ShouldNot(HaveOccurred())
			close(done)
		}()
		Eventually(output, time.Millisecond*3, time.Millisecond*3).Should(gbytes.Say("error occurred some error, retrying in 1ms\n"))
		Expect(counter).Should(Equal(1))
	})

	It("Verify WaitUntil prints out proper messages after reaching timeout", func(done Done) {

		counter := 0
		output := gbytes.NewBuffer()
		var err error
		go func() {
			defer GinkgoRecover()
			err = WaitUntil(output, time.Millisecond, time.Millisecond*2, func() error {
				counter++
				return fmt.Errorf("some error")
			})
			Expect(err.Error()).Should(Equal("timeout reached 2ms"))
			close(done)
		}()
		Eventually(output, time.Millisecond*3, time.Millisecond*3).Should(gbytes.Say("error occurred some error, retrying in 1ms\nerror occurred some error, retrying in 1ms\n"))
		Expect(counter).Should(Equal(2))
	})

	It("Verify CaptureStdout captures whatever is printed out to stdout in the callback", func() {

		var d = func() {
			fmt.Fprintf(os.Stdout, "my output")
		}

		stdout := CaptureStdout(d)
		Expect(stdout).To(Equal("my output"))

	})
})

var _ = DescribeTable("SanitizeRepoUrl", func(input string, expected string) {
	result := SanitizeRepoUrl(input)
	Expect(result).To(Equal(expected))
},
	Entry("git clone style", "git@github.com:someuser/podinfo.git", "ssh://git@github.com/someuser/podinfo.git"),
	Entry("url style", "ssh://git@github.com/someuser/podinfo.git", "ssh://git@github.com/someuser/podinfo.git"),
	// TODO: there is other code relying on everything looking like an SSH url.
	// We need to refactor the SanitizeRepoUrl function .
	// https://github.com/weaveworks/weave-gitops/issues/577
	Entry("https style", "https://github.com/weaveworks/weave-gitops.git", "ssh://git@github.com/weaveworks/weave-gitops.git"),
)
