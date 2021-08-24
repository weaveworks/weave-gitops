package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

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
		Expect(time.Since(start)).Should(BeNumerically("<=", time.Millisecond*3))
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
