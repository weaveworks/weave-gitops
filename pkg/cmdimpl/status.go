package cmdimpl

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// Status provides the implementation for the wego status application command
func Status(args []string, allParams AddParamSet) {

	// verify the app is in the wego apps folder
	appPath, err := utils.GetWegoApp(args[0])
	if err != nil {
		fmt.Printf("error getting path for app [%s] \n", args[0])
		os.Exit(1)
	}
	if !utils.Exists(appPath) {
		fmt.Printf("app provided does not exists on apps folder [%s]\n", args[0])
		os.Exit(1)
	}
	params.Name = args[0]

	// get the

}
