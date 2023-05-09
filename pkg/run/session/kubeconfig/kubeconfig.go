package kubeconfig

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	DefaultSecretPrefix = "vc-"
	KubeconfigSecretKey = "config"
)

func ReadKubeConfig(ctx context.Context, client *kubernetes.Clientset, suffix, namespace string) (*api.Config, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, GetDefaultSecretName(suffix), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not Get the %s secret in order to read kubeconfig: %v", GetDefaultSecretName(suffix), err)
	}
	config, found := secret.Data[KubeconfigSecretKey]
	if !found {
		return nil, fmt.Errorf("could not find the kube config (%s key) in the %s secret", KubeconfigSecretKey, GetDefaultSecretName(suffix))
	}
	return clientcmd.Load(config)
}

func GetDefaultSecretName(suffix string) string {
	return DefaultSecretPrefix + suffix
}
