package app

import (
	"context"
	"time"

	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type KubeFetcher interface {
	Get(ctx context.Context, client *rest.RESTClient, name, namespace string) (*v1alpha1.Application, error)
	List(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (*v1alpha1.ApplicationList, error)
}

func NewKubeAppFetcher() KubeFetcher {
	return &appKubeFetcher{}
}

type appKubeFetcher struct {
}

func (a appKubeFetcher) Get(ctx context.Context, client *rest.RESTClient, name, namespace string) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = client.Get().
		Namespace(namespace).
		Resource(apps).
		Name(name).
		Do(ctx).
		Into(result)

	return
}

func (a appKubeFetcher) List(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (result *v1alpha1.ApplicationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1alpha1.ApplicationList{}
	err = client.Get().
		Namespace(namespace).
		Resource(apps).
		VersionedParams(&opts, clientset.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}
