package flux

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

//go:embed flux
var fluxExe []byte

type fluxDep struct {
	Flux flux
}
type flux struct {
	Version string
}

var Cmd = &cobra.Command{
	Use:   "flux",
	Short: "Use flux commands",
	Run:   runCmd,
}
var exe = "temp/bin/flux-"

func init() {
	var f fluxDep
	if _, err := toml.DecodeFile("tools/dependencies.toml", &f); err != nil {
		fmt.Println(err)
	}
	exe = exe + f.Flux.Version
	_ = os.WriteFile(exe, fluxExe, 0755)
}

func runCmd(cmd *cobra.Command, args []string) {
	c := exec.Command("./"+exe, args...)

	// run command
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	} else {
		fmt.Printf("Output: %s\n", output)
	}
}
