package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"golang.org/x/term"
)

// readPasswordFromStdin reads a password from stdin and returns the input
// with trailing newline and/or carriage return removed. It also makes sure that terminal
// echoing is turned off if stdin is a terminal.
func ReadPasswordFromStdin(log logger.Logger, prompt string) (string, error) {
	var (
		out string
		err error
	)

	fmt.Fprint(os.Stdout, prompt)

	stdinFD := int(os.Stdin.Fd())

	if term.IsTerminal(stdinFD) {
		var inBytes []byte
		inBytes, err = term.ReadPassword(int(os.Stdin.Fd()))
		out = string(inBytes)
	} else {
		out, err = bufio.NewReader(os.Stdin).ReadString('\n')
	}

	if err != nil {
		log.Failuref("could not read from stdin: %v", err)
		return "", err
	}

	fmt.Println()

	return strings.TrimRight(out, "\r\n"), nil
}
