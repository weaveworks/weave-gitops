package server

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestCreateKustomization_App_Association(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	if err != nil {
		t.Fatal(err)
	}

	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	err = k.Create(ctx, ns)
	if err != nil {
		t.Fatal(err)
	}

	r := &pb.AddKustomizationReq{
		Name:      "mykustomization",
		Namespace: ns.Name,
		AppName:   "someapp",
		SourceRef: &pb.SourceRef{
			Kind: pb.SourceRef_GitRepository,
			Name: "othersource",
		},
	}

	res, err := c.AddKustomization(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(res.Success).To(BeTrue())
}
