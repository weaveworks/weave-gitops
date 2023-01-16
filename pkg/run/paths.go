package run

import (
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	RootDir    string // Absolute path to the git repository
	CurrentDir string // Path to the current working directory, relative to rootDir
	ClusterDir string // Path to the cluster dir, relative to rootDir - currentDir / "clusters" / "my-cluster"
	TargetDir  string // Path to the target workload dir, relative to rootDir. Must not contain clusterDir, may or may not be inside currentDir
}

func (p *Paths) GetAbsoluteTargetDir() string {
	// this should have been sufficiently validated as to be un-failable
	targetDir, _ := filepath.Abs(filepath.Join(p.RootDir, p.TargetDir))
	return targetDir
}

/*
func (p *Paths) GetRelativeTargetDir() (string, error) {
	absGitDir, err := filepath.Abs(p.RootDir)

	if err != nil { // not in a git repo
		return "", err
	}

	return filepath.Rel(absGitDir, p.TargetDir)
}*/

const (
	NotInGitRepoError = "not in a git repo, last checked directory: %s"
	StatError         = "unexpected error while checking for .git directory: %v"
	AbsError          = "unexpected error while getting the absolute filepath: %v"
	PermissionError   = "permission denied while checking for parent directory of: %s"
)

func findGitRepoDir() (string, error) {
	gitDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf(AbsError, err)
	}

	for {
		_, err := os.Stat(filepath.Join(gitDir, ".git"))
		if err == nil {
			break
		} else if os.IsNotExist(err) {
			gitDir = filepath.Clean(filepath.Join(gitDir, ".."))
			absGitDir, err := filepath.Abs(gitDir)
			if err != nil {
				return "", fmt.Errorf(AbsError, err)
			}

			if filepath.Dir(absGitDir) == gitDir {
				return "", fmt.Errorf(NotInGitRepoError, gitDir)
			}

			gitDir = absGitDir
		} else if os.IsPermission(err) {
			return "", fmt.Errorf(PermissionError, gitDir)
		} else {
			return "", fmt.Errorf(StatError, err)
		}
	}

	return gitDir, nil
}

func GetRelativePathToRootDir(rootDir string, path string) (string, error) {
	absGitDir, err := filepath.Abs(rootDir)

	if err != nil { // not in a git repo
		return "", err
	}

	return filepath.Rel(absGitDir, path)
}

func NewPaths(specifiedTargetDir string, specifiedRootDir string) (*Paths, error) {
	paths := Paths{}

	gitRepoRoot, err := findGitRepoDir()
	if err != nil {
		return nil, err
	}

	rootDir := specifiedRootDir
	if rootDir == "" {
		rootDir = gitRepoRoot
	}

	// check if rootDir is valid
	if _, err := os.Stat(rootDir); err != nil {
		return nil, fmt.Errorf("root directory %s does not exist", paths.RootDir)
	}

	// find absolute path of the root Dir
	paths.RootDir, err = filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths.CurrentDir, err = GetRelativePathToRootDir(rootDir, currentDir)
	if err != nil {
		return nil, err
	}

	targetPath, err := filepath.Abs(filepath.Join(currentDir, specifiedTargetDir))
	if err != nil {
		return nil, err
	}

	paths.TargetDir, err = GetRelativePathToRootDir(rootDir, targetPath)
	if err != nil {
		return nil, err
	}

	paths.ClusterDir = filepath.Join(paths.CurrentDir, "clusters", "my-cluster")

	return &paths, nil
}
