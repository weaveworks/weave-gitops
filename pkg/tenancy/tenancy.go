package tenancy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const tenantLabel = "toolkit.fluxcd.io/tenant"

var (
	namespaceTypeMeta      = typeMeta("Namespace", "v1")
	serviceAccountTypeMeta = typeMeta("ServiceAccount", "v1")
	roleBindingTypeMeta    = typeMeta("RoleBinding", "rbac.authorization.k8s.io/v1")
)

// Tenant represents a tenant that we generate resources for in the tenancy
// system.
type Tenant struct {
	Name        string   `yaml:"name"`
	Namespaces  []string `yaml:"namespaces"`
	ClusterRole string   `yaml:"clusterRole"`
}

// Validate returns an error if any of the fields isn't valid
func (t Tenant) Validate() error {
	var result error

	if err := validation.IsQualifiedName(t.Name); len(err) > 0 {
		result = multierror.Append(result, fmt.Errorf("invalid tenant name: %s", err))
	}

	if len(t.Namespaces) == 0 {
		result = multierror.Append(result, errors.New("namespaces required"))
	}

	return result
}

// CreateTenants creates resources for tenants given a file for definition.
func CreateTenants(ctx context.Context, tenants []Tenant, c client.Client) error {
	resources, err := GenerateTenantResources(tenants...)
	if err != nil {
		return fmt.Errorf("failed to generate tenant output: %w", err)
	}

	for _, resource := range resources {
		obj := convertToResource(resource)

		err = createObject(ctx, c, obj)
		if err != nil {
			return fmt.Errorf("failed to create resource %s: %w", obj.GetName(), err)
		}
	}

	return nil
}

func createObject(ctx context.Context, c client.Client, obj client.Object) error {
	return c.Create(ctx, obj)
}

func convertToResource(resource runtime.Object) client.Object {
	if resource.GetObjectKind().GroupVersionKind().Kind == "Namespace" {
		return resource.(*corev1.Namespace)
	}

	if resource.GetObjectKind().GroupVersionKind().Kind == "ServiceAccount" {
		return resource.(*corev1.ServiceAccount)
	}

	if resource.GetObjectKind().GroupVersionKind().Kind == "RoleBinding" {
		return resource.(*rbacv1.RoleBinding)
	}

	return nil
}

// ExportTenants exports all the tenants to a file.
func ExportTenants(tenants []Tenant, out io.Writer) error {
	resources, err := GenerateTenantResources(tenants...)
	if err != nil {
		return fmt.Errorf("failed to generate tenant output: %w", err)
	}

	return outputResources(out, resources)
}

func marshalOutput(out io.Writer, output runtime.Object) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	_, err = fmt.Fprintf(out, "%s", data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	return nil
}

func outputResources(out io.Writer, resources []runtime.Object) error {
	for _, v := range resources {
		if err := marshalOutput(out, v); err != nil {
			return fmt.Errorf("failed outputting tenant: %w", err)
		}

		if _, err := out.Write([]byte("---\n")); err != nil {
			return err
		}
	}

	return nil
}

// GenerateTenantResources creates all the resources for tenants.
func GenerateTenantResources(tenants ...Tenant) ([]runtime.Object, error) {
	generated := []runtime.Object{}

	for _, tenant := range tenants {
		if err := tenant.Validate(); err != nil {
			return nil, err
		}
		// TODO: validate tenant name for creation of namespace.
		tenantLabels := map[string]string{
			tenantLabel: tenant.Name,
		}
		for _, namespace := range tenant.Namespaces {
			generated = append(generated, newNamespace(namespace, tenantLabels))
			generated = append(generated, newServiceAccount(tenant.Name, namespace, tenantLabels))
			generated = append(generated, newRoleBinding(tenant.Name, namespace, tenant.ClusterRole, tenantLabels))
		}
	}

	return generated, nil
}

func newNamespace(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: namespaceTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func newServiceAccount(name, namespace string, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: serviceAccountTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func newRoleBinding(name, namespace, clusterRole string, labels map[string]string) *rbacv1.RoleBinding {
	if clusterRole == "" {
		clusterRole = "cluster-admin"
	}

	return &rbacv1.RoleBinding{
		TypeMeta: roleBindingTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     "gotk:" + namespace + ":reconciler",
			},
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

// Parse a raw tenant declaration, and parses it from the YAML and returns the
// extracted Tenants.
func Parse(filename string) ([]Tenant, error) {
	tenantsYAML, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenants file for export: %w", err)
	}

	var tenancy struct {
		Tenants []Tenant `yaml:"tenants"`
	}

	err = yaml.Unmarshal(tenantsYAML, &tenancy)
	if err != nil {
		return nil, err
	}

	return tenancy.Tenants, nil
}
