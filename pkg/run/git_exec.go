package run

import (
	"os/exec"
	"strings"
)

func GetBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func GetCommitID() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()

	if err != nil {
		return "", err
	}

	commitID := strings.TrimSpace(string(out))
	if len(commitID) > 8 {
		commitID = commitID[:8]
	}

	return commitID, nil
}

func IsDirty() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()

	if err != nil {
		return false, err
	}

	return len(out) > 0, nil
}
