package kube

import (
	"k8s.io/client-go/rest"
)

// ClusterConfig is used to hold the default *rest.Config and the cluster name.
type ClusterConfig struct {
	DefaultConfig *rest.Config
	ClusterName   string
}
