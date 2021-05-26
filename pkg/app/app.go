package app

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func NewApp(name, path, url string) App {

	app := App{
		ApiVersion: "wego.weave.works/v1alpha1",
		Kind:       "Application",
		Metadata: Metadata{
			Name: name,
		},
		Spec: Spec{
			path,
			url,
		},
	}

	return app
}

// App yaml App definition
type App struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

func (a *App) Bytes() ([]byte, error) {

	bts, err := yaml.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("error on yaml marshall %s", err)
	}

	return bts, nil
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}
