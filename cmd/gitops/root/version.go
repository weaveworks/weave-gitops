package root

import (
	"fmt"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display gitops version",
	Run:   versionRun,
	PostRun: func(cmd *cobra.Command, args []string) {
		CheckVersion(CheckpointParams())
	},
}

func versionRun(cmd *cobra.Command, args []string) {
	fmt.Println("Current Version:", version.Version)
	fmt.Println("GitCommit:", version.GitCommit)
	fmt.Println("BuildTime:", version.BuildTime)
	fmt.Println("Branch:", version.Branch)
}

// CheckVersion looks to see if there is a newer version of the software available
func CheckVersion(p *checkpoint.CheckParams) {
	log := logger.NewCLILogger(os.Stdout)
	checkResponse, err := checkpoint.Check(p)

	if err == nil && checkResponse.Outdated {
		log.Println("gitops version %s is available; please update at %s",
			checkResponse.CurrentVersion, checkResponse.CurrentDownloadURL)
	}
}

// CheckpointParams creates the structure to pass to CheckVersion
func CheckpointParams() *checkpoint.CheckParams {
	return &checkpoint.CheckParams{
		Product: "weave-gitops",
		Version: version.Version,
	}
}

// CheckpointParamsWithFlags adds the object and command from the arguments list to the checkpoint parameters
func CheckpointParamsWithFlags(params *checkpoint.CheckParams, c *cobra.Command) *checkpoint.CheckParams {
	// wego uses noun verb command syntax and the parent command will have the noun and the command passed in will be the verb
	p := params
	if params == nil {
		p = CheckpointParams()
	}

	if c.HasParent() && c.Parent().Name() != "wego" {
		p.Flags = map[string]string{
			"object":  c.Parent().Name(),
			"command": c.Name(),
		}
	}

	return p
}
