package cmdimpl

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/kube"

	"k8s.io/apimachinery/pkg/types"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/runner"

	"github.com/weaveworks/weave-gitops/pkg/logger"

	"github.com/weaveworks/weave-gitops/pkg/services/app"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type DeploymentType string

const (
	DeploymentTypeHelmRelease   DeploymentType = "helmrelease"
	DeploymentTypeKustomization DeploymentType = "kustomization"
)

type SourceType string
type ConfigType string

type StatusParams struct {
	Namespace string
	Name      string
}

// Status provides the implementation for the wego status application command
func Status(allParams StatusParams) error {

	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(cliRunner)
	kubeClient := kube.New(cliRunner)

	appService := app.New(logger.New(os.Stdout), nil, fluxClient, kubeClient, nil)

	_, err := appService.Get(types.NamespacedName{Name: allParams.Name, Namespace: allParams.Namespace})
	if err != nil {
		return err
	}

	deploymentType, err := getDeploymentType(allParams.Namespace, allParams.Name)
	if err != nil {
		return err
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
	helmObjExists, err := fluxops.HelmReleaseExists(appName, namespace)
	if err != nil {
		return "", err
	}
	if helmObjExists {
		return DeploymentTypeHelmRelease, nil
	} else {
		return DeploymentTypeKustomization, nil
	}
}
