package profiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/profiles"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	port string
)

var Cmd = &cobra.Command{
	Use:           "profile",
	Aliases:       []string{"profiles"},
	Short:         "Show information about available profiles",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: `
# Get all profiles
gitops get profiles
`,
	RunE: runCmd,
}

func init() {
	Cmd.Flags().StringVar(&port, "port", server.DefaultPort, "port the profiles API is running on")
}

func runCmd(cmd *cobra.Command, args []string) error {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return fmt.Errorf("error initializing kubernetes config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error initializing kubernetes client: %w", err)
	}

	ns, err := cmd.Parent().Parent().Flags().GetString("namespace")
	if err != nil {
		return err
	}

	return profiles.GetProfiles(context.Background(), profiles.GetOptions{
		Namespace: ns,
		ClientSet: clientSet,
		Writer:    os.Stdout,
		Port:      port,
	})
}
