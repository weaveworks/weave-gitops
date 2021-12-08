package check

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/version"

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
# Validate flux compatibility
gitops check --pre
`,
	RunE: runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("could not create client config: %w", err)
	}

	_, k8sClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return fmt.Errorf("could not create kube http client: %w", err)
	}

	k := &kube.KubeHTTP{
		Client: k8sClient,
	}

	namespacesList, err := k.GetNamespaces(ctx)
	if err != nil {
		return err
	}

	var fluxVersionIsCompatible bool

	for _, namespace := range namespacesList.Items {
		labels := namespace.GetLabels()
		if labels["app.kubernetes.io/part-of"] == "flux" &&
			labels["app.kubernetes.io/version"] == version.FluxVersion {
			fluxVersionIsCompatible = true

			fmt.Println("match", labels["app.kubernetes.io/version"])

			break
		}
	}

	fmt.Println("Expected flux version", version.FluxVersion)

	if fluxVersionIsCompatible {
		fmt.Println("current flux version is compatible")
	} else {
		fmt.Println("current flux version is not compatible")
	}

	return nil
}

func init() {
	Cmd.Flags().BoolVarP(&pre, "pre", "d", false, "TODO")
}
