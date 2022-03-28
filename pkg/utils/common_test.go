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

var _ = Describe("Test common utils", func() {
	Describe("timedRepeat", func() {
		It("succeeds at first attempt without updating the current time", func() {
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
			Expect(resultTime.Sub(start)).To(BeNumerically("==", 0))
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(BeEmpty())
		})

		It("prints out proper messages after succeeding at second attempt", func() {
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
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal("error occurred some error, retrying in 1ms\n"))
		})

		It("prints out proper messages after reaching limit", func() {
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
			Expect(err).To(MatchError("timeout reached 2s"))
			Expect(output.String()).Should(Equal(`error occurred some error, retrying in 1s
error occurred some error, retrying in 1s
`))
		})
	})

	Describe("CaptureStdout", func() {
		It("captures whatever is printed out to stdout in the callback", func() {
			var d = func() {
				fmt.Fprintf(os.Stdout, "my output")
			}

			stdout := CaptureStdout(d)
			Expect(stdout).To(Equal("my output"))
		})
	})

	Describe("Test file and folder common utils", func() {
		var td string

		BeforeEach(func() {
			var err error
			td, err = ioutil.TempDir("", "common_test-")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(td)).To(Succeed())
		})

		It("can check to see if a file exists or not", func() {
			// Existing file
			tempFile, err := ioutil.TempFile(td, "subfile")
			Expect(err).ToNot(HaveOccurred())
			Expect(Exists(tempFile.Name())).To(BeTrue())

			// Not existing file
			Expect(os.Remove(tempFile.Name())).To(Succeed())
			Expect(Exists(tempFile.Name())).To(BeFalse())
		})

		It("can check to see if a folder exists or not", func() {
			// Existing folder
			tempFolder, err := ioutil.TempDir(td, "subfolder")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(Exists(tempFolder)).To(BeTrue())

			// Not existing folder
			Expect(os.Remove(tempFolder)).ShouldNot(HaveOccurred())
			Expect(Exists(tempFolder)).To(BeFalse())
		})
	})

	Describe("MigrateToNewDirStructure", func() {
		newStructure := filepath.Join(git.WegoRoot, git.WegoAppDir)

		DescribeTable("correctly translates multiple paths into new structure", func(oldPath, newPath string) {
			Expect(MigrateToNewDirStructure(oldPath)).To(Equal(newPath))
		},
			Entry("does not move random paths", "foo", "foo"),
			Entry("does not fail on empty paths", "", ""),
			Entry(
				"migrates from apps",
				filepath.Join("apps", "foo", "foo.yaml"),
				filepath.Join(newStructure, "foo", "foo.yaml"),
			),
			Entry(
				"migrates from .wego/apps",
				filepath.Join(".wego", "apps", "foo", "foo.yaml"),
				filepath.Join(newStructure, "foo", "foo.yaml"),
			),
			Entry(
				"migrates from under clusters",
				filepath.Join("targets", "mycluster", "foo", "deploy.yaml"),
				filepath.Join(newStructure, "foo", "deploy.yaml"),
			),
			Entry(
				"migrates from under .wego clusters",
				filepath.Join(".wego", "targets", "mycluster", "foo", "source.yaml"),
				filepath.Join(newStructure, "foo", "source.yaml"),
			),
		)
	})

	Describe("FindCoreConfig", func() {
		var dir string

		BeforeEach(func() {
			var err error
			dir, err = ioutil.TempDir("", "find-core-config")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("fails to find core configuration in empty dir", func() {
			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Missing, Path: ""}))
		})

		It("fails to find core configuration in non-empty dir with no config file", func() {
			configFile, err := ioutil.ReadFile("testdata/ingress.yaml")
			Expect(err).NotTo(HaveOccurred())

			path := filepath.Join(dir, "ingress.yaml")
			Expect(ioutil.WriteFile(path, configFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Missing, Path: ""}))
		})

		It("finds embedded core configuration in dir containing file with wrong number of entries", func() {
			configFile, err := ioutil.ReadFile("testdata/config.yaml")
			Expect(err).NotTo(HaveOccurred())

			// Add an extraneous entry
			configFile = append(configFile, []byte("---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: my-ns\n")...)
			path := filepath.Join(dir, "config.yaml")
			Expect(ioutil.WriteFile(path, configFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Embedded, Path: path}))
		})

		It("finds partial core configuration in dir containing file with partial config", func() {
			partialConfigFile, err := ioutil.ReadFile("testdata/partial-config.yaml")
			Expect(err).NotTo(HaveOccurred())

			path := filepath.Join(dir, "config.yaml")
			Expect(ioutil.WriteFile(path, partialConfigFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Embedded, Path: path}))
		})

		It("finds core configuration nested in dir containing one (regardless of name)", func() {
			testConfigFile, err := ioutil.ReadFile("testdata/config.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(os.MkdirAll(filepath.Join(dir, "nested"), 0700)).Should(Succeed())
			renamedConfigFile := filepath.Join(dir, "nested", "sprug.yaml")
			Expect(ioutil.WriteFile(renamedConfigFile, testConfigFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Valid, Path: renamedConfigFile}))
		})
	})
})
