package manifests

import (
	_ "embed"
)

//go:embed crds/wego.weave.works_apps.yaml
var AppCRD []byte
