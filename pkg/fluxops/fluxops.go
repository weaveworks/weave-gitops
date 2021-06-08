package fluxops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
	"sigs.k8s.io/yaml"
)

var (
	fluxHandler interface{} = DefaultFluxHandler{}
	fluxBinary  string
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . FluxHandler
type FluxHandler interface {
	Handle(args string) ([]byte, error)
}

type DefaultFluxHandler struct{}

func (h DefaultFluxHandler) Handle(arglist string) ([]byte, error) {
	initFluxBinary()
	return utils.CallCommand(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

type quietFluxHandler struct{}

func (q quietFluxHandler) Handle(arglist string) ([]byte, error) {
	initFluxBinary()
	return utils.CallCommandSilently(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

func Override(handler FluxHandler) override.Override {
	return override.Override{Handler: &fluxHandler, Mock: handler, Original: fluxHandler}
}

// WithFluxHandler allows running a function with a different flux handler in force
func WithFluxHandler(handler FluxHandler, f func() ([]byte, error)) ([]byte, error) {
	switch fluxHandler.(type) {
	case DefaultFluxHandler:
		existingHandler := fluxHandler
		fluxHandler = handler
		defer func() {
			fluxHandler = existingHandler
		}()
		return f()
	default:
		return f()
	}
}

func FluxPath() (string, error) {
	homeDir, err := shims.UserHomeDir()
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
	return fluxHandler.(FluxHandler).Handle(strings.Join(arglist, " "))
}

func Install(namespace string) ([]byte, error) {
	return installFlux(namespace, true)
}

func QuietInstall(namespace string) ([]byte, error) {
	return installFlux(namespace, false)
}

func installFlux(namespace string, verbose bool) ([]byte, error) {
	args := []string{
		"install",
		fmt.Sprintf("--namespace=%s", namespace),
		"--components-extra=image-reflector-controller,image-automation-controller",
	}

	manifests, err := CallFlux(args...)
	if err != nil {
		return nil, err
	}
	return manifests, nil
}

func GetAllResourcesStatus(appName string) ([]byte, error) {
	args := []string{
		"get",
		"all",
		"-A",
		appName,
	}

	return WithFluxHandler(quietFluxHandler{}, func() ([]byte, error) {
		output, err := CallFlux(args...)
		if err != nil {
			return nil, err
		}
		return output, nil
	})
}

func GetAllResources(namespace string) ([]byte, error) {
	args := []string{
		//get all -n
		"get",
		"all",
		"-n",
		namespace,
	}

	return WithFluxHandler(quietFluxHandler{}, func() ([]byte, error) {
		output, err := CallFlux(args...)
		if err != nil {
			return nil, err
		}
		return output, nil
	})
}

// GetOwnerFromEnv determines the owner of a new repository based on the GITHUB_ORG
func GetOwnerFromEnv() (string, error) {
	// check for github username
	user, okUser := os.LookupEnv("GITHUB_ORG")
	if okUser {
		return user, nil
	}

	return GetUserFromHubCredentials()
}

func GetUserFromHubCredentials() (string, error) {
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
			shims.Exit(1)
		}
		fluxBinary = fluxPath
	}
}
