package manifests

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/pkg/errors"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

var Manifests [][]byte

//go:embed crds/wego.weave.works_apps.yaml
var AppCRD []byte

//go:embed wego-app/deployment.yaml
var WegoAppDeployment []byte

type deploymentParameters struct {
	Version string
}

var errInjectingValuesToTemplate = errors.New("error injecting values to template")

// GenerateWegoAppDeploymentManifest generates wego-app deployment manifest from a template
func GenerateWegoAppDeploymentManifest(deploymentTemplate []byte) ([]byte, error) {
	deploymentValues := deploymentParameters{version.Version}

	template := template.New("DeploymentTemplate")

	var err error

	template, err = template.Parse(string(deploymentTemplate))
	if err != nil {
		return nil, fmt.Errorf("error parsing template %w", err)
	}

	deploymentYaml := &bytes.Buffer{}

	err = template.Execute(deploymentYaml, deploymentValues)
	if err != nil {
		return nil, fmt.Errorf("%s %w", errInjectingValuesToTemplate, err)
	}

	return deploymentYaml.Bytes(), nil
}

//go:embed wego-app/service-account.yaml
var WegoAppServiceAccount []byte

//go:embed wego-app/service.yaml
var WegoAppService []byte

//go:embed wego-app/role.yaml
var WegoAppRole []byte

//go:embed wego-app/role-binding.yaml
var WegoAppRoleBinding []byte

func init() {
	Manifests = [][]byte{
		AppCRD,
		WegoAppServiceAccount,
		WegoAppRoleBinding,
		WegoAppRole,
		WegoAppService,
	}
}
