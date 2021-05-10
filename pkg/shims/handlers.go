package shims

// A package holding our "shims" that add a level of indirection so that parts of the code can be mocked out

import (
	"os"
)

type ExitHandler interface {
	Handle(code int)
}

type defaultExitHandler struct{}

var exitHandler ExitHandler = defaultExitHandler{}

func (h defaultExitHandler) Handle(code int) {
	os.Exit(code)
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
