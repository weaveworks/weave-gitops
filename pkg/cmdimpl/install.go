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
}

func Install(params InstallParamSet) ([]byte, error) {
	kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)

	if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(appCRD)); err != nil {
		return []byte(""), wrapError(err, "could not apply wego source")
	}

	manifests, err := fluxops.Install(params.Namespace)
	if err != nil {
		return nil, err
	}

	return append(manifests, appCRD...), nil
}
