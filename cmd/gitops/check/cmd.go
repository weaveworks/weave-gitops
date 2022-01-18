package check

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/services/check"
	"github.com/weaveworks/weave-gitops/pkg/version"

	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"

	"github.com/weaveworks/weave-gitops/pkg/flux"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

var (
	pre bool
)

var Cmd = &cobra.Command{
	Use:   "check",
	Short: "Validates flux compatibility",
	Example: `
# Validate flux and kubernetes compatibility
gitops check --pre
`,
	RunE: runCmd,
}

func init() {
	Cmd.Flags().BoolVarP(&pre, "pre", "p", true, "perform only the pre-installation checks")
}

func runCmd(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("failed getting rest config: %w", err)
	}

	_, k8sClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return fmt.Errorf("failed creating k8s client: %w", err)
	}

	kubeClient := &kube.KubeHTTP{
		Client: k8sClient,
	}

	output, err := check.Pre(ctx, kubeClient, fluxClient, version.FluxVersion)
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
