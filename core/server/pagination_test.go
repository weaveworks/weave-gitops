package server_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
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
		"000-ns1": 4,
		"001-ns2": 2,
		"002-ns3": 9,
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
		Namespace      string
		NamespaceIndex int
		EmptyK8sToken  bool
	}{
		{
			Namespace:      "000-ns1",
			NamespaceIndex: 0, // This is going to be the initial index as there are other namespaces like default, kube-node-lease, kube-public and kube-system
			EmptyK8sToken:  false,
		},
		{
			Namespace:      "002-ns3",
			NamespaceIndex: 2,
			EmptyK8sToken:  true,
		},
		{
			Namespace:      "002-ns3",
			NamespaceIndex: 2,
			EmptyK8sToken:  false,
		},
		{
			Namespace:      "002-ns3",
			NamespaceIndex: 2,
			EmptyK8sToken:  false,
		},
		{
			Namespace:      "002-ns3",
			NamespaceIndex: 2,
			EmptyK8sToken:  true,
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

		expectedNextPageToken := server.PageTokenInfo{
			Namespace:      table.Namespace,
			NamespaceIndex: table.NamespaceIndex,
		}

		// Do not validate next token in the last expected page
		// There will be namespaces from other tests that we don't need
		// to test against.
		if ind != len(tables)-1 {
			checkPageTokenInfo(g, res.NextPageToken, expectedNextPageToken, table.EmptyK8sToken)
		}

		previousNextPageToken = res.NextPageToken
	}

	// If this is the last page then just check that the namespace info
	// in the token doesn't match the current testing namespaces
	var actualNextPageInfo server.PageTokenInfo

	err = decodeFromBase64(&actualNextPageInfo, previousNextPageToken)

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(actualNextPageInfo.Namespace).ShouldNot(MatchRegexp("^00"))
	g.Expect(strconv.Itoa(actualNextPageInfo.NamespaceIndex)).ShouldNot(MatchRegexp("[0-2]"))

}

func checkPageTokenInfo(g *WithT, actualNextPageToken string, expectedPageInfo server.PageTokenInfo, expectedEmptyK8sPage bool) {
	var actualNextPageInfo server.PageTokenInfo

	err := decodeFromBase64(&actualNextPageInfo, actualNextPageToken)

	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(actualNextPageInfo.Namespace).Should(MatchRegexp(fmt.Sprintf("%s*", expectedPageInfo.Namespace)))
	g.Expect(actualNextPageInfo.NamespaceIndex).Should(Equal(expectedPageInfo.NamespaceIndex))
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
