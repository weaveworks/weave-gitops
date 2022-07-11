package tenants

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/tenancy"
)

type tenantCommandFlags struct {
	filename string
	export   bool
}

var flags tenantCommandFlags

var TenantsCommand = &cobra.Command{
	Use:   "tenants",
	Short: "create and update tenant resources",
	RunE:  createTenantsCmdRunE(),
}

func init() {
	TenantsCommand.PersistentFlags().StringVar(&flags.filename, "filename", "", "the file containing the tenant declarations")
	TenantsCommand.PersistentFlags().BoolVar(&flags.export, "export", false, "the file to export the generated resources to")

	cobra.CheckErr(TenantsCommand.MarkPersistentFlagRequired("filename"))
}

func createTenantsCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		err = tenancy.CreateTenants(ctx, flags.filename, kubeClient)
		if err != nil {
			return err
		}

		if flags.export {
			err = tenancy.ExportTenants(flags.filename, os.Stdout)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
