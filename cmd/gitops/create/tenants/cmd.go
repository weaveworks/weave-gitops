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
	Short: "Create or update tenant resources",
	Example: `
	  # Create a tenant using name and namespace flags
	  gitops create tenants --name test-tenant1 --namespace test-ns1 --namespace test-ns2

	  # Create tenants from a file
	  gitops create tenants --from-file tenants.yaml

	  # Export tenant resources to a file
	  gitops create tenants --from-file tenants.yaml --export > tenants.yaml

	  # Export tenant resources to stdout
	  gitops create tenants --from-file tenants.yaml --export
	`,
	RunE: createTenantsCmdRunE(),
}

func init() {
	TenantsCommand.Flags().StringVar(&flags.name, "name", "", "the name of the tenant to be created")
	TenantsCommand.Flags().StringSliceVar(&flags.namespaces, "namespace", []string{}, "a list of namespaces for the tenant")
	TenantsCommand.Flags().StringVar(&flags.fromFile, "from-file", "", "the file containing the tenant declarations")
	TenantsCommand.Flags().BoolVar(&flags.export, "export", false, "export in YAML format to stdout")

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
