package cmdimpl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// Status provides the implementation for the wego status application command
func Status(args []string, allParams AddParamSet) error {

	allParams.Name = args[0]

	// verify the app is in the wego apps folder
	appPath, err := utils.GetWegoAppPath(allParams.Name)
	if err != nil {
		return fmt.Errorf("error getting path for app [%s] \n", args[0])

	}

	if !utils.Exists(appPath) && allParams.Name != "wego" {
		return fmt.Errorf("app provided does not exists on apps folder [%s]\n", args[0])
	}

	deploymentType, err := getDeploymentType(allParams.Namespace, allParams.Name)
	if err != nil {
		return fmt.Errorf("error getting deployment type [%s]", err)
	}

	err = printOutLatestSuccessfulDeploymentType(allParams.Namespace, allParams.Name, deploymentType)
	if err != nil {
		return fmt.Errorf("error on latest deployment time [%s]", err)
	}

	cmd := fmt.Sprintf(`get all -A %s`,
		allParams.Name)
	_, err = fluxops.CallFlux(cmd)
	return err
}

type Yaml struct {
	Status struct {
		Conditions []struct {
			LastTransitionTime string `yaml:"lastTransitionTime"`
		} `yaml:"conditions"`
	} `yaml:"status"`
}

func printOutLatestSuccessfulDeploymentType(namespace, appName string, deploymentType DeploymentType) error {

	c := fmt.Sprintf(`kubectl \
			-n %s \
			get %s/%s -oyaml`,
		namespace,
		deploymentType,
		appName,
	)

	stdout, stderr, err := utils.CallCommandSeparatingOutputStreams(c)
	if err != nil {
		return fmt.Errorf("error getting resource info [%s %s]\n", err.Error(), string(stderr))
	}
	var yamlOutput Yaml
	err = yaml.Unmarshal(stdout, &yamlOutput)
	if err != nil {
		return fmt.Errorf("error unmarshalling yaml output [%s] \n", err.Error())
	}
	if len(yamlOutput.Status.Conditions) == 0 {
		return fmt.Errorf("error getting latest deployment time [%s] \n", stdout)
	}
	fmt.Println("Latest successful deployment time: ", yamlOutput.Status.Conditions[0].LastTransitionTime)

	return nil
}

func getDeploymentType(namespace, appName string) (DeploymentType, error) {

	c := fmt.Sprintf(`flux get all -n %s`,
		namespace,
	)

	stdout, stderr, err := utils.CallCommandSeparatingOutputStreams(c)
	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		return "", err
	}

	if len(stderr) != 0 {
		return "", fmt.Errorf("%s", stderr)
	}

	var re = regexp.MustCompile(fmt.Sprintf(`(?m)(kustomization|helmrelease)\/%s`, appName))

	matches := re.FindAllStringSubmatch(string(stdout), -1)

	if len(matches) != 1 {
		return "", fmt.Errorf("error trying to get the deployment type of the app. raw output => %s", stdout)
	}

	if len(matches[0]) != 2 {
		return "", fmt.Errorf("error trying to get the deployment type of the app. raw output => %s", matches)
	}

	return DeploymentType(matches[0][1]), nil

}
