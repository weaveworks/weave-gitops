package tenancy

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_CreateTenants(t *testing.T) {
	fc := newFakeClient(t)

	tenants, err := Parse("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	err = CreateTenants(context.TODO(), tenants, fc)
	assert.NoError(t, err)

	accounts := corev1.ServiceAccountList{}
	if err := fc.List(context.TODO(), &accounts, client.InNamespace("test-ns1")); err != nil {
		t.Fatal(err)
	}

	expected := []corev1.ServiceAccount{
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test-tenant1",
				Namespace:       "test-ns1",
				ResourceVersion: "1",
				Labels: map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant1",
				},
			},
		},
	}
	assert.Equal(t, expected, accounts.Items)
}

func Test_ExportTenants(t *testing.T) {
	out := &bytes.Buffer{}
	err := ExportTenants("testdata/example.yaml", out)
	assert.NoError(t, err)

	rendered := out.String()
	expected := readGoldenFile(t, "testdata/example.yaml.golden")

	assert.Equal(t, expected, rendered)
}

func TestGenerateTenantResources(t *testing.T) {
	generationTests := []struct {
		name   string
		tenant Tenant
		want   []runtime.Object
	}{
		{
			name: "simple tenant with one namespace",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"test-namespace-1",
				},
			},
			want: []runtime.Object{
				newNamespace("test-namespace-1", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "test-namespace-1", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "test-namespace-1", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "simple tenant with two namespaces",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"test-namespace-1",
					"test-namespace-2",
				},
			},
			want: []runtime.Object{
				newNamespace("test-namespace-1", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "test-namespace-1", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "test-namespace-1", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newNamespace("test-namespace-2", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "test-namespace-2", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "test-namespace-2", "", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
		{
			name: "tenant with custom cluster-role",
			tenant: Tenant{
				Name: "test-tenant",
				Namespaces: []string{
					"test-namespace-1",
				},
				ClusterRole: "demo-cluster-role",
			},
			want: []runtime.Object{
				newNamespace("test-namespace-1", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newServiceAccount("test-tenant", "test-namespace-1", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
				newRoleBinding("test-tenant", "test-namespace-1", "demo-cluster-role", map[string]string{
					"toolkit.fluxcd.io/tenant": "test-tenant",
				}),
			},
		},
	}

	for _, tt := range generationTests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := GenerateTenantResources(tt.tenant)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, resources); diff != "" {
				t.Fatalf("failed to generate resources:\n%s", diff)
			}
		})
	}
}

func TestGenerateTenantResources_WithErrors(t *testing.T) {
	generationTests := []struct {
		name         string
		tenant       Tenant
		errorMessage string
	}{
		{
			name: "simple tenant with no namespace",
			tenant: Tenant{
				Name:       "test-tenant",
				Namespaces: []string{},
			},
			errorMessage: "namespaces required",
		},
		{
			name: "tenant with no name",
			tenant: Tenant{
				Namespaces: []string{
					"test-namespace-1",
				},
			},
			errorMessage: "name is required",
		},
		{
			name: "tenant with no name and no namespace",
			tenant: Tenant{
				Namespaces: []string{},
			},
			errorMessage: "name is required\n\t* namespaces required",
		},
	}

	for _, tt := range generationTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateTenantResources(tt.tenant)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestGenerateTenantResources_WithMultipleTenants(t *testing.T) {
	tenant1 := Tenant{
		Name: "test-tenant1",
		Namespaces: []string{
			"test-namespace-1",
		},
	}
	tenant2 := Tenant{
		Name: "test-tenant2",
		Namespaces: []string{
			"test-namespace-1",
		},
	}

	resourceForTenant1, err := GenerateTenantResources(tenant1)
	assert.NoError(t, err)
	resourceForTenant2, err := GenerateTenantResources(tenant2)
	assert.NoError(t, err)
	resourceForTenants, err := GenerateTenantResources(tenant1, tenant2)
	assert.NoError(t, err)
	assert.Equal(t, append(resourceForTenant1, resourceForTenant2...), resourceForTenants)
}

func TestParse(t *testing.T) {
	tenants, err := Parse("testdata/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(tenants), 2)
	assert.Equal(t, len(tenants[1].Namespaces), 2)
	assert.Equal(t, tenants[1].Namespaces[1], "test-ns3")
}

func Test_newNamespace(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	ns := newNamespace("test-namespace-1", labels)
	assert.Equal(t, ns.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")
}

func Test_newServiceAccount(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	sa := newServiceAccount("test-tenant", "test-namespace", labels)
	assert.Equal(t, sa.Name, "test-tenant")
	assert.Equal(t, sa.Namespace, "test-namespace")
	assert.Equal(t, sa.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")
}

func Test_newRoleBinding(t *testing.T) {
	labels := map[string]string{
		"toolkit.fluxcd.io/tenant": "test-tenant",
	}

	rb := newRoleBinding("test-tenant", "test-namespace", "", labels)
	assert.Equal(t, rb.Name, "test-tenant")
	assert.Equal(t, rb.Namespace, "test-namespace")
	assert.Equal(t, rb.RoleRef.Name, "cluster-admin")
	assert.Equal(t, rb.Labels["toolkit.fluxcd.io/tenant"], "test-tenant")

	rb = newRoleBinding("test-tenant", "test-namespace", "test-cluster-role", labels)
	assert.Equal(t, rb.RoleRef.Name, "test-cluster-role")
}

func readGoldenFile(t *testing.T, filename string) string {
	t.Helper()

	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}
