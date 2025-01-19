package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomwright/dasel/v2"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/validation"
)

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

func URLToRepoName(url string) string {
	return strings.TrimSuffix(filepath.Base(url), ".git")
}

func ValidateNamespace(ns string) error {
	if errList := validation.ValidateNamespaceName(ns, false); len(errList) != 0 {
		return fmt.Errorf("invalid namespace: %s", strings.Join(errList, ", "))
	}

	return nil
}

const (
	coreManifestCount = 2
	coreManifestName  = "ww-gitops"
)

type ConfigStatus int

const (
	Missing ConfigStatus = iota
	Partial
	Embedded
	Valid
)

func (cs ConfigStatus) String() string {
	switch cs {
	case Missing:
		return "Missing"
	case Partial:
		return "Partial"
	case Embedded:
		return "Embedded"
	case Valid:
		return "Valid"
	default:
		return "UnknownStatus"
	}
}

type WalkResult struct {
	Status ConfigStatus
	Path   string
}

func (wr WalkResult) Error() string {
	return fmt.Sprintf("found %s: with status: %s", wr.Path, wr.Status)
}

func FindCoreConfig(dir string) WalkResult {
	err := filepath.WalkDir(dir,
		func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			r := bytes.NewReader(data)
			decoder := yaml.NewDecoder(r)
			docs := []map[string]interface{}{}

			for {
				var entry map[string]interface{}
				if err := decoder.Decode(&entry); errors.Is(err, io.EOF) {
					break
				}

				docs = append(docs, entry)
			}

			rootNode := dasel.ValueOf(docs)
			var foundHelmRelease, foundHelmRepository bool

			val, err := dasel.Select(rootNode, fmt.Sprintf("all().filter(equal(kind,HelmRelease),equal(metadata.name,%s)).count()", coreManifestName))
			if err != nil {
				return nil
			}
			foundHelmRelease = val.Interfaces()[0] != 0

			val, err = dasel.Select(rootNode, fmt.Sprintf("all().filter(equal(kind,HelmRepository),equal(metadata.name,%s)).count()", coreManifestName))
			if err != nil {
				return nil
			}
			foundHelmRepository = val.Interfaces()[0] != 0

			if foundHelmRelease != foundHelmRepository {
				return WalkResult{Status: Partial, Path: path}
			}
			if !foundHelmRelease && !foundHelmRepository {
				return nil
			}

			// retrieve the number of top-level entries from the file
			val, err = dasel.Select(rootNode, "all().count()")
			if err != nil {
				return nil
			}

			if val.Interfaces()[0] != coreManifestCount {
				return WalkResult{Status: Embedded, Path: path}
			}

			return WalkResult{Status: Valid, Path: path}
		})

	var val WalkResult
	if errors.As(err, &val) {
		return val
	}

	return WalkResult{Status: Missing, Path: ""}
}
