package main

import (
	"fmt"
	"os"

	multi_cluster "github.com/weaveworks/weave-gitops/cmd/multi-cluster/cmd"
)

func main() {
	if err := multi_cluster.NewServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
