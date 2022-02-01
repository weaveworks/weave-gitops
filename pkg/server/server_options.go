package server

import "github.com/weaveworks/weave-gitops/pkg/kube"

// ApplicationsOptions includes all the options that can be set for an
// ApplicationsServer.
type ApplicationsOptions struct {
	ClientGetter kube.ClientGetter
	KubeGetter   kube.KubeGetter
}

// ApplicationsOption defines the signature of a function that can be used
// to set an option for an ApplicationsServer.
type ApplicationsOption func(*ApplicationsOptions)

// WithClientGetter allows for setting a ClientGetter.
func WithClientGetter(clientGetter kube.ClientGetter) ApplicationsOption {
	return func(args *ApplicationsOptions) {
		args.ClientGetter = clientGetter
	}
}

// WithKubeGetter allows for setting a KubeGetter.
func WithKubeGetter(kubeGetter kube.KubeGetter) ApplicationsOption {
	return func(args *ApplicationsOptions) {
		args.KubeGetter = kubeGetter
	}
}
