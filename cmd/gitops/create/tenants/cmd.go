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
	name       string
	namespaces []string
	fromFile   string
	export     bool
}

var flags tenantCommandFlags

var TenantsCommand = &cobra.Command{
	Use:   "tenants",
	Short: "create and update tenant resources",
	RunE:  createTenantsCmdRunE(),
}

func init() {
	TenantsCommand.PersistentFlags().StringVar(&flags.name, "name", "", "the file containing the tenant declarations")
	TenantsCommand.PersistentFlags().StringSliceVar(&flags.namespaces, "namespace", []string{}, "the file containing the tenant declarations")
	TenantsCommand.PersistentFlags().StringVar(&flags.fromFile, "from-file", "", "the file containing the tenant declarations")
	TenantsCommand.PersistentFlags().BoolVar(&flags.export, "export", false, "the file to export the generated resources to")

	cobra.CheckErr(TenantsCommand.MarkPersistentFlagRequired("from-file"))
}

func createTenantsCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if flags.export {
			err := tenancy.ExportTenants(flags.fromFile, os.Stdout)
			if err != nil {
				return err
			}

			return nil
		}

		ctx := context.Background()

		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		if flags.fromFile != "" {
			tenants, err := tenancy.Parse(flags.fromFile)
			if err != nil {
				return fmt.Errorf("failed to parse tenants file %s for export: %w", flags.fromFile, err)
			}

			err = tenancy.CreateTenants(ctx, tenants, kubeClient)
			if err != nil {
				return err
			}
		}

		if flags.name != "" {
			if len(flags.namespaces) == 0 {
				return fmt.Errorf("at least one namespace is required for tenant: %s", flags.name)
			}

			tenant := []tenancy.Tenant{
				{
					Name:       flags.name,
					Namespaces: flags.namespaces,
				},
			}

			err = tenancy.CreateTenants(ctx, tenant, kubeClient)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
