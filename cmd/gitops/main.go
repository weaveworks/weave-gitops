package main

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
)

func main() {
	if err := root.RootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
