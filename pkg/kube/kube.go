package kube

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"encoding/json"

	"github.com/pkg/errors"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Resource interface {
	metav1.Object
	runtime.Object
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const kubectlPath = "kubectl"

type ClusterStatus int

const (
	Unknown ClusterStatus = iota
	Unmodified
	FluxInstalled
	WeGOInstalled
)

// Function to translate ClusterStatus to a string
func (cs ClusterStatus) String() string {
	return toStatusString[cs]
}

var toStatusString = map[ClusterStatus]string{
	Unknown:       "Unable to talk to the cluster",
	Unmodified:    "No flux or wego installed",
	FluxInstalled: "Flux installed",
	WeGOInstalled: "Wego installed",
}

//counterfeiter:generate . Kube
type Kube interface {
	Apply(manifests []byte, namespace string) ([]byte, error)
	Delete(manifests []byte, namespace string) ([]byte, error)
	DeleteByName(name, kind, namespace string) ([]byte, error)
	SecretPresent(ctx context.Context, string, namespace string) (bool, error)
	GetApplications(ctx context.Context, namespace string) ([]wego.Application, error)
	FluxPresent(ctx context.Context) (bool, error)
	GetClusterName(ctx context.Context) (string, error)
	GetClusterStatus(ctx context.Context) ClusterStatus
	GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error)
	GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error
	GetSecret(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error)
}

type KubeClient struct {
	runner runner.Runner
}

func New(cliRunner runner.Runner) *KubeClient {
	return &KubeClient{
		runner: cliRunner,
	}
}

var _ Kube = &KubeClient{}

func (k *KubeClient) Apply(manifests []byte, namespace string) ([]byte, error) {
	args := []string{
		"apply",
		"--namespace", namespace,
		"-f", "-",
	}

	return k.runKubectlCmdWithInput(args, manifests)
}

func (k *KubeClient) Delete(manifests []byte, namespace string) ([]byte, error) {
	args := []string{
		"delete",
		"--namespace", namespace,
		"-f", "-",
	}

	return k.runKubectlCmdWithInput(args, manifests)
}

func (k *KubeClient) DeleteByName(name, kind, namespace string) ([]byte, error) {
	args := []string{
		"delete",
		"--namespace", namespace,
		kind, name,
	}

	return k.runKubectlCmd(args)
}

func (k *KubeClient) GetClusterName(ctx context.Context) (string, error) {
	args := []string{
		"config", "current-context",
	}

	out, err := k.runKubectlCmd(args)
	if err != nil {
		return "", errors.Wrap(err, "failed to get kubectl current-context")
	}

	clusterName := sanitize(string(bytes.TrimSuffix(out, []byte("\n"))))
	return clusterName, nil
}

func sanitize(name string) string {
	reRemoveUnAllowed := regexp.MustCompile(`[^a-z0-9\s-]+`)
	reNoDupDashes := regexp.MustCompile(`^--+`)
	reNoOutsideDashes := regexp.MustCompile(`^-+|-$`)

	replaceUnderscores := strings.ReplaceAll(strings.ToLower(name), "_", "-")
	notAllowed := reRemoveUnAllowed.ReplaceAllString(replaceUnderscores, "")
	noDupDashes := reNoDupDashes.ReplaceAllString(notAllowed, "")
	return reNoOutsideDashes.ReplaceAllString(noDupDashes, "")
}

func (k *KubeClient) GetClusterStatus(ctx context.Context) ClusterStatus {
	// Checking wego presence
	if _, err := k.runKubectlCmd([]string{"get", "crd", "apps.wego.weave.works"}); err == nil {
		return WeGOInstalled
	}

	// Checking flux presence
	if _, err := k.runKubectlCmd([]string{"get", "namespace", "flux-system"}); err == nil {
		return FluxInstalled
	}

	hostPortError := "was refused - did you specify the right host or port?"
	if out, err := k.runKubectlCmd([]string{"get", "deployment", "coredns", "-n", "kube-system"}); err == nil ||
		!strings.Contains(string(out), hostPortError) {
		return Unmodified
	}

	return Unknown
}

// FluxPresent checks flux presence in the cluster
func (k *KubeClient) FluxPresent(ctx context.Context) (bool, error) {
	out, err := k.runKubectlCmd([]string{"get", "namespace", "flux-system"})
	if err != nil {
		if strings.Contains(string(out), "not found") {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// SecretPresent checks for a specific secret within a specified namespace
func (k *KubeClient) SecretPresent(ctx context.Context, secretName, namespace string) (bool, error) {
	out, err := k.runKubectlCmd([]string{"get", "secret", secretName, "-n", namespace})
	if err != nil {
		if strings.Contains(string(out), "not found") {
			return false, nil
		}
	}

	return true, nil
}

func (k *KubeClient) GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
	cmd := []string{"get", "app", name.Name, "-n", name.Namespace, "-o", "json"}
	o, err := k.runKubectlCmd(cmd)

	if err != nil {
		return nil, fmt.Errorf("could not run kubectl command: %s", err)
	}

	a := wego.Application{}

	if err := json.Unmarshal(o, &a); err != nil {
		return nil, fmt.Errorf("could not unmarshal json: %s", err)
	}

	return &a, nil
}

func (k *KubeClient) GetApplications(ctx context.Context, ns string) ([]wego.Application, error) {
	cmd := []string{"get", "apps", "-n", ns, "-o", "json"}
	output, err := k.runKubectlCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("could not get applications: %s", err)
	}

	a := wego.ApplicationList{}
	if err := json.Unmarshal(output, &a); err != nil {
		return nil, fmt.Errorf("could not unmarshal applications json: %s", err)
	}

	return a.Items, nil
}

func (k *KubeClient) GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error {
	return errors.New("method not implemented, use the go-client implementation of the kube interface")
}

func (k *KubeClient) GetSecret(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error) {
	cmd := []string{"get", "secret", name.Name, "-n", name.Namespace, "-o", "json"}

	output, err := k.runKubectlCmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("could not get secret %w", err)
	}

	s := &corev1.Secret{}
	if err := json.Unmarshal(output, s); err != nil {
		return nil, fmt.Errorf("error unmarshalling secret: %w", err)
	}

	return s, nil

}

func (k *KubeClient) runKubectlCmd(args []string) ([]byte, error) {
	out, err := k.runner.Run(kubectlPath, args...)
	if err != nil {
		return out, fmt.Errorf("failed to run kubectl with output: %s", string(out))
	}

	return out, nil
}

func (k *KubeClient) runKubectlCmdWithInput(args []string, input []byte) ([]byte, error) {
	out, err := k.runner.RunWithStdin(kubectlPath, args, input)
	if err != nil {
		return out, fmt.Errorf("failed to run kubectl with output: %s", string(out))
	}

	return out, nil
}
