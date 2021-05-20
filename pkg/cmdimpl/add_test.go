package cmdimpl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

const testKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

var FailFluxHandler = &fluxopsfakes.FakeFluxHandler{
	HandleStub: func(arglist string) ([]byte, error) {
		commandEnd := strings.Index(arglist, " ")
		command := arglist[0:commandEnd]
		if strings.HasPrefix(command, "install") || strings.HasPrefix(command, "add") {
			return nil, fmt.Errorf("failed")
		}
		return fluxops.DefaultFluxHandler{}.Handle(arglist)
	},
}

type failGitProviderHandler struct{}

func (h failGitProviderHandler) CreateOrgRepository(_ gitprovider.Client, _ gitprovider.OrgRepositoryRef, _ gitprovider.RepositoryInfo, _ ...gitprovider.RepositoryCreateOption) error {
	shims.Exit(1)
	return nil
}

type statusHandler struct{}

func (h statusHandler) GetClusterName() (string, error) {
	return "test", nil
}

func (h statusHandler) GetClusterStatus() status.ClusterStatus {
	return status.FluxInstalled
}

func createTestPrivateKeyFile() (*os.File, error) {
	tmpFile, err := ioutil.TempFile("", "private-key")
	if err != nil {
		return nil, err
	}
	return tmpFile, ioutil.WriteFile(tmpFile.Name(), []byte(testKey), 0600)
}

func ensureFluxVersion() error {
	path := os.Getenv("GITHUB_WORKSPACE")
	if path == "" {
		path = "../.."
	}
	if version.FluxVersion == "undefined" {
		// stoml hasn't been downloaded when unit tests run
		stomldir, err := ioutil.TempDir("", "stoml")
		if err != nil {
			return err
		}
		defer os.RemoveAll(stomldir)

		stomlpath := filepath.Join(stomldir, "stoml")
		stomlurl := fmt.Sprintf("https://github.com/freshautomations/stoml/releases/download/v0.4.0/stoml_%s_amd64", runtime.GOOS)

		err = utils.CallCommandForEffectWithDebug(fmt.Sprintf("curl --progress-bar -fLo %s %s", stomlpath, stomlurl))
		if err != nil {
			return err
		}

		err = utils.CallCommandForEffectWithDebug(fmt.Sprintf("chmod +x %s", stomlpath))
		if err != nil {
			return err
		}

		deppath, err := filepath.Abs(path + "/tools/dependencies.toml")
		if err != nil {
			return err
		}
		out, err := utils.CallCommand(fmt.Sprintf("%s %s flux.version", stomlpath, deppath))
		if err != nil {
			return err
		}
		version.FluxVersion = strings.TrimRight(string(out), "\n")
		flux.SetupFluxBin()
	}
	return nil
}

func handleGitLsRemote(arglist ...interface{}) ([]byte, []byte, error) {
	commandEnd := strings.Index(arglist[0].(string), " ")
	command := arglist[0].(string)[0:commandEnd]
	if strings.HasPrefix(command, "git ls-remote") {
		return []byte{}, []byte{}, nil
	}
	return nil, nil, fmt.Errorf("NO!")
}

var _ = Describe("Dry Run Add Test", func() {
	It("Verify that the dry-run flag leaves clusters and repos unchanged", func() {
		By("Executing a dry-run add and failing/exiting if any of the flux actions were invoked", func() {
			Expect(os.Setenv("GITHUB_ORG", "archaeopteryx")).Should(Succeed())
			Expect(os.Setenv("GITHUB_TOKEN", "archaeopteryx")).Should(Succeed())
			Expect(ensureFluxVersion()).Should(Succeed())
			fgphandler := failGitProviderHandler{}
			shandler := statusHandler{}
			privateKeyFile, err := createTestPrivateKeyFile()
			Expect(err).To(BeNil())
			privateKeyFileName := privateKeyFile.Name()
			defer os.Remove(privateKeyFileName)
			_ = override.WithOverrides(
				func() override.Result {
					Add([]string{"."},
						AddParamSet{
							Name:       "wanda",
							Url:        "ssh://git@github.com/foobar/quux.git",
							Path:       "./",
							Branch:     "main",
							PrivateKey: privateKeyFileName,
							DryRun:     true,
							Namespace:  "wego-system"})
					return override.Result{}
				},
				utils.OverrideFailure(utils.CallCommandForEffectWithInputPipeOp),
				utils.OverrideFailure(utils.CallCommandForEffectWithDebugOp),
				utils.OverrideBehavior(utils.CallCommandForEffectOp, handleGitLsRemote),
				fluxops.Override(FailFluxHandler),
				shims.OverrideGitProvider(fgphandler),
				status.Override(shandler))
		})
	})
})
