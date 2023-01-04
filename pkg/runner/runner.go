package runner

import (
	"os/exec"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Runner is an entity which can execute commands with whatever backing medium behind it.
//
//counterfeiter:generate . Runner
type Runner interface {
	// Run takes a command name and some arguments and will apply it to the current environment.
	Run(command string, args ...string) ([]byte, error)
}

// CLIRunner will use exec.Command as runtime medium.
type CLIRunner struct{}

// Run will use exec.Command as a backing medium and run CombinedOutput.
func (*CLIRunner) Run(c string, args ...string) ([]byte, error) {
	cmd := exec.Command(c, args...)
	return cmd.CombinedOutput()
}
