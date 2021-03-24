package version

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Display wego version",
	Run:   runCmd,
}

func extractLatestRelease(data []interface{}) (string, error) {
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

func runCmd(cmd *cobra.Command, args []string) {
	releases, err := GetReleases()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	latestRelease, err := extractLatestRelease(releases)
	if err != nil {
		fmt.Printf("Failed to extract version information; local version is: %s\n", Version)
		os.Exit(1)
	}
	lt, err := LessThan(Version, latestRelease)
	if err != nil {
		fmt.Printf("Failed comparing versions; local version is: %s\n", Version)
		os.Exit(1)
	}
	if lt {
		fmt.Printf("Current version is: %s. A newer version (%s) is available.\n", Version, latestRelease)
	} else {
		fmt.Println(Version)
	}
}
