package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/version"
)

// The current wego version
var Version = "v0.0.0"

// LessThan compares two version strings of the form: "vX.X.X"
func LessThan(s1, s2 string) (bool, error) {
	v1, err := version.ParseSemantic(s1)
	if err != nil {
		return false, err
	}
	v2, err := version.ParseSemantic(s2)
	if err != nil {
		return false, err
	}
	return v1.LessThan(v2), nil
}

// GetReleases looks up all releases for wego to determine if a new release is available
func GetReleases() ([]interface{}, error) {
	releaseQuery := "https://api.github.com/repos/weaveworks/weave-gitops/releases"
	// Create a new request using http
	req, err := http.NewRequest("GET", releaseQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate version request; local version is: %s\n", Version)
	}
	// Send req using http Client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve latest version information; local version is: %s\n", Version)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var data []interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse version information; local version is: %s\n", Version)
	}
	return data, nil
}

// ExtractLatestRelease finds the most recent release tag in the set of wego releases.
// The data is a slice of "map[string]interface{}"; each release corresponds to one map.
func ExtractLatestRelease(data []interface{}) (string, error) {
	version := "v0.0.0"

	for _, release := range data {
		relMap, ok := release.(map[string]interface{})
		if !ok {
			continue
		}
		relVersion, ok := relMap["tag_name"].(string)
		if !ok {
			continue
		}
		lt, err := LessThan(version, relVersion)
		if err != nil {
			continue
		}
		if lt {
			version = relVersion
		}
	}

	return version, nil
}
