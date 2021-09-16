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

	It("Verify timedRepeat succeeds at first attempt without updating the current time", func() {

		var output bytes.Buffer
		start := time.Now()
		resultTime, err := timedRepeat(
			&output,
			start,
			time.Millisecond,
			time.Millisecond,
			func(currentTime time.Time) time.Time {
				return currentTime.Add(time.Millisecond)
			},
			func() error {
				return nil
			})
		Expect(resultTime.Sub(start)).Should(BeNumerically("==", 0))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(output.String()).To(BeEmpty())

	})

	It("Verify timedRepeat prints out proper messages after succeeding at second attempt", func() {

		counter := 0

		var output bytes.Buffer
		start := time.Now()
		resultTime, err := timedRepeat(
			&output,
			start,
			time.Millisecond,
			time.Millisecond*10,
			func(currentTime time.Time) time.Time {
				return currentTime.Add(time.Millisecond)
			},
			func() error {
				if counter == 0 {
					counter++
					return fmt.Errorf("some error")
				}
				return nil
			})
		Expect(resultTime.Sub(start)).Should(BeNumerically("==", time.Millisecond))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(output.String()).To(Equal("error occurred some error, retrying in 1ms\n"))

	})

	It("Verify timedRepeat prints out proper messages after reaching limit", func() {

		var output bytes.Buffer
		start := time.Now()
		resultTime, err := timedRepeat(
			&output,
			start,
			time.Second,
			time.Second*2,
			func(currentTime time.Time) time.Time {
				return currentTime.Add(time.Second)
			},
			func() error {
				return fmt.Errorf("some error")
			})
		Expect(resultTime.Sub(start)).Should(BeNumerically("==", time.Second*2))
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
