package runner

import (
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

// Runner is an entity which can execute commands with whatever backing medium behind it.
//go:generate counterfeiter -o fakes/fake_runner.go . Runner
type Runner interface {
	// Run takes a command name and some arguments and will apply it to the current environment.
	Run(command string, args ...string) ([]byte, error)
	RunWithStdin(command string, args []string, stdinData []byte) ([]byte, error)
}

// CLIRunner will use exec.Command as runtime medium.
type CLIRunner struct{}

// Run will use exec.Command as a backing medium and run CombinedOutput.
func (*CLIRunner) Run(c string, args ...string) ([]byte, error) {
	cmd := exec.Command(c, args...)
	return cmd.CombinedOutput()
}

func (*CLIRunner) RunWithStdin(c string, args []string, stdinData []byte) ([]byte, error) {
	cmd := exec.Command(c, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to get stdinpipe")
	}

	go func() {
		defer stdin.Close()
		_, _ = io.WriteString(stdin, string(stdinData))
	}()

	return cmd.CombinedOutput()
}
