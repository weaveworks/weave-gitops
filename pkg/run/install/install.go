package install

import (
	"errors"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

// FindGitRepoDir finds git repo root directory
func FindGitRepoDir() (string, error) {
	gitDir := "."

	for {
		if _, err := os.Stat(filepath.Join(gitDir, ".git")); err == nil {
			break
		}

		gitDir = filepath.Join(gitDir, "..")

		if gitDir == "/" {
			return "", errors.New("not in a git repo")
		}
	}

	return filepath.Abs(gitDir)
}

// GetRelativePathToRootDir gets relative path to a directory from the git root. It returns an error if there's no git repo.
func GetRelativePathToRootDir(rootDir string, path string) (string, error) {
	absGitDir, err := filepath.Abs(rootDir)

	if err != nil { // not in a git repo
		return "", err
	}

	return filepath.Rel(absGitDir, path)
}

// isPodStatusConditionPresentAndEqual returns true when conditionType is present and equal to status.
func isPodStatusConditionPresentAndEqual(conditions []corev1.PodCondition, conditionType corev1.PodConditionType, status corev1.ConditionStatus) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}

	return false
}
