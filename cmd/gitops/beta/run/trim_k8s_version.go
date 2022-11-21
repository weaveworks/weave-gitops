package run

import (
	"regexp"
	"strings"
)

func trimK8sVersion(version string) string {
	// match v1.18.0 with regex
	regex := regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
	firstPart := regex.FindString(version)

	return strings.TrimPrefix(firstPart, "v")
}
