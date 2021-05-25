package shims

// A package holding our "shims" that add a level of indirection so that parts of the code can be mocked out

import (
	"os"

	"github.com/weaveworks/weave-gitops/pkg/override"
)

// Handler for mocking os.Exit()
type ExitHandler interface {
	Handle(code int)
}

type defaultExitHandler struct{}

func (h defaultExitHandler) Handle(code int) {
	os.Exit(code)
}

var exitHandler interface{} = defaultExitHandler{}

func OverrideExit(handler ExitHandler) override.Override {
	return override.Override{Handler: &exitHandler, Mock: handler, Original: exitHandler}
}

// Handler implementation to ignore exits
type IgnoreExitHandler struct{}

func (h IgnoreExitHandler) Handle(code int) {
}

// Function being mocked
func Exit(code int) {
	exitHandler.(ExitHandler).Handle(code)
}

// Handler for mocking os.UserHomeDir()
type HomeDirHandler interface {
	Handle() (string, error)
}

type defaultHomeDirHandler struct{}

var homeDirHandler interface{} = defaultHomeDirHandler{}

func (h defaultHomeDirHandler) Handle() (string, error) {
	return os.UserHomeDir()
}

func OverrideHomeDir(handler HomeDirHandler) override.Override {
	return override.Override{Handler: &homeDirHandler, Mock: handler, Original: homeDirHandler}
}

// Function being mocked
func UserHomeDir() (string, error) {
	return homeDirHandler.(HomeDirHandler).Handle()
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
