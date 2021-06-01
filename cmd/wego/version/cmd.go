package version

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave-gitops/pkg/version"

	"github.com/spf13/cobra"
)

// The current wego version
var Version = "v0.0.0"
var GitCommit = ""
var Branch = ""
var BuildTime = ""

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Display wego version",
	Run:   runCmd,
}

func runCmd(cmd *cobra.Command, args []string) {
	logStr, err := CheckVersion()
	if err != nil {
		log.Error(err)
	}

	if logStr != "" {
		log.Infof(logStr)
	}

	fmt.Println("Current Version ", Version)
	fmt.Println("GitCommit: ", GitCommit)
	fmt.Println("BuildTime: ", BuildTime)
	fmt.Println("Branch: ", Branch)
	fmt.Println("Flux Version: ", version.FluxVersion)
}

func CheckVersion() (string, error) {
	checkResponse, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: "weave-gitops",
		Version: Version,
	})

	if err != nil {
		return "", fmt.Errorf("Unable to retrieve latest version: %v", err)
	}

	if err == nil && checkResponse.Outdated {
		return fmt.Sprintf("wego version %s is available; please update at %s",
			checkResponse.CurrentVersion, checkResponse.CurrentDownloadURL), nil
	}

	return "", nil
}
