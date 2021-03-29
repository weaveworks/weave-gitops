package version

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weaveworks/go-checkpoint"
)

// The current wego version
var Version = "v0.0.0"

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Display wego version",
	Run:   runCmd,
}

func runCmd(cmd *cobra.Command, args []string) {
	if checkResponse, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: "weave-gitops",
		Version: Version,
	}); err == nil && checkResponse.Outdated {
		log.Infof("wego version %s is available; please update at %s",
			checkResponse.CurrentVersion, checkResponse.CurrentDownloadURL)
	} else {
		fmt.Println(Version)
	}
}
