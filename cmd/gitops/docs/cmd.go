package docs

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docDir  string
	docFunc func()
)

var Cmd = &cobra.Command{
	Use:     "docs",
	Short:   "Generate hard copy of CLI documentation",
	Example: "gitops docs",
	Hidden:  true,
	Run:     runCmd,
}

func runCmd(cmd *cobra.Command, args []string) {
	docFunc()
}

func init() {
	Cmd.Flags().StringVarP(&docDir, "directory", "d", ".", "directory in which to place generated documents")
	docFunc = genDocs
}

func genDocs() {
	doc.GenMarkdownTree(Cmd.Root(), docDir)
}
