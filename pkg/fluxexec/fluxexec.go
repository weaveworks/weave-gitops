// This package is a port of TF-Exec, but for Flux.

package fluxexec

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

type Flux struct {
	execPath   string
	workingDir string
	env        map[string]string

	logger logr.Logger
}

func NewFlux(workingDir, execPath string) (*Flux, error) {
	if workingDir == "" {
		return nil, fmt.Errorf("flux cannot be initialised with empty workdir")
	}

	if _, err := os.Stat(workingDir); err != nil {
		return nil, fmt.Errorf("error initialising Flux with workdir %s: %w", workingDir, err)
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
		env:        make(map[string]string),
		logger:     logr.Discard(),
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
func (flux *Flux) SetLogger(logger logr.Logger) {
	flux.logger = logger
}

func (flux *Flux) SetEnvVar(key, value string) {
	flux.env[key] = value
}
