package flux

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

//go:embed bin/flux
var fluxExe []byte

type dependencies struct {
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
var path = "temp/bin"
var exePath string

func init() {
	var dep dependencies
	if _, err := toml.DecodeFile("tools/dependencies.toml", &dep); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	exePath = path + "/flux-" + dep.Flux.Version
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(exePath, fluxExe, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func runCmd(cmd *cobra.Command, args []string) {
	c := exec.Command("./"+exePath, args...)

	// run command
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Output: %s\n", output)
	}
}
