package cmdimpl

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

//go:embed manifests/app-crd.yaml
var appCRD []byte

type InstallParamSet struct {
	Namespace string
	DryRun    bool
}

func Install(params InstallParamSet) error {
	present, err := checkFluxPresent()
	if err != nil {
		return wrapError(err, "could not verify flux presence in the cluster")
	}

	if present {
		fmt.Println("We've verified that a flux-system namespace is present indicating you probably has flux installed in your cluster. \nCurrently we don't support running wego and flux side by side.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
		shims.Exit(1)
	}

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

func checkFluxPresent() (bool, error) {
	out, err := utils.CallCommandSilently("kubectl get namespace flux-system")
	if err != nil {
		if strings.Contains(string(out), "not found") {
			return false, nil
		}
	}

	return true, nil
}
