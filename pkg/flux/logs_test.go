package flux

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner/runnerfakes"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

var testFluxLogResponse = []byte(`2021-04-12T19:53:58.545Z info Alert - Starting EventSource
2021-04-12T19:53:58.545Z info Receiver - Starting EventSource
2021-04-12T19:53:58.545Z info Provider - Starting EventSource
2021-04-12T19:53:58.646Z info Alert - Starting Controller
2021-04-12T19:53:58.647Z info Alert - Starting workers
2021-04-12T19:53:58.652Z info Provider - Starting Controller
2021-04-12T19:53:58.652Z info Provider - Starting workers
2021-04-12T19:54:23.373Z info GitRepository/flux-system.flux-system - Discarding event, no alerts found for the involved object
2021-04-13T20:37:20.565Z info Kustomization/flux-system.flux-system - Discarding event, no alerts found for the involved object
2021-04-13T20:37:21.213Z info GitRepository/podinfo.flux-system - Discarding event, no alerts found for the involved object
2021-04-13T20:39:30.367Z info GitRepository/flux-system.flux-system - Discarding event, no alerts found for the involved object

2021-04-12T19:54:02.383Z info  - metrics server is starting to listen
2021-04-12T19:54:02.384Z info  - starting manager
2021-04-12T19:54:02.385Z info  - starting metrics server
2021-04-12T19:54:02.486Z info  - starting file server
2021-04-12T19:54:02.486Z info HelmRepository - Starting EventSource
2021-04-12T19:54:02.486Z info Bucket - Starting EventSource
2021-04-12T19:54:02.486Z info HelmChart - Starting EventSource
2021-04-12T19:54:02.486Z info HelmChart - Starting EventSource
2021-04-12T19:54:02.487Z info HelmChart - Starting EventSource
2021-04-12T19:54:02.587Z info GitRepository - Starting workers
2021-04-12T19:54:02.588Z info HelmChart - Starting Controller
2021-04-12T19:54:02.588Z info HelmRepository - Starting workers
2021-04-12T19:54:02.588Z info HelmChart - Starting workers
2021-04-12T19:54:02.588Z info Bucket - Starting Controller
2021-04-12T19:54:02.589Z info Bucket - Starting workers
2021-04-12T21:02:22.808Z info GitRepository/flux-system.flux-system - Reconciliation finished in 873.5428ms, next run in 1m0s
2021-04-12T21:03:23.646Z info GitRepository/flux-system.flux-system - Reconciliation finished in 907.3404ms, next run in 1m0s`)

// Test Setup

const defaultTestFluxVersion = "0.12.0"

var (
	fluxClient *FluxClient
	cliRunner  *runnerfakes.FakeRunner
	osysClient *osysfakes.FakeOsys
	homeDir    string
)

func init() {
	cliRunner = &runnerfakes.FakeRunner{}
	osysClient = &osysfakes.FakeOsys{}
	osysClient.UserHomeDirStub = func() (string, error) {
		return homeDir, nil
	}
	fluxClient = New(osysClient, cliRunner)
}

var _ = Describe("Log Fetching Test", func() {
	It("Verify that log fetching works correctly", func() {
		By("Invoking getLastLogForNamespaces", func() {
			result, err := getLastLogForNamespaces(testFluxLogResponse)
			Expect(err).To(BeNil())
			Expect(len(result)).To(Equal(11))

			emptyResult, err := getLastLogForNamespaces(nil)
			Expect(err).To(BeNil())
			Expect(len(emptyResult)).To(Equal(0))
		})
	})
})

var _ = Describe("Latest Status For All Namespaces Test", func() {
	It("Verify that the bulk namespace operation works correctly on the success path", func() {
		By("Invoking the operation with a mock command", func() {
			cliRunner.RunStub = func(cmd string, args ...string) ([]byte, error) {
				return testFluxLogResponse, nil
			}
			_, err := processStatus()
			Expect(err).To(BeNil())
		})
	})
	It("Verify that the bulk namespace operation works correctly on the failure path", func() {
		By("Invoking the operation with a mock command", func() {
			cliRunner.RunStub = func(cmd string, args ...string) ([]byte, error) {
				return nil, fmt.Errorf("failed")
			}
			_, err := processStatus()
			Expect(err).To(Not(BeNil()))
		})
	})
})

func processStatus() ([]byte, error) {
	strs, err := fluxClient.GetLatestStatusAllNamespaces()
	if err != nil {
		return nil, err
	} else {
		return []byte(strs[0]), nil
	}
}

func TestSetup(t *testing.T) {
	homeDir = "/home/user"
	binPath, err := fluxClient.GetBinPath()
	require.NoError(t, err)
	require.Equal(t, binPath, filepath.Join(homeDir, ".wego", "bin"))

	exePath, err := fluxClient.GetExePath()
	require.NoError(t, err)
	require.Equal(t, exePath, filepath.Join(homeDir, ".wego", "bin", "flux-"+version.FluxVersion))
}

var _ = Describe("Set up flux bin", func() {
	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "a-home-dir")
		Expect(err).ShouldNot(HaveOccurred())
		homeDir = dir
	})

	AfterEach(func() {
		Expect(os.RemoveAll(homeDir)).To(Succeed())
		version.FluxVersion = defaultTestFluxVersion
	})

	Context("Set up flux from embedded binary", func() {
		It("Sets up flux from binary embedded during build", func() {
			Expect(osysClient.Getenv(fluxBinaryPathEnvVar)).Should(Equal(""))

			version.FluxVersion = "0.11.0"
			fluxPath := filepath.Join(homeDir, ".wego", "bin")
			exe11Path := filepath.Join(fluxPath, "flux-"+version.FluxVersion)
			Expect(exe11Path).ShouldNot(BeAnExistingFile())
			Expect(fluxPath).ShouldNot(BeADirectory())
			fluxClient.SetupBin()
			Expect(fluxPath).Should(BeADirectory())
			Expect(exe11Path).Should(BeAnExistingFile())

			version.FluxVersion = defaultTestFluxVersion
			exe12Path := filepath.Join(fluxPath, "flux-"+version.FluxVersion)
			Expect(exe12Path).ShouldNot(BeAnExistingFile())
			fluxClient.SetupBin()
			Expect(exe12Path).Should(BeAnExistingFile())
		})
	})

	Context("Set up flux from binary referenced by env var", func() {
		var envVal string

		BeforeEach(func() {
			osysClient = &osysfakes.FakeOsys{}
			osysClient.UserHomeDirStub = func() (string, error) {
				return homeDir, nil
			}
			osysClient.SetenvStub = func(_, val string) error {
				envVal = val
				return nil
			}
			osysClient.GetenvStub = func(envVar string) string {
				return envVal
			}
			osysClient.ExitStub = func(code int) {}
			fluxClient = New(osysClient, cliRunner)
		})

		It("Fails if passed a bad binary path", func() {
			Expect(osysClient.Setenv(fluxBinaryPathEnvVar, "a-path-pointing-nowhere")).Should(Succeed())
			fluxClient.SetupBin()
			Expect(osysClient.ExitCallCount()).Should(Equal(1))
		})

		It("Copies a referenced binary into the flux executable location", func() {
			dummyBinary := []byte("dummy")
			dummyPath := filepath.Join(homeDir, ".dummyBinary")
			Expect(os.WriteFile(dummyPath, dummyBinary, 0555)).Should(Succeed())
			Expect(osysClient.Setenv(fluxBinaryPathEnvVar, dummyPath)).Should(Succeed())

			fluxClient.SetupBin()
			exePath := filepath.Join(homeDir, ".wego", "bin", "flux-"+version.FluxVersion)
			Expect(osysClient.ExitCallCount()).Should(Equal(0))

			bin, err := os.ReadFile(exePath)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(bin).Should(Equal(dummyBinary))
		})
	})
})
