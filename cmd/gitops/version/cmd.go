package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current wego version
var Version = "v0.0.0"
var GitCommit = ""
var Branch = ""
var BuildTime = ""

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Display gitops version",
	Run:   runCmd,
}

func runCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Current Version:", Version)
	fmt.Println("GitCommit:", GitCommit)
	fmt.Println("BuildTime:", BuildTime)
	fmt.Println("Branch:", Branch)
}
