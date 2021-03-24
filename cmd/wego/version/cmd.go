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

// exitOnError prints an error to stderr and exit; if error is nil, does nothing
func exitOnError(err interface{}, msgs ...string) {
	if err == nil {
		return
	}
	if len(msgs) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msgs[0])
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(1)
}

func runCmd(cmd *cobra.Command, args []string) {
	releases, err := GetReleases()
	exitOnError(err)

	latestRelease, err := ExtractLatestRelease(releases)
	exitOnError(err, fmt.Sprintf("Failed to extract version information; local version is: %s\n", Version))

	lt, err := LessThan(Version, latestRelease)
	exitOnError(err, fmt.Sprintf("Failed comparing versions; local version is: %s\n", Version))

	if lt {
		fmt.Printf("Current version is: %s. A newer version (%s) is available.\n", Version, latestRelease)
	} else {
		fmt.Println(Version)
	}
}
