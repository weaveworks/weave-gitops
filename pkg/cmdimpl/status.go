package cmdimpl

import (
	"fmt"
	"os"
	"os/exec"

	fluxBin "github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/shims"

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

	// get the app status from flux
	exePath, err := fluxBin.GetFluxExePath()
	if err != nil {
		fmt.Fprintf(shims.Stderr(), "error getting flux path: %v\n", err)
		os.Exit(1)
	}

	c := exec.Command(exePath, "get", "all", "-A", params.Name)
	sourceOutput, err := c.CombinedOutput()
	if err != nil {
		fmt.Fprintf(shims.Stderr(), "error getting git source status: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(sourceOutput))

}
