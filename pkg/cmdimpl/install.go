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
}

func Install(params InstallParamSet) ([]byte, error) {
	present, err := checkFluxPresent()
	if err != nil {
		return []byte(""), wrapError(err, "could not verify flux presence in the cluster")
	}

	if present {
		fmt.Println("We've verified that a flux-system namespace is present indicating you probably has flux installed in your cluster. \nCurrently we don't support running wego and flux side by side.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
		shims.Exit(1)
	}

	kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)

	if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(appCRD)); err != nil {
		return []byte(""), wrapError(err, "could not apply wego manifests")
	}

	manifests, err := fluxops.Install(params.Namespace)
	if err != nil {
		return nil, err
	}

	return append(manifests, appCRD...), nil
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
