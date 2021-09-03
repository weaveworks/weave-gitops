package manifests

import (
	"bytes"
	_ "embed"

	"github.com/weaveworks/weave-gitops/cmd/wego/version"
)

//go:embed crds/wego.weave.works_apps.yaml
var AppCRD []byte

//go:embed wego-app/deployment.yaml
var wegoAppDeployment []byte

func WegoAppDeployment() []byte {
	return bytes.ReplaceAll(wegoAppDeployment, []byte("VERSION"), []byte(version.Version))
}

//go:embed wego-app/service-account.yaml
var WegoAppServiceAccount []byte

//go:embed wego-app/service.yaml
var WegoAppService []byte

//go:embed wego-app/role.yaml
var WegoAppRole []byte

//go:embed wego-app/role-binding.yaml
var WegoAppRoleBinding []byte
