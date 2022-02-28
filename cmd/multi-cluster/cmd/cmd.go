package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Short:   "Runs the multi-cluster",
		PreRunE: preRunCmd,
		RunE:    runCmd,
	}

	return cmd
}

func preRunCmd(cmd *cobra.Command, args []string) error {
	return nil
}

func runCmd(cmd *cobra.Command, args []string) error {
	restConfig := config.GetConfigOrDie()
	cl, err := client.New(restConfig, client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	clustersCm := &corev1.ConfigMap{}
	clusterCmKey := types.NamespacedName{
		Name:      "clusters",
		Namespace: "default",
	}

	err = cl.Get(context.Background(), clusterCmKey, clustersCm)
	if err != nil {
		fmt.Printf("failed to list pods in namespace default: %v\n", err)
		os.Exit(1)
	}

	clusters := []cluster{}

	err = yaml.Unmarshal([]byte(clustersCm.Data["clusters"]), &clusters)
	if err != nil {
		fmt.Printf("failed to unmarshaling clusters: %v\n", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal([]byte(clustersCm.Data["clusters"]), &clusters)
	if err != nil {
		fmt.Printf("failed to unmarshaling clusters: %v\n", err)
		os.Exit(1)
	}

	cluster := clusters[0]

	clusterSecret := &corev1.Secret{}
	clusterSecretKey := types.NamespacedName{
		Name:      cluster.SecretRef,
		Namespace: "default",
	}

	err = cl.Get(context.Background(), clusterSecretKey, clusterSecret)
	if err != nil {
		fmt.Printf("failed to list pods in namespace default: %v\n", err)
		os.Exit(1)
	}

	leafConfig := &rest.Config{
		Host:        cluster.Server,
		BearerToken: string(clusterSecret.Data["token"]),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // TODO: proper certs loading
		},
		Impersonate: rest.ImpersonationConfig{
			UserName: "luiz.filho@weave.works",
			Groups:   []string{"developers"},
		},
	}

	leafClient, err := client.New(leafConfig, client.Options{})
	if err != nil {
		fmt.Println("failed to create leaf client", err)
		os.Exit(1)
	}

	podList := &corev1.PodList{}

	err = leafClient.List(context.Background(), podList, client.InNamespace("developers"))
	if err != nil {
		fmt.Printf("failed to list pods in namespace default: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(podList)
	return nil
}

type cluster struct {
	Name                     string `yaml:"name"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
	SecretRef                string `yaml:"secretRef"`
}
