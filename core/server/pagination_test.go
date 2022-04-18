package server_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"

	"k8s.io/apimachinery/pkg/util/rand"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/server"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
)

func TestPagination(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cfg := makeGRPCServer(k8sEnv.Rest, t)

	k, err := client.New(k8sEnv.Rest, client.Options{
		Scheme: kube.CreateScheme(),
	})
	g.Expect(err).NotTo(HaveOccurred())

	testingKustomizations := map[string]int{
		"ns1": 4,
		"ns2": 2,
		"ns3": 9,
	}

	for nsName, kustomizationsSize := range testingKustomizations {
		ns := newNamespaceWithPrefix(ctx, k, g, nsName+"-")

		for i := 0; i < kustomizationsSize; i++ {
			kust := &kustomizev1.Kustomization{
				Spec: kustomizev1.KustomizationSpec{
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind: "GitRepository",
					},
				},
			}
			kust.Namespace = ns.Name
			kust.Name = fmt.Sprintf("%s-k-%d", nsName, i)

			g.Expect(k.Create(ctx, kust)).To(Succeed())
		}
	}

	updateNamespaceCache(cfg)

	var pageSize int32 = 3

	// NextPageToken values
	tables := []struct {
		Namespace     string
		EmptyK8sToken bool
	}{
		{
			Namespace:     "ns1",
			EmptyK8sToken: false,
		},
		{
			Namespace:     "ns3",
			EmptyK8sToken: true,
		},
		{
			Namespace:     "ns3",
			EmptyK8sToken: false,
		},
		{
			Namespace:     "ns3",
			EmptyK8sToken: false,
		},
		{
			Namespace:     "ns3",
			EmptyK8sToken: true,
		},
	}

	var previousNextPageToken string

	for ind, table := range tables {
		res, err := c.ListKustomizations(ctx, &pb.ListKustomizationsRequest{
			Pagination: &pb.Pagination{
				PageSize:  pageSize,
				PageToken: previousNextPageToken,
			},
		})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res.Kustomizations).To(HaveLen(int(pageSize)))

		if ind == len(tables)-1 {
			g.Expect(res.NextPageToken).To(BeEmpty())
		} else {
			checkPageTokenInfo(g, res.NextPageToken, table.Namespace, table.EmptyK8sToken)
		}

		previousNextPageToken = res.NextPageToken
	}
}

func checkPageTokenInfo(g *WithT, actualNextPageToken string, expectedNamespace string, expectedEmptyK8sPage bool) {
	var actualNextPageInfo server.PageTokenInfo

	err := decodeFromBase64(&actualNextPageInfo, actualNextPageToken)

	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(actualNextPageInfo.Namespace).Should(MatchRegexp(fmt.Sprintf("%s*", expectedNamespace)))
	g.Expect(actualNextPageInfo.K8sPageToken == "").Should(Equal(expectedEmptyK8sPage))
}

func newNamespaceWithPrefix(ctx context.Context, k client.Client, g *GomegaWithT, nsPrefix string) corev1.Namespace {
	ns := corev1.Namespace{}
	ns.Name = nsPrefix + rand.String(5)

	g.Expect(k.Create(ctx, &ns)).To(Succeed())

	return ns
}

func decodeFromBase64(v interface{}, enc string) error {
	return json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(enc))).Decode(v)
}
