package manifests

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"text/template"
)

const (
	wegoManifestsDir     = "wego-app"
	profilesManifestsDir = "profiles-server"
)

var (
	//go:embed crds/wego.weave.works_apps.yaml
	AppCRD []byte
	//go:embed wego-app/*
	wegoAppTemplates embed.FS
	//go:embed profiles-server/*
	profilesServerTemplates embed.FS
)

type Params struct {
	AppVersion      string
	ProfilesVersion string
	Namespace       string
}

// GenerateManifests generates weave-gitops manifests from a template
func GenerateManifests(params Params) ([][]byte, error) {
	manifests := [][]byte{}

	appManifests, err := readTemplateDirectory(params, wegoAppTemplates, wegoManifestsDir)
	if err != nil {
		return nil, err
	}

	manifests = append(manifests, appManifests...)

	profilesManifests, err := readTemplateDirectory(params, profilesServerTemplates, profilesManifestsDir)
	if err != nil {
		return nil, err
	}

	manifests = append(manifests, profilesManifests...)

	return manifests, nil
}

func executeTemplate(name string, tplData string, params Params) ([]byte, error) {
	template, err := template.New(name).Parse(tplData)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %s: %w", name, err)
	}

	yaml := &bytes.Buffer{}

	err = template.Execute(yaml, params)
	if err != nil {
		return nil, fmt.Errorf("error injecting values to template: %w", err)
	}

	return yaml.Bytes(), nil
}

func readTemplateDirectory(params Params, templateFiles embed.FS, templatestDir string) ([][]byte, error) {
	templates, err := fs.ReadDir(templateFiles, templatestDir)
	if err != nil {
		return nil, fmt.Errorf("failed reading templates directory: %w", err)
	}

	var manifests [][]byte

	for _, template := range templates {
		tplName := template.Name()

		data, err := fs.ReadFile(templateFiles, filepath.Join(templatestDir, tplName))
		if err != nil {
			return nil, fmt.Errorf("failed reading template %s: %w", tplName, err)
		}

		manifest, err := executeTemplate(tplName, string(data), params)
		if err != nil {
			return nil, fmt.Errorf("failed executing template: %s: %w", tplName, err)
		}

		manifests = append(manifests, manifest)
	}

	return manifests, nil
}
