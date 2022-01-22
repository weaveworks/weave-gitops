package source

import (
	"context"

	"github.com/fluxcd/source-controller/api/v1beta1"
	"k8s.io/client-go/rest"
)

const (
	buckets          = "buckets"
	gitRepositories  = "gitrepositories"
	helmRepositories = "helmrepositories"
)

type KubeCreator interface {
	CreateBucket(ctx context.Context, client *rest.RESTClient, bucket *v1beta1.Bucket) (result *v1beta1.Bucket, err error)
	CreateGitRepository(ctx context.Context, client *rest.RESTClient, gitRepository *v1beta1.GitRepository) (result *v1beta1.GitRepository, err error)
	CreateHelmRepository(ctx context.Context, client *rest.RESTClient, helmRepository *v1beta1.HelmRepository) (result *v1beta1.HelmRepository, err error)
}

func NewKubeCreator() KubeCreator {
	return &sourceCreator{}
}

type sourceCreator struct {
}

func (s sourceCreator) CreateBucket(ctx context.Context, client *rest.RESTClient, bucket *v1beta1.Bucket) (result *v1beta1.Bucket, err error) {
	result = &v1beta1.Bucket{}
	err = client.Post().
		Namespace(bucket.ObjectMeta.Namespace).
		Resource(buckets).
		Name(bucket.ObjectMeta.Name).
		Body(bucket).
		Do(ctx).
		Into(result)

	return
}

func (s sourceCreator) CreateGitRepository(ctx context.Context, client *rest.RESTClient, gitRepository *v1beta1.GitRepository) (result *v1beta1.GitRepository, err error) {
	result = &v1beta1.GitRepository{}
	err = client.Post().
		Namespace(gitRepository.ObjectMeta.Namespace).
		Resource(gitRepositories).
		Name(gitRepository.ObjectMeta.Name).
		Body(gitRepository).
		Do(ctx).
		Into(result)

	return
}

func (s sourceCreator) CreateHelmRepository(ctx context.Context, client *rest.RESTClient, helmRepository *v1beta1.HelmRepository) (result *v1beta1.HelmRepository, err error) {
	result = &v1beta1.HelmRepository{}
	err = client.Post().
		Namespace(helmRepository.ObjectMeta.Namespace).
		Resource(helmRepositories).
		Name(helmRepository.ObjectMeta.Name).
		Body(helmRepository).
		Do(ctx).
		Into(result)

	return
}
