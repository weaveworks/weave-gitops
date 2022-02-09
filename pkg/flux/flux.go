package flux

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Flux
type Flux interface {
	GetVersion() (string, error)
	GetAllResourcesStatus(name string, namespace string) ([]byte, error)
	GetLatestStatusAllNamespaces() ([]string, error)
	PreCheck() (string, error)
}

const (
	PartOfLabelKey   = "app.kubernetes.io/part-of"
	PartOfLabelValue = "flux"
	VersionLabelKey  = "app.kubernetes.io/version"
)

type FluxClient struct {
	osys   osys.Osys
	runner runner.Runner
}

func New(osysClient osys.Osys, cliRunner runner.Runner) *FluxClient {
	return &FluxClient{
		osys:   osysClient,
		runner: cliRunner,
	}
}

var _ Flux = &FluxClient{}

func (f *FluxClient) GetAllResourcesStatus(name string, namespace string) ([]byte, error) {
	args := []string{
		"get", "all", "--namespace", namespace, name,
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to get flux resources status: %w", err)
	}

	return out, nil
}

func (f *FluxClient) GetVersion() (string, error) {
	out, err := f.runFluxCmd("-v")
	if err != nil {
		return "", err
	}
	// change string format to match our versioning standard
	version := strings.ReplaceAll(string(out), "flux version ", "v")

	return version, nil
}

func (f *FluxClient) runFluxCmd(args ...string) ([]byte, error) {
	fluxPath, err := f.fluxPath()
	if err != nil {
		return []byte{}, errors.Wrap(err, "error getting flux binary path")
	}

	out, err := f.runner.Run(fluxPath, args...)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run flux with output: %s and error: %w", string(out), err)
	}

	return out, nil
}

func (f *FluxClient) fluxPath() (string, error) {
	return "flux", nil
}

func (f *FluxClient) PreCheck() (string, error) {
	args := []string{
		"check",
		"--pre",
	}

	output, err := f.runFluxCmd(args...)
	if err != nil {
		return "", fmt.Errorf("failed running flux pre check %w", err)
	}

	return string(output), nil
}
