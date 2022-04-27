package main

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/gitops-server/cmd/gitops-server/cmd"
)

func main() {
	if err := cmd.NewCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
