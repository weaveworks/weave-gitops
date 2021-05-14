package shims

// A package holding our "shims" that add a level of indirection so that parts of the code can be mocked out

import (
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	cgitprovider "github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

// Handler for mocking os.Exit()
type ExitHandler interface {
	Handle(code int)
}

type defaultExitHandler struct{}

var exitHandler ExitHandler = defaultExitHandler{}

func (h defaultExitHandler) Handle(code int) {
	os.Exit(code)
}

type IgnoreExitHandler struct{}

func (h IgnoreExitHandler) Handle(code int) {
}

func WithExitHandler(handler ExitHandler, fun func()) {
	originalHandler := exitHandler
	exitHandler = handler
	defer func() {
		exitHandler = originalHandler
	}()
	fun()
}

func Exit(code int) {
	exitHandler.Handle(code)
}

// Handler for mocking os.UserHomeDir()
type HomeDirHandler interface {
	Handle() (string, error)
}

type defaultHomeDirHandler struct{}

var homeDirHandler HomeDirHandler = defaultHomeDirHandler{}

func (h defaultHomeDirHandler) Handle() (string, error) {
	return os.UserHomeDir()
}

func WithHomeDirHandler(handler HomeDirHandler, fun func() (string, error)) (string, error) {
	originalHandler := homeDirHandler
	homeDirHandler = handler
	defer func() {
		homeDirHandler = originalHandler
	}()
	return fun()
}

func UserHomeDir() (string, error) {
	return homeDirHandler.Handle()
}

// GitProvider shim
type GitProviderHandler interface {
	CreateOrgRepository(provider gitprovider.Client, orgRepoRef gitprovider.OrgRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) error
}

type defaultGitProviderHandler struct{}

var gitProviderHandler GitProviderHandler = defaultGitProviderHandler{}

func (h defaultGitProviderHandler) CreateOrgRepository(provider gitprovider.Client, orgRepoRef gitprovider.OrgRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) error {
	return cgitprovider.CreateOrgRepository(provider, orgRepoRef, repoInfo, opts...)
}

func WithGitProviderHandler(handler GitProviderHandler, fun func() error) error {
	originalHandler := gitProviderHandler
	gitProviderHandler = handler
	defer func() {
		gitProviderHandler = originalHandler
	}()
	return fun()
}

func CreateOrgRepository(provider gitprovider.Client, orgRepoRef gitprovider.OrgRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) error {
	return gitProviderHandler.CreateOrgRepository(provider, orgRepoRef, repoInfo, opts...)
}

type FileStreams struct {
	stdin, stdout, stderr *os.File
}

var fileStreams = FileStreams{}

func Stdin() *os.File {
	if fileStreams.stdin != nil {
		return fileStreams.stdin
	}
	return os.Stdin
}

func Stdout() *os.File {
	if fileStreams.stdout != nil {
		return fileStreams.stdout
	}
	return os.Stdout
}

func Stderr() *os.File {
	if fileStreams.stderr != nil {
		return fileStreams.stderr
	}
	return os.Stderr
}

func WithStdin(stdin *os.File, fun func()) {
	originalStdin := fileStreams.stdin
	fileStreams.stdin = stdin
	defer func() {
		fileStreams.stdin = originalStdin
	}()
	fun()
}

func WithStdout(stdout *os.File, fun func()) {
	originalStdout := fileStreams.stdout
	fileStreams.stdout = stdout
	defer func() {
		fileStreams.stdout = originalStdout
	}()
	fun()
}

func WithStderr(stderr *os.File, fun func()) {
	originalStderr := fileStreams.stderr
	fileStreams.stderr = stderr
	defer func() {
		fileStreams.stderr = originalStderr
	}()
	fun()
}
