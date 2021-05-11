package shims

// A package holding our "shims" that add a level of indirection so that parts of the code can be mocked out

import (
	"os"
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
