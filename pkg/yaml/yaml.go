package yaml

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
)

const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`

type AppManager struct {
	apps []App
	keys map[string]interface{}
}

func (a *AppManager) getApps() (*[]App, error) {

	apps := make([]App, 0)

	wegoAppsPath, err := utils.GetWegoAppsPath()
	if err != nil {
		return nil, err
	}

	appYamlPath := filepath.Join(wegoAppsPath, "app.yaml")
	if utils.Exists(appYamlPath) {
		appsReader, err := os.Open(appYamlPath)
		if err != nil {
			return nil, err
		}
		if err = yaml.NewEncoder(appsReader).Encode(&apps); err != nil {
			return nil, err
		}
	}

	return &apps, nil

}

func FromParamSetToApp(params cmdimpl.AddParamSet) App {
	return App{
		AppName: params.Name,
		AppPath: params.Path,
		AppURL:  params.Url,
	}
}

func (a *AppManager) AppendApp(app App) error {

	//apps, err := a.getApps()
	//if err != nil {
	//	return err
	//}
	//
	//return App{}

	return nil
}

type App struct {
	AppName string
	AppPath string
	AppURL  string
}

func AppendWegoApp(params cmdimpl.AddParamSet) error {

	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	if err != nil {
		return err
	}

	var populated bytes.Buffer
	if err := t.Execute(&populated, App{
		AppName: params.Name,
		AppPath: params.Path,
		AppURL:  params.Url,
	}); err != nil {
		return err
	}

	// does file exist
	// if so read the file
	// if not create it
	// append the new app to the struct
	// write new array apps value
	//

	//return ioutil.WriteFile(appYamlName, populated.Bytes(), 0644)

	return nil
}
