package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/core/multicluster"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	hubClusterName  string
	leafClusterName string
	namespace       string
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Short: "adds a new cluster",
		RunE:  runE,
	}

	cmd.Flags().StringVarP(&hubClusterName, "hub-cluster-name", "", "", "hub cluster name")
	cmd.Flags().StringVarP(&leafClusterName, "leaf-cluster-name", "", "", "leaf cluster name")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace")
	cmd.MarkFlagRequired("hub-cluster-name")
	cmd.MarkFlagRequired("leaft-cluster-name")

	return cmd
}

func runE(cmd *cobra.Command, args []string) error {
	leafRestConfig, err := config.GetConfigWithContext(leafClusterName)
	if err != nil {
		log.Fatalf("failed getting leaf cluster config: %w", err)
	}

	leafClient, err := client.New(leafRestConfig, client.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	svcAcct := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{Name: "weave-gitops-server", Namespace: namespace},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, leafClient, svcAcct, func() error {
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating service account: %v\n", err)
		os.Exit(1)
	}

	wait.Poll(time.Millisecond*500, time.Second*10, func() (bool, error) {
		if err := leafClient.Get(ctx, client.ObjectKeyFromObject(svcAcct), svcAcct); err != nil {
			log.Fatalf("failed getting service account: %s", err)
		}

		if len(svcAcct.Secrets) > 0 {
			return true, nil
		}

		return false, nil
	})

	svcAcctSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{Name: svcAcct.Secrets[0].Name, Namespace: namespace},
	}
	if err := leafClient.Get(ctx, client.ObjectKeyFromObject(svcAcctSecret), svcAcctSecret); err != nil {
		log.Fatalf("failed getting service account secret: %s", err)
	}

	// Acting on Hub Cluster
	leafClusterToken := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{Name: leafClusterName + "-token", Namespace: namespace},
		Data:       svcAcctSecret.Data,
	}

	hubRestConfig, err := config.GetConfigWithContext(hubClusterName)
	if err != nil {
		log.Fatalf("failed creating hub client: %w", err)
	}

	hubClient, err := client.New(hubRestConfig, client.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting config: %v\n", err)
		os.Exit(1)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, hubClient, leafClusterToken, func() error {
		return nil
	})

	clustersCm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      multicluster.ClustersConfigMapName,
			Namespace: namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, hubClient, clustersCm, func() error {
		clusters := []multicluster.Cluster{}

		if val, ok := clustersCm.Data["clusters"]; ok {
			err = yaml.Unmarshal([]byte(val), &clusters)
			if err != nil {
				return fmt.Errorf("failed to unmarshaling clusters: %w", err)
			}
		}

		clusters = append(clusters, multicluster.Cluster{
			Name:      leafClusterName,
			Server:    hubRestConfig.Host,
			SecretRef: leafClusterToken.Name,
		})

		clustersData, err := yaml.Marshal(clusters)
		if err != nil {
			return fmt.Errorf("failed marshalling clusters list: %w", err)
		}

		clustersCm.Data = map[string]string{
			"clusters": string(clustersData),
		}

		return nil
	})

	return nil
}

func main() {
	if err := newCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
