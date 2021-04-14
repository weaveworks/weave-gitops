package flux

import (
	"os/exec"
	"strings"
)

// GetLast log for each flux namespace
func GetLatestStatusAllNamespaces() ([]string, error) {
	exePath, err := GetFluxExePath()
	if err != nil {
		return nil, err
	}

	c := exec.Command(exePath, "logs --level=error --all-namespaces")
	logs, err := c.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return getLastLogForNamespaces(logs)
}

// Example log "2021-04-12T19:54:02.588Z info HelmChart - Starting workers"
// function gets last entry for above example HelmChart
func getLastLogForNamespaces(logs []byte) ([]string, error) {
	logsArray := strings.Split(string(logs), "\n")
	namespaces := make(map[string]string)
	for _, line := range logsArray {
		splitLine := strings.Split(string(line), " ")
		namespaces[splitLine[2]] = line
	}

	result := []string{}
	for _, log := range namespaces {
		result = append(result, log)
	}

	return result, nil
}
