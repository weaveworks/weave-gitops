package main

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops-server/cmd"
)

func main() {
	cobra.CheckErr(cmd.NewCommand().Execute())
}
