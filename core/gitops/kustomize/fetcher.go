package kustomize

import (
	"context"
	"time"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type Fetcher interface {
	Get(ctx context.Context, client *rest.RESTClient, name, namespace string) (*v1beta2.Kustomization, error)
	List(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (*v1beta2.KustomizationList, error)
}

func NewKustomizationFetcher() Fetcher {
	return &kustomizeFetcher{}
}

type kustomizeFetcher struct {
}

func (a kustomizeFetcher) Get(ctx context.Context, client *rest.RESTClient, name, namespace string) (result *v1beta2.Kustomization, err error) {
	result = &v1beta2.Kustomization{}
	err = client.Get().
		Namespace(namespace).
		Resource(kustomizations).
		Name(name).
		Do(ctx).
		Into(result)

	return
}

func (a kustomizeFetcher) List(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (result *v1beta2.KustomizationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1beta2.KustomizationList{}
	err = client.Get().
		Namespace(namespace).
		Resource(kustomizations).
		VersionedParams(&opts, clientset.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}
