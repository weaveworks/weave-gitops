package yaml

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

//const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
//kind: Application
//metadata:
//  name: {{ .AppName }}
//spec:
//  path: {{ .AppPath }}
//  url: {{ .AppURL }}
//`

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
		if currentApp.AppName == newApp.AppName {
			newApps = append(newApps, newApp)
		} else {
			newApps = append(newApps, currentApp)
		}
	}

	a.apps = newApps

	return a.persistApps()

}

type App struct {
	AppName string
	AppPath string
	AppURL  string
}
