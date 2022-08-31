package runner_test

import (
	"io"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runner Suite")
}

func CaptureStdout(c func()) string {
	r, w, _ := os.Pipe()
	tmp := os.Stdout

	defer func() {
		os.Stdout = tmp
	}()

	os.Stdout = w

	c()

	w.Close()

	stdout, _ := io.ReadAll(r)

	return string(stdout)
}
