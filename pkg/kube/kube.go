package kube

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/pkg/errors"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

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
	Unknown:       "Unknown",
	Unmodified:    "Unmodified",
	FluxInstalled: "FluxInstalled",
	WeGOInstalled: "WeGOInstalled",
}

//counterfeiter:generate . Kube
type Kube interface {
	Apply(manifests []byte, namespace string) ([]byte, error)
	Delete(manifests []byte, namespace string) ([]byte, error)
	SecretPresent(ctx context.Context, string, namespace string) (bool, error)
	GetApplications(ctx context.Context, namespace string) ([]wego.Application, error)
	FluxPresent(ctx context.Context) (bool, error)
	GetClusterName(ctx context.Context) (string, error)
	GetClusterStatus(ctx context.Context) ClusterStatus
	GetApplication(ctx context.Context, name string) (*wego.Application, error)
	LabelExistsInCluster(ctx context.Context, label string) error
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

	out, err := k.runKubectlCmdWithInput(args, manifests)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (k *KubeClient) Delete(manifests []byte, namespace string) ([]byte, error) {
	args := []string{
		"delete",
		"--namespace", namespace,
		"-f", "-",
	}

	out, err := k.runKubectlCmdWithInput(args, manifests)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (k *KubeClient) GetClusterName(ctx context.Context) (string, error) {
	args := []string{
		"config", "current-context",
	}

	out, err := k.runKubectlCmd(args)
	if err != nil {
		return "", errors.Wrap(err, "failed to get kubectl current-context")
	}

	return string(bytes.TrimSuffix(out, []byte("\n"))), nil
}

func (k *KubeClient) GetClusterStatus(ctx context.Context) ClusterStatus {
	// Checking wego presence
	if k.resourceLookup("get crd apps.wego.weave.works") == nil {
		return WeGOInstalled
	}

	// Checking flux presence
	if k.resourceLookup("get namespace flux-system") == nil {
		return FluxInstalled
	}

	if k.resourceLookup("get deployment coredns -n kube-system") == nil {
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

func (k *KubeClient) GetApplication(ctx context.Context, name string) (*wego.Application, error) {
	cmd := []string{"get", "app", name, "-o", "json"}
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

func (k *KubeClient) resourceLookup(args string) error {
	_, err := k.runKubectlCmd(strings.Split(args, " "))
	if err != nil {
		return err
	}

	return nil
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
func (k *KubeClient) LabelExistsInCluster(ctx context.Context, label string) error {
	cmd := []string{"get", "app", "-l", fmt.Sprintf("weave-gitops.weave.works/app-identifier=%s", label)}
	o, err := k.runKubectlCmd(cmd)
	if err != nil {
		return fmt.Errorf("could not run kubectl command: %s", err)
	}
	if !strings.Contains(string(o), "No resources found") {
		return fmt.Errorf("unable to create resource, resource already exists in cluster")
	}
	return nil
}
