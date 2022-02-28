package main

import (
	"fmt"
	"os"

	ui "github.com/weaveworks/weave-gitops/cmd/multi-cluster/cmd"
)

func main() {
	if err := ui.NewCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
