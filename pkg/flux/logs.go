package flux

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/utils"
)

// GetLast log for each flux namespace
func GetLatestStatusAllNamespaces() ([]string, error) {
	exePath, err := GetFluxExePath()
	if err != nil {
		return nil, err
	}

	logs, err := utils.CallCommand(fmt.Sprintf("%s logs --all-namespaces", exePath))
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
		if len(splitLine) < 3 {
			continue
		}
		namespaces[splitLine[2]] = line
	}

	result := []string{}
	for _, log := range namespaces {
		result = append(result, log)
	}

	return result, nil
}
