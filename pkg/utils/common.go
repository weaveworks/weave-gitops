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

	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

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

// SanitizeRepoUrl accepts a url like git@github.com:someuser/podinfo.git and converts it into
// a string like ssh://git@github.com/someuser/podinfo.git. This helps standardize the different
// user inputs that might be provided.
func SanitizeRepoUrl(url string) string {
	trimmed := ""

	if !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(url, sshPrefix) {
		trimmed = strings.TrimPrefix(url, sshPrefix)
	}

	httpsPrefix := "https://github.com/"
	if strings.HasPrefix(url, httpsPrefix) {
		trimmed = strings.TrimPrefix(url, httpsPrefix)
	}

	if trimmed != "" {
		return "ssh://git@github.com/" + trimmed
	}

	return url
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

func CleanCommitMessage(msg string) string {
	str := strings.ReplaceAll(msg, "\n", " ")
	if len(str) > 50 {
		str = str[:49] + "..."

	}
	return str
}

func CleanCommitCreatedAt(createdAt time.Time) string {
	return strings.Split(createdAt.String(), " +0000")[0]
}

func ConvertCommitHashToShort(hash string) string {
	return hash[:7]
}

func ConvertCommitURLToShort(url string) string {
	urlArray := strings.SplitAfter(url, "commit/")
	path := urlArray[0]
	hash := urlArray[1][:7]
	return path + hash
}

func CreateRepoSecretName(targetName string, repoURL string) string {
	return fmt.Sprintf("wego-%s-%s", targetName, UrlToRepoName(repoURL))
}

func StartK8sTestEnvironment() (client.Client, func(), error) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../manifests/crds",
			"../../tools/testcrds",
		},
	}

	var err error
	cfg, err := testEnv.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start testEnv: %w", err)
	}

	scheme := kube.CreateScheme()

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		ClientDisableCacheFor: []client.Object{
			&wego.Application{},
			&corev1.Namespace{},
			&corev1.Secret{},
		},
		Scheme: scheme,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create controller manager: %w", err)
	}

	go func() {
		err := k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	return k8sManager.GetClient(), func() {
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	}, nil
}
