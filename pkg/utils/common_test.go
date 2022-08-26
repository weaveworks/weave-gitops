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

	Describe("FindCoreConfig", func() {
		var dir string

		BeforeEach(func() {
			var err error
			dir, err = os.MkdirTemp("", "find-core-config")
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
