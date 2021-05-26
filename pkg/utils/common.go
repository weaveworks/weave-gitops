package utils

import (
	"fmt"
	"os"
	"path/filepath"
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

// GetContextName returns the current context name
func GetContextName() (string, error) {
	c := "kubectl config current-context"
	currentCluster, stderr, err := CallCommandSeparatingOutputStreams(c)
	if err != nil {
		return "", fmt.Errorf("error getting resource info [%s %s]\n", err.Error(), string(stderr))
	}
	return string(currentCluster), nil
}

// GetWegoRepoName returns the name of the wego repo for the cluster (the repo holding controller defs)
func GetWegoRepoName() (string, error) {
	clusterName, err := GetContextName()
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
