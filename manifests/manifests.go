package manifests

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"text/template"
)

const (
	wegoManifestsDir = "wego-app"
)

var (
	//go:embed crds/wego.weave.works_apps.yaml
	AppCRD []byte
	//go:embed wego-app/*
	wegoAppTemplates embed.FS
)

type Params struct {
	AppVersion string
	Namespace  string
}

// GenerateManifests generates weave-gitops manifests from a template
func GenerateManifests(params Params) ([][]byte, error) {
	templates, err := fs.ReadDir(wegoAppTemplates, wegoManifestsDir)
	if err != nil {
		return nil, fmt.Errorf("failed reading templates directory: %w", err)
	}

	var manifests [][]byte

	for _, template := range templates {
		tplName := template.Name()

		data, err := fs.ReadFile(wegoAppTemplates, filepath.Join(wegoManifestsDir, tplName))
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
