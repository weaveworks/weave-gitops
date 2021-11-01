package models

import (
	// "crypto/md5"
	// "fmt"
	// "path/filepath"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	// "github.com/weaveworks/weave-gitops/pkg/kube"
	// "k8s.io/apimachinery/pkg/runtime/schema"
)

type AutomationType string

type ConfigType string

type SourceType string

const (
	AutomationTypeHelm      AutomationType = "helm"
	AutomationTypeKustomize AutomationType = "kustomize"

	ConfigTypeUserRepo ConfigType = ""
	ConfigTypeNone     ConfigType = "NONE"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"
)

type Application struct {
	Name                string
	Namespace           string
	HelmSourceURL       string
	GitSourceURL        gitproviders.RepoURL
	ConfigURL           gitproviders.RepoURL
	Branch              string
	Path                string
	AutomationType      AutomationType
	SourceType          SourceType
	HelmTargetNamespace string
}

func IsExternalConfigUrl(url string) bool {
	return strings.ToUpper(url) != string(ConfigTypeNone) &&
		strings.ToUpper(url) != string(ConfigTypeUserRepo)
}
