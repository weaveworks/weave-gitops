package manifests

import (
	_ "embed"
)

//go:embed crds/app.yaml
var AppCRD []byte
