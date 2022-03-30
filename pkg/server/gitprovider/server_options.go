package gitprovider

import "github.com/weaveworks/weave-gitops/pkg/kube"

// AuthOptions includes all the options that can be set for an
// GitProviderAuth server.
type AuthOptions struct {
	ClientGetter kube.ClientGetter
	KubeGetter   kube.KubeGetter
}

// AuthOption defines the signature of a function that can be used
// to set an option for an GitProviderAuth server.
type AuthOption func(*AuthOptions)

// WithClientGetter allows for setting a ClientGetter.
func WithClientGetter(clientGetter kube.ClientGetter) AuthOption {
	return func(args *AuthOptions) {
		args.ClientGetter = clientGetter
	}
}

// WithKubeGetter allows for setting a KubeGetter.
func WithKubeGetter(kubeGetter kube.KubeGetter) AuthOption {
	return func(args *AuthOptions) {
		args.KubeGetter = kubeGetter
	}
}
