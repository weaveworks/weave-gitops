package cmdimpl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type DeploymentType string
type SourceType string
type ConfigType string

type StatusParams struct {
	Namespace string
	Name      string
}

// Status provides the implementation for the wego status application command
func Status(allParams StatusParams) error {
	deploymentType, err := getDeploymentType(allParams.Namespace, allParams.Name)
	if err != nil {
		return fmt.Errorf("error getting deployment type [%s]", err)
	}

	latestDeploymentTime, err := getLatestSuccessfulDeploymentTime(allParams.Namespace, allParams.Name, deploymentType)
	if err != nil {
		return fmt.Errorf("error on latest deployment time [%s]", err)
	}
	fmt.Printf("Latest successful deployment time: %s\n", latestDeploymentTime)

	output, err := fluxops.GetAllResourcesStatus(allParams.Name)
	if err != nil {
		return fmt.Errorf("error getting flux app resources status [%s", err)
	}

	fmt.Println(string(output))

	return err
}

type Yaml struct {
	Status struct {
		Conditions []struct {
			LastTransitionTime string `yaml:"lastTransitionTime"`
		} `yaml:"conditions"`
	} `yaml:"status"`
}

func getLatestSuccessfulDeploymentTime(namespace, appName string, deploymentType DeploymentType) (string, error) {
	c := fmt.Sprintf(`kubectl \
            -n %s \
            get %s/%s -oyaml`,
		namespace,
		deploymentType,
		appName,
	)

	stdout, stderr, err := utils.CallCommandSeparatingOutputStreams(c)
	if err != nil {
		return "", fmt.Errorf("error getting resource info [%s %s]", err.Error(), string(stderr))
	}
	var yamlOutput Yaml
	err = yaml.Unmarshal(stdout, &yamlOutput)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling yaml output [%s]", err.Error())
	}
	if len(yamlOutput.Status.Conditions) == 0 {
		return "", fmt.Errorf("error getting latest deployment time [%s]", stdout)
	}

	return yamlOutput.Status.Conditions[0].LastTransitionTime, nil
}

func getDeploymentType(namespace, appName string) (DeploymentType, error) {
	stdout, err := fluxops.GetAllResources(namespace)
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return "", err
	}

	var re = regexp.MustCompile(fmt.Sprintf(`(?m)(kustomization|helmrelease)\/%s[[:space:]]`, appName))

	matches := re.FindAllStringSubmatch(string(stdout), -1)

	if len(matches) != 1 {
		return "", fmt.Errorf("error trying to get the deployment type of the app. raw output => %s", stdout)
	}

	return DeploymentType(matches[0][1]), nil
}
