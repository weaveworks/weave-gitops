package run

import (
	"errors"
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

func findGitRepoDir() (string, error) {
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
