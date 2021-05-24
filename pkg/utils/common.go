package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func GetWegoLocalPath() (string, error) {

	wegoRepoName, err := GetWegoRepoName()
	if err != nil {
		return "", err
	}

	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	wegoRepoPath := filepath.Join(reposDir, wegoRepoName)
	return wegoRepoPath, os.MkdirAll(wegoRepoPath, 0755)
}

func GetWegoAppsPath() (string, error) {

	LocalWegoPath, err := GetWegoLocalPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(LocalWegoPath, "apps"), nil
}

func GetWegoAppPath(app string) (string, error) {

	localWegoAppsPath, err := GetWegoAppsPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(localWegoAppsPath, app), nil
}

func GetAppsPath(appName string) (string, error) {
	wegoAppsPath, err := GetWegoAppsPath()
	if err != nil {
		return "", err
	}

	appYamlPath := filepath.Join(wegoAppsPath, appName)

	return appYamlPath, nil
}

// GetClusterName returns the cluster name associated with the current context in ~/.kube/config
func GetClusterName() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	config, err := ioutil.ReadFile(filepath.Join(homeDir, ".kube", "config"))
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		return "", err
	}
	return data["current-context"].(string), nil
}

// GetWegoRepoName returns the name of the wego repo for the cluster (the repo holding controller defs)
func GetWegoRepoName() (string, error) {
	clusterName, err := GetClusterName()
	if err != nil {
		return "", err
	}
	return clusterName + "-wego", nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
