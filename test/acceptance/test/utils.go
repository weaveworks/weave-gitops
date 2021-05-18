/**
* All common util functions and golbal constants will go here.
**/
package acceptance

import (
	"os"
	"os/exec"
	"time"
)

const EVENTUALLY_DEFAULT_TIME_OUT time.Duration = 60 * time.Second

var WEGO_BIN_PATH string

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// showItems displays the current set of a specified object type in tabular format
func showItems(itemType string) {
	runCommandPassThrough([]string{}, "kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
}

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
