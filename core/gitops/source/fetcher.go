package source

import (
	"context"
	"time"

	"github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type KubeFetcher interface {
	GetBuckets(ctx context.Context, client *rest.RESTClient, name, namespace string) (*v1beta1.Bucket, error)
	ListBuckets(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (*v1beta1.BucketList, error)
	GetGitRepository(ctx context.Context, client *rest.RESTClient, name, namespace string) (*v1beta1.GitRepository, error)
	ListGitRepositories(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (*v1beta1.GitRepositoryList, error)
	GetHelmRepository(ctx context.Context, client *rest.RESTClient, name, namespace string) (*v1beta1.HelmRepository, error)
	ListHelmRepositories(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (*v1beta1.HelmRepositoryList, error)
}

func NewSourceFetcher() KubeFetcher {
	return &sourceFetcher{}
}

type sourceFetcher struct {
}

func (a sourceFetcher) GetBuckets(ctx context.Context, client *rest.RESTClient, name, namespace string) (result *v1beta1.Bucket, err error) {
	result = &v1beta1.Bucket{}
	err = client.Get().
		Namespace(namespace).
		Resource(buckets).
		Name(name).
		Do(ctx).
		Into(result)

	return
}

func (a sourceFetcher) ListBuckets(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (result *v1beta1.BucketList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1beta1.BucketList{}
	err = client.Get().
		Namespace(namespace).
		Resource(buckets).
		VersionedParams(&opts, clientset.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}

func (a sourceFetcher) GetGitRepository(ctx context.Context, client *rest.RESTClient, name, namespace string) (result *v1beta1.GitRepository, err error) {
	result = &v1beta1.GitRepository{}
	err = client.Get().
		Namespace(namespace).
		Resource(gitRepositories).
		Name(name).
		Do(ctx).
		Into(result)

	return
}

func (a sourceFetcher) ListGitRepositories(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (result *v1beta1.GitRepositoryList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1beta1.GitRepositoryList{}
	err = client.Get().
		Namespace(namespace).
		Resource(gitRepositories).
		VersionedParams(&opts, clientset.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}

func (a sourceFetcher) GetHelmRepository(ctx context.Context, client *rest.RESTClient, name, namespace string) (result *v1beta1.HelmRepository, err error) {
	result = &v1beta1.HelmRepository{}
	err = client.Get().
		Namespace(namespace).
		Resource(helmRepositories).
		Name(name).
		Do(ctx).
		Into(result)

	return
}

func (a sourceFetcher) ListHelmRepositories(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (result *v1beta1.HelmRepositoryList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1beta1.HelmRepositoryList{}
	err = client.Get().
		Namespace(namespace).
		Resource(helmRepositories).
		VersionedParams(&opts, clientset.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}
