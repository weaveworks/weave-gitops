package fluxops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
	"sigs.k8s.io/yaml"
)

var (
	fluxHandler FluxHandler = defaultFluxHandler{}
	fluxBinary  string
)

const fluxSystemNamespace = `apiVersion: v1
kind: Namespace
metadata:
  name: flux-system
`

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . FluxHandler
type FluxHandler interface {
	Handle(args string) ([]byte, error)
}

type defaultFluxHandler struct{}

func (h defaultFluxHandler) Handle(arglist string) ([]byte, error) {
	initFluxBinary()
	return utils.CallCommand(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

func FluxPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func SetFluxHandler(h FluxHandler) {
	fluxHandler = h
}

func CallFlux(arglist ...string) ([]byte, error) {
	return fluxHandler.Handle(strings.Join(arglist, " "))
}

func Install(namespace string) ([]byte, error) {
	args := []string{
		"install",
		fmt.Sprintf("--namespace=%s", namespace),
		"--export",
	}

	if namespace != "flux-system" {
		if err := utils.CallCommandForEffectWithInputPipe("kubectl apply -f -", fluxSystemNamespace); err != nil {
			return nil, err
		}
	}

	return CallFlux(args...)
}

// GetOwnerFromEnv determines the owner of a new repository based on the GITHUB_ORG
func GetOwnerFromEnv() (string, error) {
	// check for github username
	user, okUser := os.LookupEnv("GITHUB_ORG")
	if okUser {
		return user, nil
	}

	return getUserFromHubCredentials()
}

// GetRepoName returns the name of the wego repo for the cluster (the repo holding controller defs)
func GetRepoName() (string, error) {
	clusterName, err := status.GetClusterName()
	if err != nil {
		return "", err
	}
	return clusterName + "-wego", nil
}

func getUserFromHubCredentials() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// check for existing ~/.config/hub
	config, err := ioutil.ReadFile(filepath.Join(homeDir, ".config", "hub"))
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		return "", err
	}

	return data["github.com"].([]interface{})[0].(map[string]interface{})["user"].(string), nil
}

func initFluxBinary() {
	if fluxBinary == "" {
		fluxPath, err := FluxPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve wego executable path: %v", err)
			os.Exit(1)
		}
		fluxBinary = fluxPath
	}
}
