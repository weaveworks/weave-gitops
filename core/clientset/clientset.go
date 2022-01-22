package clientset

import (
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
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
	_ = wego.AddToScheme(Scheme)
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
}

func NewClientSets() Set {
	return &clientSets{}
}

func (c clientSets) AppClient() (*rest.RESTClient, error) {
	return newRestClient(&wego.GroupVersion)
}

func (c clientSets) KustomizationClient() (*rest.RESTClient, error) {
	return newRestClient(&kustomizev2.GroupVersion)
}

func (c clientSets) SourceClient() (*rest.RESTClient, error) {
	return newRestClient(&sourcev1.GroupVersion)
}

func newRestClient(groupVersion *schema.GroupVersion) (*rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("kube.NewClient cluster config: %w", err)
	}

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
