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

var FailFluxHandler = &fluxopsfakes.FakeFluxHandler{
	HandleStub: func(arglist string) ([]byte, error) {
		commandEnd := strings.Index(arglist, " ")
		command := arglist[0:commandEnd]
		if strings.HasPrefix(command, "install") || strings.HasPrefix(command, "add") {
			return nil, fmt.Errorf("failed")
		}
		return []byte(`✚ deploy key: ssh-rsa ID==

► secret 'secret name' created in 'wego-system' namespace`), nil
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

func (h fakeGitRepoHandler) CreateRepository(_ string, _ string, private bool) error {
	access = private
	return nil
}

func (h fakeGitRepoHandler) RepositoryExists(_ string, _ string) (bool, error) {
	return false, gitprovider.ErrNotFound
}

func (h fakeGitRepoHandler) UploadDeployKey(_, _ string, _ []byte) error {
	return nil
}

type fakeGitRepoHandlerDeployKey struct{}

func (h fakeGitRepoHandlerDeployKey) CreateRepository(_ string, _ string, private bool) error {
	return nil
}

func (h fakeGitRepoHandlerDeployKey) RepositoryExists(_ string, _ string) (bool, error) {
	return true, nil
}

func (h fakeGitRepoHandlerDeployKey) UploadDeployKey(_, _ string, _ []byte) error {
	return nil
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

var failGitClient = gitfakes.FakeGit{
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

var ignoreGitClient = gitfakes.FakeGit{
	CloneStub: func(ctx context.Context, a, b, c string) (bool, error) {
		fmt.Println("ignoring clone")
		return false, nil
	},
	CommitStub: func(commit git.Commit) (string, error) {
		fmt.Println("ignoring commit")
		return "", nil
	},
	HeadStub: func() (string, error) {
		fmt.Println("ignoring head")
		return "", nil
	},
	PushStub: func(ctx context.Context) error {
		fmt.Println("ignoring push")
		return nil
	},
	StatusStub: func() (bool, error) {
		fmt.Println("ignoring status")
		return false, nil
	},
	WriteStub: func(a string, b []byte) error {
		fmt.Println("ignoring write")
		return nil
	},
}

var _ = Describe("Test helm manifest from git repo", func() {
	It("Verify helm manifest files generation from git ", func() {
		expected := `create helmrelease simple-name-dot-my-chart \
            --source="GitRepository/source-name" \
            --chart="./my-chart" \
            --interval=1m \
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
				Expect(generateHelmManifestGit("simple-name-dot-my-chart", "source-name", "./my-chart")).Should(Equal([]byte("foo")))
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
					return []byte(`✚ deploy key: ssh-rsa ID==

► secret 'test-repo-with-manifests-path' created in 'wego-system' namespace`), nil
				} else {
					Expect(args).Should(Equal(expectedSource))
					return []byte("bar"), nil
				}
			},
		}

		fgphandler := fakeGitRepoHandlerDeployKey{}
		_ = override.WithOverrides(
			func() override.Result {
				params.DryRun = false
				params.Namespace = "aNamespace"
				params.Branch = "aBranch"

				// source type will come into play when we have helmrepo support

				Expect(generateSource(
					"sname", "ssh://git@github.com/auser/arepo", "git")).Should(Equal([]byte("bar")))
				return override.Result{}
			},
			fluxops.Override(fakeHandler),
			utils.OverrideIgnore(utils.CallCommandForEffectWithInputPipeOp),
			gitproviders.Override(fgphandler),
		)
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
				params.Namespace = "wego-system"
				Expect(generateHelmManifestHelm("simple-name", "testchart")).Should(Equal([]byte("foo")))
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
			_ = override.WithOverrides(
				func() override.Result {
					deps := &AddDependencies{
						GitClient: &failGitClient,
					}

					err := Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "ssh://git@github.com/foobar/quux.git",
							Path:           "./",
							Branch:         "main",
							DryRun:         true,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "",
							Path:           "./foo",
							Branch:         "main",
							DryRun:         true,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "",
							AppConfigUrl:   "none",
							Path:           "./foo",
							Branch:         "main",
							DryRun:         true,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "",
							AppConfigUrl:   "ssh://git@github.com/aUser/aRepo",
							Path:           "./foo",
							Branch:         "main",
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

var _ = Describe("Wet Run Add Test", func() {
	It("Verify that paths through add work correctly when not using --dry-run", func() {
		By("Executing a regular add and ensuring calls work", func() {
			Expect(os.Setenv("GITHUB_ORG", "archaeopteryx")).Should(Succeed())
			Expect(os.Setenv("GITHUB_TOKEN", "archaeopteryx")).Should(Succeed())
			Expect(ensureFluxVersion()).Should(Succeed())
			fgphandler := fakeGitRepoHandlerDeployKey{}
			shandler := statusHandler{}
			_ = override.WithOverrides(
				func() override.Result {
					deps := &AddDependencies{
						GitClient: &ignoreGitClient,
					}
					gitDir, err := ioutil.TempDir("", "git-")
					Expect(err).To(BeNil())
					defer os.RemoveAll(gitDir)
					_, err = deps.GitClient.Init(gitDir, "a url we ignore", "main")
					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "ssh://git@github.com/foobar/quux.git",
							Path:           "./",
							Branch:         "main",
							DryRun:         false,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "",
							Path:           "./foo",
							Branch:         "main",
							DryRun:         false,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "",
							AppConfigUrl:   "none",
							Path:           "./foo",
							Branch:         "main",
							DryRun:         false,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					err = Add([]string{"."},
						AddParamSet{
							Name:           "",
							Url:            "",
							AppConfigUrl:   "ssh://git@github.com/aUser/aRepo",
							Path:           "./foo",
							Branch:         "main",
							DryRun:         false,
							Namespace:      "wego-system",
							DeploymentType: string(DeployTypeKustomize),
						}, deps)

					Expect(err).To(BeNil())
					return override.Result{}
				},
				utils.OverrideIgnore(utils.CallCommandForEffectWithInputPipeOp),
				utils.OverrideIgnore(utils.CallCommandForEffectWithDebugOp),
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
