package root

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docDir  string
	docFunc func() error
)

var docsCmd = &cobra.Command{
	Use:     "docs",
	Short:   "Generate hard copy of CLI documentation",
	Example: "gitops docs",
	Hidden:  true,
	RunE:    runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {
	return docFunc()
}

func init() {
	docsCmd.Flags().StringVarP(&docDir, "directory", "d", ".", "directory in which to place generated documents")

	docFunc = genDocs
}

func genDocs() error {
	return doc.GenMarkdownTree(docsCmd.Root(), docDir)
}
