package cmdimpl

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// Status provides the implementation for the wego status application command
func Status(allParams AddParamSet) error {
	// verify the app is in the wego apps folder
	fluxRepoName, err := fluxops.GetRepoName()
	if err != nil {
		return wrapError(err, "could not get repo name")
	}
	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	fluxRepo := filepath.Join(reposDir, fluxRepoName)
	appPath := filepath.Join(fluxRepo, "apps", allParams.Name)
	if err != nil {
		return fmt.Errorf("error getting path for app [%s], err: %s \n", allParams.Name, err)
	}

	if !utils.Exists(appPath) && allParams.Name != "wego" {
		return fmt.Errorf("app provided does not exist in apps folder [%s]\n", appPath)
	}

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
		return "", fmt.Errorf("error getting resource info [%s %s]\n", err.Error(), string(stderr))
	}
	var yamlOutput Yaml
	err = yaml.Unmarshal(stdout, &yamlOutput)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling yaml output [%s] \n", err.Error())
	}
	if len(yamlOutput.Status.Conditions) == 0 {
		return "", fmt.Errorf("error getting latest deployment time [%s] \n", stdout)
	}

	return yamlOutput.Status.Conditions[0].LastTransitionTime, nil
}

func getDeploymentType(namespace, appName string) (DeploymentType, error) {

	stdout, err := fluxops.GetAllResources(namespace)
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return "", err
	}

	var re = regexp.MustCompile(fmt.Sprintf(`(?m)(kustomization|helmrelease)\/%s`, appName))

	matches := re.FindAllStringSubmatch(string(stdout), -1)

	if len(matches) != 1 {
		return "", fmt.Errorf("error trying to get the deployment type of the app. raw output => %s", stdout)
	}

	if len(matches[0]) != 2 {
		return "", fmt.Errorf("error trying to get the deployment type of the app. raw matches => %s", matches)
	}

	return DeploymentType(matches[0][1]), nil

}
