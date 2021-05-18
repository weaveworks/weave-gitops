package cmdimpl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// Status provides the implementation for the wego status application command
func Status(args []string, allParams AddParamSet) {

	// verify the app is in the wego apps folder
	appPath, err := utils.GetWegoAppPath(args[0])
	if err != nil {
		fmt.Printf("error getting path for app [%s] \n", args[0])
		os.Exit(1)
	}
	if !utils.Exists(appPath) && args[0] != "wego" {
		fmt.Printf("app provided does not exists on apps folder [%s]\n", args[0])
		os.Exit(1)
	}
	params.Name = args[0]

	// Get latest time
	c := fmt.Sprintf(`kubectl \
			-n %s \
			get kustomizations/testing-app -oyaml`,
		"wego-system",
	)

	stdout, stderr, err := utils.CallCommandSeparatingOutputStreams(c)
	if err != nil {
		fmt.Printf("error getting resource info [%s %s]\n", err.Error(), string(stderr))
		os.Exit(1)
	}
	var yamlOutput Yaml
	err = yaml.Unmarshal(stdout, &yamlOutput)
	if err != nil {
		fmt.Printf("error unmarshalling yaml output [%s] \n", err.Error())
		os.Exit(1)
	}
	if len(yamlOutput.Status.Conditions) == 0 {
		fmt.Printf("error getting latest deployment time [%s] \n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Latest successful deployment time: ", yamlOutput.Status.Conditions[0].LastTransitionTime)

	cmd := fmt.Sprintf(`get all -A %s`,
		params.Name)
	_, err = fluxops.CallFlux(cmd)
	checkAddError(err)
}

type Yaml struct {
	Status struct {
		Conditions []struct {
			LastTransitionTime string `yaml:"lastTransitionTime"`
		} `yaml:"conditions"`
	} `yaml:"status"`
}
