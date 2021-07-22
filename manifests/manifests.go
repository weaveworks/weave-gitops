package manifests

import (
	_ "embed"
)

//go:embed crds/wego.weave.works_apps.yaml
var AppCRD []byte

//go:embed api-service/service-account.yaml
var ServiceAccountApiService []byte

//go:embed api-service/role.yaml
var RoleApiService []byte

//go:embed api-service/rolebinding.yaml
var RoleBindingApiService []byte
