package crd_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/services/crd"
)

const defaultClusterName = "Default"

var k8sEnv *testutils.K8sTestEnv

func TestMain(m *testing.M) {
	var err error

	k8sEnv, err = testutils.StartK8sTestEnvironment([]string{
		"../../../manifests/crds",
		"../../../tools/testcrds",
	})

	if err != nil {
		panic(err)
	}

	code := m.Run()

	k8sEnv.Stop()

	os.Exit(code)
}

func newService(ctx context.Context, k8sEnv *testutils.K8sTestEnv) (crd.Fetcher, error) {
	_, clustersManager, err := createClient(k8sEnv)
	if err != nil {
		return nil, err
	}

	log := logr.Discard()

	return crd.NewFetcher(ctx, log, clustersManager), nil
}

func createClient(k8sEnv *testutils.K8sTestEnv) (clustersmngr.Client, clustersmngr.ClustersManager, error) {
	ctx := context.Background()
	log := logr.Discard()

	singleCluster, err := cluster.NewSingleCluster(defaultClusterName, k8sEnv.Rest, nil)
	if err != nil {
		return nil, nil, err
	}

	fetcher := &clustersmngrfakes.FakeClusterFetcher{}
	fetcher.FetchReturns([]cluster.Cluster{singleCluster}, nil)

	clustersManager := clustersmngr.NewClustersManager(
		[]clustersmngr.ClusterFetcher{fetcher},
		nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		log,
	)

	if err := clustersManager.UpdateClusters(ctx); err != nil {
		return nil, nil, err
	}
	if err := clustersManager.UpdateNamespaces(ctx); err != nil {
		return nil, nil, err
	}

	client, err := clustersManager.GetServerClient(ctx)

	return client, clustersManager, err
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
	g *gomega.GomegaWithT,
	k client.Client,
	info CRDInfo,
) v1.CustomResourceDefinition {
	resource := v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", info.Plural, info.Group),
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: info.Group,
			Names: v1.CustomResourceDefinitionNames{
				Plural:   info.Plural,
				Singular: info.Singular,
				Kind:     info.Kind,
			},
			Scope: "Namespaced",
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:    "v1beta1",
					Served:  true,
					Storage: true,
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]v1.JSONSchemaProps{},
						},
					},
				},
			},
		},
	}

	err := k.Create(ctx, &resource)

	if !info.NoTest {
		g.Expect(err).ToNot(gomega.HaveOccurred(), "should be able to create crd: %s", resource.GetName())
	}

	return resource
}
