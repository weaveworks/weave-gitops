package server

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := newTestScheme(t)

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		WithIndex(&corev1.Event{}, "type", client.IndexerFunc(func(o client.Object) []string {
			event := o.(*corev1.Event)
			return []string{event.Type}
		})).Build()
}

func newTestScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		pacv2beta2.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	return scheme
}

func makePolicy(t *testing.T, opts ...func(p *pacv2beta2.Policy)) *pacv2beta2.Policy {
	t.Helper()
	policy := &pacv2beta2.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "weave.policies.missing-owner-label",
		},
		Spec: pacv2beta2.PolicySpec{
			Name:     "Missing Owner Label",
			Severity: "high",
			Code:     "foo",
			Targets: pacv2beta2.PolicyTargets{
				Labels:     []map[string]string{{"my-label": "my-value"}},
				Kinds:      []string{},
				Namespaces: []string{},
			},
			Standards: []pacv2beta2.PolicyStandard{},
		},
	}
	for _, o := range opts {
		o(policy)
	}
	return policy
}

func getServer(t *testing.T, clients map[string]client.Client, namespaces map[string][]corev1.Namespace) (pb.CoreServer, error) {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	clientsPool.ClientsReturns(clients)
	clientsPool.ClientStub = func(name string) (client.Client, error) {
		if c, found := clients[name]; found && c != nil {
			return c, nil
		}
		return nil, fmt.Errorf("cluster %s not found", name)
	}
	clustersClient := clustersmngr.NewClient(clientsPool, namespaces, logr.Discard())
	fakeFactory := &clustersmngrfakes.FakeClustersManager{}
	fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)
	fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)

	return createServer(t, serverOptions{
		clustersManager: fakeFactory,
	})
}

func createServer(t *testing.T, o serverOptions) (pb.CoreServer, error) {
	c := o.client
	if c == nil {
		c = createClient(t, o.clusterState...)
	}

	if o.cluster == "" {
		o.cluster = "Default"
	}

	return NewCoreServer(
		CoreServerConfig{
			log:             testr.New(t),
			RestCfg:         &rest.Config{},
			clusterName:     o.cluster,
			ClustersManager: o.clustersManager,
		},
	)
}

type serverOptions struct {
	clusterState          []runtime.Object
	client                client.Client
	namespace             string
	ns                    string
	profileHelmRepository *types.NamespacedName
	clustersManager       clustersmngr.ClustersManager
	capiEnabled           bool
	cluster               string
}
