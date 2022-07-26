package nsaccess

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var userName = "test-user"

func TestFilterAccessibleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	testEnv := &envtest.Environment{}
	testEnv.ControlPlane.GetAPIServer().Configure().Append("--authorization-mode=RBAC")

	testCfg, err := testEnv.Start()
	g.Expect(err).NotTo(HaveOccurred())

	defer func() {
		err := testEnv.Stop()
		if err != nil {
			t.Error(err)
		}
	}()

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	adminClient, err := client.New(testCfg, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	list := &corev1.NamespaceList{}
	g.Expect(adminClient.List(ctx, list)).To(Succeed())

	t.Run("returns a namespace that the user has access to", func(t *testing.T) {
		accessibleNS := &corev1.Namespace{}
		accessibleNS.Name = "accessible-ns"
		g.Expect(adminClient.Create(ctx, accessibleNS)).To(Succeed())
		defer removeNs(t, adminClient, accessibleNS)

		inaccessibleNS := &corev1.Namespace{}
		inaccessibleNS.Name = "nope"
		g.Expect(adminClient.Create(ctx, inaccessibleNS)).To(Succeed())
		defer removeNs(t, adminClient, inaccessibleNS)

		roleName := types.NamespacedName{Namespace: accessibleNS.Name, Name: "test-role"}
		rules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{"mygroup"},
				Resources: []string{"coolresource"},
				Verbs:     []string{"get", "list"},
			},
		}

		userCfg := newRestConfigWithRole(t, testCfg, roleName, rules)

		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())

		checker := NewChecker(rules)

		filtered, err := checker.FilterAccessibleNamespaces(ctx, userCfg, list.Items)
		if err != nil {
			t.Error(err)
		}

		if len(filtered) != 1 {
			t.Errorf("expected filtered length to be 1, received %v", len(filtered))
		}

		ok := false
		for _, ns := range filtered {
			if ns.Name == inaccessibleNS.Name {
				t.Error("inaccessible NS should not have appeared")
			}

			if ns.Name == accessibleNS.Name {
				ok = true
			}
		}

		if ok == false {
			t.Error("expected the accessible namespace to exist in the list of filtered namespaces")
		}
	})
	t.Run("filters out namespaces that do not have the right resources", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns := newNamespace(context.Background(), adminClient, NewGomegaWithT(t))
		defer removeNs(t, adminClient, ns)

		roleName := makeRole(ns)

		roleRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				// Don't allow "pods" in the role
				Resources: []string{"secrets", "events", "namespaces"},
				Verbs:     []string{"get", "list"},
			},
		}

		userCfg := newRestConfigWithRole(t, testCfg, roleName, roleRules)

		requiredRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "events", "pods", "namespaces"},
				Verbs:     []string{"get", "list"},
			},
		}

		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())

		checker := NewChecker(requiredRules)
		filtered, err := checker.FilterAccessibleNamespaces(ctx, userCfg, list.Items)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(filtered).To(HaveLen(0))

	})
	t.Run("filters out namespaces that do not have the right verbs", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns := newNamespace(context.Background(), adminClient, NewGomegaWithT(t))
		defer removeNs(t, adminClient, ns)

		roleName := makeRole(ns)

		roleRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "pods", "events", "namespaces"},
				// Don't allow listing "pods" in the role
				Verbs: []string{"get"},
			},
		}

		userCfg := newRestConfigWithRole(t, testCfg, roleName, roleRules)

		requiredRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "events", "pods", "namespaces"},
				Verbs:     []string{"get", "list"},
			},
		}

		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())

		checker := NewChecker(requiredRules)
		filtered, err := checker.FilterAccessibleNamespaces(ctx, userCfg, list.Items)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(filtered).To(HaveLen(0))
	})
	t.Run("works when api groups are defined in multiple roles", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns := newNamespace(context.Background(), adminClient, NewGomegaWithT(t))
		defer removeNs(t, adminClient, ns)

		roleName := makeRole(ns)

		roleRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "events", "namespaces"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		}

		userCfg := newRestConfigWithRole(t, testCfg, roleName, roleRules)

		requiredRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "events", "pods", "namespaces"},
				Verbs:     []string{"get", "list"},
			},
		}

		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())

		checker := NewChecker(requiredRules)
		filtered, err := checker.FilterAccessibleNamespaces(ctx, userCfg, list.Items)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(filtered).To(HaveLen(1))
	})
	t.Run("works when a user has * permissions on a resource", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ns := newNamespace(context.Background(), adminClient, NewGomegaWithT(t))
		defer removeNs(t, adminClient, ns)

		userName = userName + "-" + rand.String(5)

		roleName := makeRole(ns)

		roleRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "events", "namespaces"},
				Verbs:     []string{"*"},
			},
		}

		userCfg := newRestConfigWithRole(t, testCfg, roleName, roleRules)

		requiredRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "events", "namespaces"},
				Verbs:     []string{"get", "list"},
			},
		}

		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())

		checker := NewChecker(requiredRules)
		filtered, err := checker.FilterAccessibleNamespaces(ctx, userCfg, list.Items)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(filtered).To(HaveLen(1))
	})
}

func newNamespace(ctx context.Context, k client.Client, g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k.Create(ctx, ns)).To(Succeed())

	return ns
}

func makeRole(ns *corev1.Namespace) types.NamespacedName {
	return types.NamespacedName{Namespace: ns.Name, Name: fmt.Sprintf("test-role-%v", rand.String(5))}
}

func createRole(t *testing.T, cl client.Client, key types.NamespacedName, rules []rbacv1.PolicyRule) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "test-role", Namespace: key.Namespace},
		Rules:      rules,
	}
	if err := cl.Create(context.TODO(), role); err != nil {
		t.Fatalf("failed to write role: %s", err)
	}

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-role-binding", Namespace: key.Namespace},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     userName,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     role.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if err := cl.Create(context.TODO(), binding); err != nil {
		t.Fatalf("failed to write role-binding: %s", err)
	}
}

func newRestConfigWithRole(t *testing.T, testCfg *rest.Config, roleName types.NamespacedName, rules []rbacv1.PolicyRule) *rest.Config {
	t.Helper()

	scheme, err := kube.CreateScheme()
	if err != nil {
		t.Fatal(err)
	}

	adminClient, err := client.New(testCfg, client.Options{
		Scheme: scheme,
	})

	if err != nil {
		t.Fatal(err)
	}

	createRole(t, adminClient, roleName, rules)

	userCfg := *testCfg

	userCfg.Impersonate = rest.ImpersonationConfig{
		UserName: userName,
	}

	return &userCfg
}

func removeNs(t *testing.T, k client.Client, ns *corev1.Namespace) {
	t.Helper()

	if err := k.Delete(context.Background(), ns); err != nil {
		t.Error(err)
	}
}
