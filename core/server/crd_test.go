package server_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestIsAvailable(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := t.Context()

	c := makeGRPCServer(ctx, t, k8sEnv.Rest)

	scheme, err := kube.CreateScheme()
	g.Expect(err).NotTo(HaveOccurred())

	_, err = client.New(k8sEnv.Rest, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	resp, err := c.IsCRDAvailable(ctx, &pb.IsCRDAvailableRequest{Name: "customobjects.example.com"})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp.GetClusters()).To(HaveKey("Default"))
	g.Expect(resp.GetClusters()["Default"]).To(BeFalse())

	newCRD(ctx, g, k,
		CRDInfo{
			Singular: "customobject",
			Group:    "example.com",
			Plural:   "customobjects",
			Kind:     "CustomObject",
		})

	resp, err = c.IsCRDAvailable(ctx, &pb.IsCRDAvailableRequest{Name: "customobjects.example.com"})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp.GetClusters()).To(HaveKey("Default"))
	g.Expect(resp.GetClusters()["Default"]).To(BeTrue())
}

type CRDInfo struct {
	Group    string
	Plural   string
	Singular string
	Kind     string
	NoTest   bool
}

func newCRD(
	ctx context.Context,
	g *GomegaWithT,
	k client.Client,
	info CRDInfo,
) {
	resource := extv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", info.Plural, info.Group),
		},
		Spec: extv1.CustomResourceDefinitionSpec{
			Group: info.Group,
			Names: extv1.CustomResourceDefinitionNames{
				Plural:   info.Plural,
				Singular: info.Singular,
				Kind:     info.Kind,
			},
			Scope: "Namespaced",
			Versions: []extv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1beta1",
					Served:  true,
					Storage: true,
					Schema: &extv1.CustomResourceValidation{
						OpenAPIV3Schema: &extv1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]extv1.JSONSchemaProps{},
						},
					},
				},
			},
		},
	}

	err := k.Create(ctx, &resource)

	if !info.NoTest {
		g.Expect(err).NotTo(HaveOccurred())
	}
}
