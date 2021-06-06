package cmdimpl

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	// "github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	// "github.com/go-git/go-git/v5/config"
	// "github.com/go-git/go-git/v5/storage/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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
		return nil, nil
	},
}

var access bool

type statusHandler struct{}

func (h statusHandler) GetClusterName() (string, error) {
	return "test", nil
}

func (h statusHandler) GetClusterStatus() status.ClusterStatus {
	return status.FluxInstalled
}

type fakeGitRepoHandler struct{}

func (h fakeGitRepoHandler) CreateRepository(name string, owner string, private bool) error {
	access = private
	return nil
}

func (h fakeGitRepoHandler) RepositoryExists(ame string, owner string) (bool, error) {
	return false, gitprovider.ErrNotFound
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

var fakeGitClient = gitfakes.FakeGit{
	CloneStub: func(ctx context.Context, a, b, c string) (bool, error) {
		fmt.Println("failing clone")
		shims.Exit(1)
		return false, nil
	},
	CommitStub: func(commit git.Commit) (string, error) {
		fmt.Println("failing commit")
		shims.Exit(1)
		return "", nil
	},
	HeadStub: func() (string, error) {
		fmt.Println("failing head")
		shims.Exit(1)
		return "", nil
	},
	InitStub: func(a, b, c string) (bool, error) {
		fmt.Println("failing init")
		shims.Exit(1)
		return false, nil
	},
	OpenStub: func(a string) (*gogit.Repository, error) {
		fmt.Println("failing open")
		shims.Exit(1)
		return nil, nil
	},
	PushStub: func(ctx context.Context) error {
		fmt.Println("failing push")
		shims.Exit(1)
		return nil
	},
	StatusStub: func() (bool, error) {
		fmt.Println("failing status")
		shims.Exit(1)
		return false, nil
	},
	WriteStub: func(a string, b []byte) error {
		fmt.Println("failing write")
		shims.Exit(1)
		return nil
	},
}

var _ = Describe("Test helm manifest from git repo", func() {
	It("Verify helm manifest files generation from git ", func() {

		expected := `create helmrelease simple-name-dot-my-chart \
            --source="GitRepository/simple-name" \
            --chart="./my-chart" \
            --interval=5m \
            --export \
            --namespace=wego-system`

		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				Expect(args).Should(Equal(expected))
				return []byte("foo"), nil
			},
		}

		_ = override.WithOverrides(
			func() override.Result {
				params.DryRun = false
				params.Namespace = "wego-system"
				Expect(generateHelmManifest("simple-name", "./my-chart")).Should(Equal([]byte("foo")))
				return override.Result{}
			},
			fluxops.Override(fakeHandler))
	})
})

var _ = Describe("Test source manifest", func() {
	It("Verify source manifest files generation ", func() {
		secretCall := true
		expectedSecret := `create secret git "sname" \
            --url="ssh://git@github.com/auser/arepo" \
            --private-key-file="/tmp/pkey" \
            --namespace="aNamespace"`

		expectedSource := `create source git "sname" \
            --url="ssh://git@github.com/auser/arepo" \
            --branch="aBranch" \
            --secret-ref="sname" \
            --interval=30s \
            --export \
            --namespace="aNamespace"`
		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				if secretCall {
					Expect(args).Should(Equal(expectedSecret))
					secretCall = false
					return nil, nil
				} else {
					Expect(args).Should(Equal(expectedSource))
					return []byte("bar"), nil
				}
			},
		}

		_ = override.WithOverrides(
			func() override.Result {
				params.DryRun = false
				params.Namespace = "aNamespace"
				params.PrivateKey = "/tmp/pkey"
				params.Branch = "aBranch"

				// source type will come into play when we have helmrepo support
				Expect(generateSource(
					"sname", "ssh://git@github.com/auser/arepo", "git")).Should(Equal([]byte("bar")))
			})
	})
})

var _ = Describe("Test helm manifest from helm repo", func() {
	It("Verify helm manifest generation from helm ", func() {

		expected := `create helmrelease simple-name \
            --source="HelmRepository/simple-name" \
            --chart="testchart" \
            --interval=5m \
            --export \
            --namespace=wego-system`

		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				Expect(args).Should(Equal(expected))
				return []byte("foo"), nil
			},
		}

		_ = override.WithOverrides(
			func() override.Result {
				params.DryRun = false
				params.Name = "simple-name"
				params.Namespace = "wego-system"
				params.Chart = "testchart"
				Expect(generateHelmManifestHelm()).Should(Equal([]byte("foo")))
				return override.Result{}
			},
			fluxops.Override(fakeHandler))
	})

})

var _ = Describe("Test helm source from helm repo", func() {
	It("Verify helm source generation from helm ", func() {

		expected := `create source helm test \
            --url="https://github.io/testrepo" \
            --interval=30s \
            --export \
            --namespace=wego-system `

		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				Expect(args).Should(Equal(expected))
				return []byte("foo"), nil
			},
		}

		_ = override.WithOverrides(
			func() override.Result {
				params.DryRun = false
				params.Name = "test"
				params.Url = "https://github.io/testrepo"
				params.Namespace = "wego-system"
				params.Chart = "testChart"
				Expect(generateSourceManifestHelm()).Should(Equal([]byte("foo")))
				return override.Result{}
			},
			fluxops.Override(fakeHandler))
	})

})

var _ = Describe("Dry Run Add Test", func() {
	It("Verify that the dry-run flag leaves clusters and repos unchanged", func() {
		By("Executing a dry-run add and failing/exiting if any of the flux actions were invoked", func() {
			Expect(os.Setenv("GITHUB_ORG", "archaeopteryx")).Should(Succeed())
			Expect(os.Setenv("GITHUB_TOKEN", "archaeopteryx")).Should(Succeed())
			Expect(ensureFluxVersion()).Should(Succeed())
			fgphandler := fakeGitRepoHandler{}
			shandler := statusHandler{}
			privateKeyFile, err := createTestPrivateKeyFile()
			Expect(err).To(BeNil())
			privateKeyFileName := privateKeyFile.Name()
			defer os.Remove(privateKeyFileName)
			_ = override.WithOverrides(
				func() override.Result {
					deps := &AddDependencies{
						GitClient: &fakeGitClient,
					}

					err := Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "ssh://git@github.com/foobar/quux.git",
							Path:           "./",
							Branch:         "main",
							PrivateKey:     privateKeyFileName,
							DryRun:         true,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					return override.Result{}
				},
				utils.OverrideFailure(utils.CallCommandForEffectWithInputPipeOp),
				utils.OverrideFailure(utils.CallCommandForEffectWithDebugOp),
				utils.OverrideBehavior(utils.CallCommandForEffectOp, handleGitLsRemote),
				utils.OverrideBehavior(utils.CallCommandSeparatingOutputStreamsOp,
					func(args ...interface{}) ([]byte, []byte, error) {
						case0Kubectl := `kubectl config current-context`
						Expect(args[0].(string)).Should(Equal(case0Kubectl))
						switch (args[0]).(string) {
						case case0Kubectl:
							return []byte("my-cluster"), []byte(""), nil
						default:
							return nil, nil, fmt.Errorf("arguments not expected %s", args)
						}

					}),
				fluxops.Override(FailFluxHandler),
				gitproviders.Override(fgphandler),
				status.Override(shandler))
		})
	})
})
