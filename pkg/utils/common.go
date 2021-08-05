package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	validation "k8s.io/apimachinery/pkg/api/validation"
)

var commitMessage string

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// WaitUntil runs checkDone until a timeout is reached
func WaitUntil(out io.Writer, poll, timeout time.Duration, checkDone func() error) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		err := checkDone()
		if err == nil {
			return nil
		}
		fmt.Fprintf(out, "error occurred %s, retrying in %s\n", err, poll.String())
	}
	return fmt.Errorf("timeout reached %s", timeout.String())
}

type callback func()

func CaptureStdout(c callback) string {
	r, w, _ := os.Pipe()
	tmp := os.Stdout
	defer func() {
		os.Stdout = tmp
	}()
	os.Stdout = w
	c()
	w.Close()
	stdout, _ := ioutil.ReadAll(r)

	return string(stdout)
}

func SetCommmitMessageFromArgs(cmd string, url, path, name string) {
	commitMessage = fmt.Sprintf("%s %s %s %s", cmd, url, path, name)
}

func SetCommmitMessage(msg string) {
	commitMessage = msg
}

func GetCommitMessage() string {
	return commitMessage
}

func UrlToRepoName(url string) string {
	return strings.TrimSuffix(filepath.Base(url), ".git")
}

func GetOwnerFromUrl(url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %s", url)
	}
	return parts[len(parts)-2], nil
}

func ValidateNamespace(ns string) error {
	if errList := validation.ValidateNamespaceName(ns, false); len(errList) != 0 {
		return fmt.Errorf("invalid namespace: %s", strings.Join(errList, ", "))
	}

	return nil
}
