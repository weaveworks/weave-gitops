package models

type AutomationType string

const (
	AutomationTypeHelm      AutomationType = "helm"
	AutomationTypeKustomize AutomationType = "kustomize"
)

type SourceType string

const (
	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"
)

type Application struct {
	Name                string
	Namespace           string
	SourceURL           string
	ConfigURL           string
	Branch              string
	Path                string
	AutomationType      AutomationType
	SourceType          SourceType
	HelmTargetNamespace string
}
