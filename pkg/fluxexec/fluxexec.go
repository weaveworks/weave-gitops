// This package is a port of TF-Exec, but for Flux.

package fluxexec

import (
	"fmt"
	"io"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type Flux struct {
	execPath   string
	workingDir string
	env        map[string]string

	stdout io.Writer
	stderr io.Writer
	logger logger.Logger
}

func NewFlux(workingDir string, execPath string) (*Flux, error) {
	if workingDir == "" {
		return nil, fmt.Errorf("flux cannot be initialised with empty workdir")
	}

	if _, err := os.Stat(workingDir); err != nil {
		return nil, fmt.Errorf("error initialising Flux with workdir %s: %s", workingDir, err)
	}

	if execPath == "" {
		err := fmt.Errorf("please supply the path to a Flux executable using execPath, e.g. using the github.com/weaveworks/weave-gitops/pkg/flux-install module")

		return nil, &ErrNoSuitableBinary{
			err: err,
		}
	}

	flux := Flux{
		execPath:   execPath,
		workingDir: workingDir,
		env:        nil, // explicit nil means copy os.Environ
		logger:     nil,
	}

	return &flux, nil
}

// WorkingDir returns the working directory for Flux.
func (flux *Flux) WorkingDir() string {
	return flux.workingDir
}

// ExecPath returns the path to the Flux executable.
func (flux *Flux) ExecPath() string {
	return flux.execPath
}

// SetLogger specifies a logger for tfexec to use.
func (flux *Flux) SetLogger(logger logger.Logger) {
	flux.logger = logger
}

// SetStdout specifies a writer to stream stdout to for every command.
//
// This should be used for information or logging purposes only, not control
// flow. Any parsing necessary should be added as functionality to this package.
func (flux *Flux) SetStdout(w io.Writer) {
	flux.stdout = w
}

// SetStderr specifies a writer to stream stderr to for every command.
//
// This should be used for information or logging purposes only, not control
// flow. Any parsing necessary should be added as functionality to this package.
func (flux *Flux) SetStderr(w io.Writer) {
	flux.stderr = w
}
