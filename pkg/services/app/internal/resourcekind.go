package internal

import (
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ResourceKindApplication    ResourceKind = "Application"
	ResourceKindSecret         ResourceKind = "Secret"
	ResourceKindGitRepository  ResourceKind = "GitRepository"
	ResourceKindHelmRepository ResourceKind = "HelmRepository"
	ResourceKindKustomization  ResourceKind = "Kustomization"
	ResourceKindHelmRelease    ResourceKind = "HelmRelease"
)

type ResourceKind string

type ResourceRef struct {
	Kind           ResourceKind
	Name           string
	RepositoryPath string
}

func (rk ResourceKind) ToGVR() (schema.GroupVersionResource, error) {
	switch rk {
	case ResourceKindApplication:
		return kube.GVRApp, nil
	case ResourceKindSecret:
		return kube.GVRSecret, nil
	case ResourceKindGitRepository:
		return kube.GVRGitRepository, nil
	case ResourceKindHelmRepository:
		return kube.GVRHelmRepository, nil
	case ResourceKindHelmRelease:
		return kube.GVRHelmRelease, nil
	case ResourceKindKustomization:
		return kube.GVRKustomization, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("no matching schema.GroupVersionResource to the ResourceKind: %s", string(rk))
	}
}
