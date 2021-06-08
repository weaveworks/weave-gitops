package cmdimpl

import (
	_ "embed"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

//go:embed manifests/app-crd.yaml
var appCRD []byte

type InstallParamSet struct {
	Namespace string
	DryRun    bool
}

func Install(params InstallParamSet) error {
	manifests, err := fluxops.Install(params.Namespace, params.DryRun)
	if err != nil {
		return fmt.Errorf("error on install %s", err)
	}

	if params.DryRun {
		fmt.Print(string(manifests))
		fmt.Println(string(appCRD))
	} else {
		kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)
		if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(appCRD)); err != nil {
			return wrapError(err, "could not apply wego manifests")
		}
	}

	return nil
}
