package yaml

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// App yaml App definition
type App struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

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

// Metadata
type Metadata struct {
	Name string `yaml:"name"`
}

// Spec
type Spec struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

type AppManager struct {
	apps []App
}

func (a *AppManager) getApps() error {

	apps := make([]App, 0)

	yamlPath, err := GetAppsYamlPath()
	if err != nil {
		return err
	}

	if err = decodeYAMLFileToStruct(yamlPath, &apps); err != nil {
		return err
	}

	a.apps = apps

	return nil

}

func GetAppsYamlPath() (string, error) {
	wegoAppsPath, err := utils.GetWegoAppsPath()
	if err != nil {
		return "", err
	}

	appYamlPath := filepath.Join(wegoAppsPath, "app.yaml")

	return appYamlPath, nil
}

func decodeYAMLFileToStruct(yamlFilePath string, apps *[]App) error {
	if utils.Exists(yamlFilePath) {
		appsReader, err := os.Open(yamlFilePath)
		if err != nil {
			return err
		}
		if err = yaml.NewEncoder(appsReader).Encode(&apps); err != nil {
			return err
		}
	}

	return nil
}

func (a *AppManager) persistApps() error {

	bts, err := yaml.Marshal(a.apps)
	if err != nil {
		return err
	}

	yamlPath, err := GetAppsYamlPath()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(yamlPath, bts, 0644)

}

func (a *AppManager) AddApp(newApp App) error {

	err := a.getApps()
	if err != nil {
		return err
	}

	newApps := make([]App, 0)

	for _, currentApp := range a.apps {
		if currentApp.Metadata.Name == newApp.Metadata.Name {
			newApps = append(newApps, newApp)
		} else {
			newApps = append(newApps, currentApp)
		}
	}

	a.apps = newApps

	return a.persistApps()

}
