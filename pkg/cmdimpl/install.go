package cmdimpl

import (
	_ "embed"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
)

//go:embed manifests/app-crd.yaml
var appCRD []byte

type InstallParamSet struct {
	Namespace string
}

func Install(params InstallParamSet) ([]byte, error) {
	manifests, err := fluxops.QuietInstall(params.Namespace)
	if err != nil {
		return nil, err
	}

	return append(manifests, appCRD...), nil
}
