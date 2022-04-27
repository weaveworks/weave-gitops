package main

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/weaveworks/weave-gitops/gitops/cmd/gitops/root"
)

func main() {
	if err := root.RootCmd(resty.New()).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
