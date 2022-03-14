package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/weaveworks/weave-gitops/pkg/git"

	validation "k8s.io/apimachinery/pkg/api/validation"
)

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// WaitUntil runs checkDone until an error is NOT returned, or a timeout is reached.

// To continue polling, return an error.
func WaitUntil(out io.Writer, poll, timeout time.Duration, checkDone func() error) error {
	_, err := timedRepeat(
		out,
		time.Now(),
		poll,
		timeout,
		func(currentTime time.Time) time.Time {
			time.Sleep(poll)
			return time.Now()
		},
		checkDone)

	return err
}

// timedRepeat runs checkDone until a timeout is reached by updating the current time via a specified operation
func timedRepeat(out io.Writer, start time.Time, poll, timeout time.Duration, updater func(currentTime time.Time) time.Time, checkDone func() error) (time.Time, error) {
	currentTime := start
	endTime := currentTime.Add(timeout)

	for ; currentTime.Before(endTime); currentTime = updater(currentTime) {
		err := checkDone()
		if err == nil {
			return currentTime, nil
		}

		fmt.Fprintf(out, "error occurred %s, retrying in %s\n", err, poll.String())
	}

	return currentTime, fmt.Errorf("timeout reached %s", timeout.String())
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

func UrlToRepoName(url string) string {
	return strings.TrimSuffix(filepath.Base(url), ".git")
}

func ValidateNamespace(ns string) error {
	if errList := validation.ValidateNamespaceName(ns, false); len(errList) != 0 {
		return fmt.Errorf("invalid namespace: %s", strings.Join(errList, ", "))
	}

	return nil
}

func PrintTable(writer io.Writer, header []string, rows [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
}

func MigrateToNewDirStructure(orig string) string {
	if orig == "" {
		return orig
	}

	f := strings.Split(orig, string(os.PathSeparator))

	switch len(f) {
	case 1:
		// single file
		return orig
	case 2:
		// handles the case apps/ and clusters/
		return filepath.Join(git.WegoRoot, orig)
	default:
		// used for paths with apps under clusters
		return filepath.Join(git.WegoRoot, git.WegoAppDir, f[len(f)-2], f[len(f)-1])
	}
}
