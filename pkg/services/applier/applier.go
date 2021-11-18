package applier

import (
	// "bytes"
	"context"
	// "crypto/md5"
	"fmt"
	// "os"
	// "path/filepath"
	// "strings"

	// "github.com/fluxcd/go-git-providers/gitprovider"
	// sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	// "github.com/fluxcd/source-controller/pkg/sourceignore"
	// "github.com/weaveworks/weave-gitops/pkg/flux"
	// "github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	// wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// "sigs.k8s.io/kustomize/api/types"
	// "sigs.k8s.io/yaml"
)

type ClusterApplier interface {
	CreateNamespace(ctx context.Context, namespace string) error
	ApplyManifests(ctx context.Context, cluster models.Cluster, namespace string, manifests []automation.AutomationManifest) error
}

type ClusterApplySvc struct {
	Kube   kube.Kube
	Client client.Client
	Logger logger.Logger
}

var _ ClusterApplier = &ClusterApplySvc{}

func NewClusterApplier(kubeClient kube.Kube, rawClient client.Client, logger logger.Logger) ClusterApplier {
	return &ClusterApplySvc{
		Kube:   kubeClient,
		Client: rawClient,
		Logger: logger,
	}
}

func (a *ClusterApplySvc) CreateNamespace(ctx context.Context, namespace string) error {
	var ns corev1.Namespace

	ns.Name = namespace
	return a.Client.Create(ctx, &ns)
}

func (a *ClusterApplySvc) ApplyManifests(ctx context.Context, cluster models.Cluster, namespace string, manifests []automation.AutomationManifest) error {
	for _, manifest := range manifests {
		fmt.Printf("NS: %s\nCONTENT: %s\n", namespace, manifest.Content)

		if err := a.Kube.Apply(ctx, manifest.Content, namespace); err != nil {
			return fmt.Errorf("could not apply manifest: %w", err)
		}
	}

	return nil
}
