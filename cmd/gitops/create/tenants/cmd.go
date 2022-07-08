package tenants

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/tenancy"
)

type tenantCommandFlags struct {
	filename string
	export   string
}

var flags tenantCommandFlags

func TenantsCommand() *cobra.Command {
	tenantsCmd := &cobra.Command{
		Use:   "tenants",
		Short: "create and update tenant resources",
		RunE:  createTenantsCmdRunE(),
	}

	tenantsCmd.Flags().String(
		flags.filename,
		"",
		"the file containing the tenant declarations",
	)
	tenantsCmd.Flags().String(
		flags.export,
		"",
		"the file to export the generated resources to",
	)
	cobra.CheckErr(tenantsCmd.MarkFlagRequired(flags.filename))
	cobra.CheckErr(viper.BindPFlag(flags.filename, tenantsCmd.Flags().Lookup(flags.filename)))

	return tenantsCmd
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

		if flags.export != "" {
			err = tenancy.ExportTenants(flags.export, os.Stdout)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
