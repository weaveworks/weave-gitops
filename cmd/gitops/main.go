package main

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
)

func main() {
	client := adapters.NewHTTPClient().EnableCLIAuth()

	if err := root.RootCmd(client).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
