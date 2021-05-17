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
	apps    []App
	appName string
}

func NewAppManager(appName string) AppManager {
	return AppManager{
		appName: appName,
	}
}

func (a *AppManager) getApps() error {

	apps := make([]App, 0)

	appsPath, err := utils.GetAppsPath(a.appName)
	if err != nil {
		return err
	}
	if !utils.Exists(appsPath) {
		if err := os.MkdirAll(appsPath, 0755); err != nil {
			return err
		}
	}

	yamlPath := filepath.Join(appsPath, "app.yaml")

	if err = decodeYAMLFileToStruct(yamlPath, &apps); err != nil {
		return err
	}

	a.apps = apps

	return nil

}

func GetAppsYamlPath(appName string) (string, error) {
	appsPath, err := utils.GetAppsPath(appName)
	if err != nil {
		return "", err
	}

	appYamlPath := filepath.Join(appsPath, "app.yaml")

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

	yamlPath, err := GetAppsYamlPath(a.appName)
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

	apps := make([]App, 0)

	newAppExistsAlready := false
	for _, currentApp := range a.apps {
		if currentApp.Metadata.Name == newApp.Metadata.Name {
			newAppExistsAlready = true
			apps = append(apps, newApp)
		} else {
			apps = append(apps, currentApp)
		}
	}

	if !newAppExistsAlready {
		apps = append(apps, newApp)
	}

	a.apps = apps

	return a.persistApps()

}
