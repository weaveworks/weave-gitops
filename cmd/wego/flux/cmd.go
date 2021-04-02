package flux

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

//nolint
//go:embed flux
var test []byte

var Cmd = &cobra.Command{
	Use:   "flux",
	Short: "Use flux commands",
	Run:   runCmd,
}

func init() {
	_ = os.WriteFile("bin/flux", test, 0755)
}

func runCmd(cmd *cobra.Command, args []string) {
	c := exec.Command("./bin/flux", args...)

	// run command
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Output: %s\n", output)
	}
}
