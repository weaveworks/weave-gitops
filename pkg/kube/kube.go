package kube

import (
	"context"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Resource interface {
	metav1.Object
	runtime.Object
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ClusterStatus int

const (
	Unknown ClusterStatus = iota
	Unmodified
	FluxInstalled
	GitOpsInstalled
)

// Function to translate ClusterStatus to a string
func (cs ClusterStatus) String() string {
	return toStatusString[cs]
}

var toStatusString = map[ClusterStatus]string{
	Unknown:         "Unable to talk to the cluster",
	Unmodified:      "No flux or gitops installed",
	FluxInstalled:   "Flux installed",
	GitOpsInstalled: "GitOps installed",
}

//counterfeiter:generate . Kube
type Kube interface {
	Apply(ctx context.Context, manifest []byte, namespace string) error
	Delete(ctx context.Context, manifest []byte) error
	DeleteByName(ctx context.Context, name string, gvr schema.GroupVersionResource, namespace string) error
	SecretPresent(ctx context.Context, string, namespace string) (bool, error)
	GetApplications(ctx context.Context, namespace string) ([]wego.Application, error)
	FluxPresent(ctx context.Context) (bool, error)
	GetClusterName(ctx context.Context) (string, error)
	GetClusterStatus(ctx context.Context) ClusterStatus
	GetApplication(ctx context.Context, name types.NamespacedName) (*wego.Application, error)
	GetResource(ctx context.Context, name types.NamespacedName, resource Resource) error
	GetSecret(ctx context.Context, name types.NamespacedName) (*corev1.Secret, error)
}
