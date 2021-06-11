package runner

import "os/exec"

// Runner is an entity which can execute commands with whatever backing medium behind it.
//go:generate counterfeiter -o fakes/fake_runner.go . Runner
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
