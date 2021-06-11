package kube

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/runner"
)

type Kube interface {
	Apply(manifests []byte, namespace string) ([]byte, error)
}

type KubeClient struct {
	runner runner.Runner
}

func New() Kube {
	return &KubeClient{
		runner: &runner.CLIRunner{},
	}
}

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

func (k *KubeClient) runKubectlCmdWithInput(args []string, input []byte) ([]byte, error) {
	kubectlPath := "kubectl"

	out, err := k.runner.RunWithStdin(kubectlPath, args, input)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run kubectl with output: %s", string(out))
	}

	return out, nil
}
