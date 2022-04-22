package kube

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource interface {
	client.Object
}

type WegoConfig struct {
	FluxNamespace string
	WegoNamespace string
	ConfigRepo    string
}

// This is a wrapper around the client.Client to provide extra information
// specifically designed to query the K8s API directly instead of relying on
// `kubectl` to be present in the PATH.
type KubeHTTP struct {
	client.Client
	ClusterName string
}
