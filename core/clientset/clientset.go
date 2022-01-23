package clientset

import (
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

var (
	Scheme         = runtime.NewScheme()
	Codecs         = serializer.NewCodecFactory(Scheme)
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

func init() {
	_ = sourcev1.AddToScheme(Scheme)
	_ = kustomizev2.AddToScheme(Scheme)
	_ = helmv2.AddToScheme(Scheme)
	_ = v1alpha2.AddToScheme(Scheme)
	_ = corev1.AddToScheme(Scheme)
	_ = extensionsv1.AddToScheme(Scheme)
	_ = appsv1.AddToScheme(Scheme)
}

type Set interface {
	AppClient() (*rest.RESTClient, error)
	KustomizationClient() (*rest.RESTClient, error)
	SourceClient() (*rest.RESTClient, error)
}

type clientSets struct {
	config *rest.Config
}

func NewClientSets(config *rest.Config) Set {
	return &clientSets{
		config: config,
	}
}

func (c clientSets) AppClient() (*rest.RESTClient, error) {
	return newRestClientWithConfig(c.config, &v1alpha2.GroupVersion)
}

func (c clientSets) KustomizationClient() (*rest.RESTClient, error) {
	return newRestClientWithConfig(c.config, &kustomizev2.GroupVersion)
}

func (c clientSets) SourceClient() (*rest.RESTClient, error) {
	return newRestClientWithConfig(c.config, &sourcev1.GroupVersion)
}

func newRestClientWithConfig(config *rest.Config, groupVersion *schema.GroupVersion) (*rest.RESTClient, error) {
	config.GroupVersion = groupVersion
	config.APIPath = "/apis"

	config.NegotiatedSerializer = Codecs.WithoutConversion()

	client, err := rest.RESTClientFor(config)

	if err != nil {
		return nil, fmt.Errorf("error creating rest client: %w", err)
	} else {
		return client, nil
	}
}
