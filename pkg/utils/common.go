package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
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

func GetAppHash(url, path string) string {
	h := md5.New()
	h.Write([]byte(url + path))
	return hex.EncodeToString(h.Sum(nil))
}

func SetCommmitMessageFromArgs(cmd string, args []string, name string) {
	commitMessage = fmt.Sprintf("wego %s %s %s", cmd, args, name)
}

func GetCommitMessage() string {
	return commitMessage
}
