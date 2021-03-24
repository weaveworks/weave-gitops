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

func runCmd(cmd *cobra.Command, args []string) {
	releases, err := GetReleases()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	latestRelease, err := ExtractLatestRelease(releases)
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
