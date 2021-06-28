package manifests

import (
	_ "embed"
)

//go:embed crds/wego.weave.works_applications.yaml
var AppCRD []byte
