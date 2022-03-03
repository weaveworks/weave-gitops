package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/git"
)

var td string
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
	Describe("Test file and folder common utils", func() {
		BeforeEach(func() {
			td, err := ioutil.TempDir("", "common_test-")
			Expect(err).ShouldNot(HaveOccurred())
			defer os.RemoveAll(td)
		})
		It("can check to see if a file exists or not", func() {

			// Existing file
			tempFile, err := ioutil.TempFile(td, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(Exists(tempFile.Name())).To(BeTrue())

			// Not existing file
			Expect(os.Remove(tempFile.Name())).ShouldNot(HaveOccurred())
			Expect(Exists(tempFile.Name())).To(BeFalse())
		})
		It("can check to see if a folder exists or not", func() {
			// Existing folder
			tempFolder, err := ioutil.TempDir(td, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(Exists(tempFolder)).To(BeTrue())

			// Not existing folder
			Expect(os.Remove(tempFolder)).ShouldNot(HaveOccurred())
			Expect(Exists(tempFolder)).To(BeFalse())

		})
	})
	Describe("Convert old paths to new directory structure", func() {

		It("correctly translates multiple paths into new structure", func() {
			tests := []struct {
				orig string
				exp  string
			}{
				{"foo", "foo"},
				{filepath.Join("apps", "foo", "foo.yaml"), filepath.Join(git.WegoRoot, git.WegoAppDir, "foo", "foo.yaml")},
				{filepath.Join(".wego", "apps", "foo", "foo.yaml"), filepath.Join(git.WegoRoot, git.WegoAppDir, "foo", "foo.yaml")},
				{filepath.Join("targets", "mycluster", "foo", "deploy.yaml"), filepath.Join(git.WegoRoot, git.WegoAppDir, "foo", "deploy.yaml")},
				{filepath.Join(".wego", "targets", "mycluster", "foo", "source.yaml"), filepath.Join(git.WegoRoot, git.WegoAppDir, "foo", "source.yaml")},
				{"", ""},
			}
			for _, i := range tests {
				Expect(MigrateToNewDirStructure(i.orig)).To(Equal(i.exp))
			}
		})
	})
})
