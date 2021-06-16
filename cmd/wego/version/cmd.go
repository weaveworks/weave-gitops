package version

import (
	"fmt"
	"strings"

	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"

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
	fmt.Println("Current Version:", Version)
	fmt.Println("GitCommit:", GitCommit)
	fmt.Println("BuildTime:", BuildTime)
	fmt.Println("Branch:", Branch)
	version, err := CheckFluxVersion()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Flux Version:", version)
}

func CheckVersion() (string, error) {
	checkResponse, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: "weave-gitops",
		Version: Version,
	})

	if err != nil {
		return "", fmt.Errorf("unable to retrieve latest version: %v", err)
	}

	if checkResponse.Outdated {
		return fmt.Sprintf("wego version %s is available; please update at %s",
			checkResponse.CurrentVersion, checkResponse.CurrentDownloadURL), nil
	}

	return "", nil
}

func CheckFluxVersion() (string, error) {
	output, err := fluxops.CallFlux("-v")
	if err != nil {
		return "", err
	}

	// change string format to match other version info
	version := strings.ReplaceAll(string(output), "flux version ", "v")

	return version, nil
}
