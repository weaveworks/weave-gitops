package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Runner is an entity which can execute commands with whatever backing medium behind it.
//counterfeiter:generate . Runner
type Runner interface {
	// Run takes a command name and some arguments and will apply it to the current environment.
	Run(command string, args ...string) ([]byte, error)
	// RunWithStream takes a command name and some argument and outputs the command in realtime to stdout
	RunWithOutputStream(command string, args ...string) ([]byte, error)
	// RunWithStdin take a command name, arguments and passes stdin data to the command
	RunWithStdin(command string, args []string, stdinData []byte) ([]byte, error)
}

// CLIRunner will use exec.Command as runtime medium.
type CLIRunner struct{}

// Run will use exec.Command as a backing medium and run CombinedOutput.
func (*CLIRunner) Run(c string, args ...string) ([]byte, error) {
	cmd := exec.Command(c, args...)
	return cmd.CombinedOutput()
}

// Run will use exec.Command as a backing medium and run CombinedOutput.
func (*CLIRunner) RunWithOutputStream(c string, args ...string) ([]byte, error) {
	cmd := exec.Command(c, args...)

	var out strings.Builder

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	stdoutScanner := bufio.NewScanner(stdoutReader)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for stdoutScanner.Scan() {
			data := stdoutScanner.Text()
			fmt.Println(data)
			out.WriteString(data)
			out.WriteRune('\n')
		}
	}()

	stderrScanner := bufio.NewScanner(stderrReader)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for stderrScanner.Scan() {
			data := stderrScanner.Text()
			fmt.Println(data)
			out.WriteString(data)
			out.WriteRune('\n')
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	wg.Wait()
	err = cmd.Wait()

	return []byte(out.String()), err
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
