package flux

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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

func TestFlux(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flux Tests")
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
			_, _, err := utils.WithResultsFrom(utils.CallCommandOp, testFluxLogResponse, nil, nil, processStatus)
			Expect(err).To(BeNil())
		})
	})
	It("Verify that the bulk namespace operation works correctly on the failure path", func() {
		By("Invoking the operation with a mock command", func() {
			_, _, err := utils.WithResultsFrom(utils.CallCommandOp, nil, nil, fmt.Errorf("failed"), processStatus)
			Expect(err).To(Not(BeNil()))
		})
	})
})

func processStatus() ([]byte, []byte, error) {
	strs, err := GetLatestStatusAllNamespaces()
	if err != nil {
		return nil, nil, err
	} else {
		return []byte(strs[0]), nil, nil
	}
}

// Test Setup

type localExitHandler struct {
	action func(int)
}

func (h localExitHandler) Handle(code int) {
	h.action(code)
}

type localHomeDirHandler struct {
	action func() (string, error)
}

func (h localHomeDirHandler) Handle() (string, error) {
	return h.action()
}

func TestSetup(t *testing.T) {
	_, err := GetFluxBinPath()
	require.NoError(t, err)

	_, err = GetFluxExePath()
	require.NoError(t, err)
}

func TestSetupFluxBin(t *testing.T) {
	version.FluxVersion = "0.11.0"
	SetupFluxBin()
	homeDir, err := shims.UserHomeDir()
	require.NoError(t, err)

	fluxPath := fmt.Sprintf("%v/.wego/bin", homeDir)
	require.DirExists(t, fluxPath)
	binPath := fmt.Sprintf("%v/flux-%v", fluxPath, version.FluxVersion)
	require.FileExists(t, binPath)

	version.FluxVersion = "0.12.0"
	SetupFluxBin()
	require.NoFileExists(t, binPath)
	binPath = fmt.Sprintf("%v/flux-%v", fluxPath, version.FluxVersion)
	require.FileExists(t, binPath)
}

var _ = Describe("Flux Setup Failure", func() {
	It("Verify that exit is called with expected code", func() {
		By("Executing a code path that contains checkError", func() {
			exitCode := -1
			_ = override.WithOverrides(
				func() override.Result {
					checkError(fmt.Errorf("An error"))
					return override.Result{}
				},
				shims.OverrideExit(localExitHandler{action: func(code int) { exitCode = code }}))
			Expect(exitCode).To(Equal(1))
		})
	})

	It("Verify that os.UserHomeDir failures are handled correctly", func() {
		By("Setting the shim to fail and invoking calls that will trigger it", func() {
			res := override.WithOverrides(
				func() override.Result {
					out, err := fluxops.QuietInstall("flux-system")
					return override.Result{Output: out, Err: err}
				},
				shims.OverrideExit(shims.IgnoreExitHandler{}),
				shims.OverrideHomeDir(localHomeDirHandler{action: func() (string, error) { return "", fmt.Errorf("failed") }}))
			Expect(res.Err).To(Not(BeNil()))
		})
	})

})
