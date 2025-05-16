package nsaccess

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func Test_simplerChecker_FilterAccessibleNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	testEnv := &envtest.Environment{}

	testCfg, err := testEnv.Start()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		err := testEnv.Stop()
		if err != nil {
			t.Error(err)
		}
	})

	adminClient, err := client.New(testCfg, client.Options{
		Scheme: scheme.Scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	// The aggregated cluster role controller is not running in the simplified testEnv control plane.
	// Prepare the default admin user-facing role for the following tests.
	cr := &rbacv1.ClusterRole{}
	g.Expect(adminClient.Get(ctx, client.ObjectKey{Name: "admin"}, cr)).To(Succeed())
	g.Expect(cr.Rules).To(BeEmpty())
	cr.Rules = append(cr.Rules, rbacv1.PolicyRule{
		APIGroups: []string{""},
		Resources: []string{"configmaps"},
		Verbs:     []string{"get", "list", "watch"},
	})
	g.Expect(adminClient.Update(ctx, cr)).To(Succeed())

	t.Run("cluster-admin has access to all namespaces", func(t *testing.T) {
		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())
		g.Expect(list.Items).To(Not(BeEmpty()))

		cs, err := kubernetes.NewForConfig(testCfg)
		g.Expect(err).NotTo(HaveOccurred())

		res, err := NewChecker().FilterAccessibleNamespaces(ctx, cs.AuthorizationV1(), list.Items)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res).To(ContainElements(list.Items))
	})

	t.Run("standard user has access to no namespaces", func(t *testing.T) {
		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())
		g.Expect(list.Items).To(Not(BeEmpty()))

		user, err := testEnv.AddUser(envtest.User{Name: "no-access-user"}, nil)
		g.Expect(err).NotTo(HaveOccurred())
		cs, err := kubernetes.NewForConfig(user.Config())
		g.Expect(err).NotTo(HaveOccurred())

		res, err := NewChecker().FilterAccessibleNamespaces(ctx, cs.AuthorizationV1(), list.Items)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res).To(BeEmpty())
	})

	t.Run("namespace owner has access to namespace", func(t *testing.T) {
		username := "test-user"

		ns := &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		g.Expect(adminClient.Create(ctx, ns)).To(Succeed())

		rb := &rbacv1.RoleBinding{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-ns-admins",
				Namespace: ns.Name,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "admin",
			},
			Subjects: []rbacv1.Subject{{
				Kind: "User",
				Name: username,
			}},
		}
		g.Expect(adminClient.Create(ctx, rb)).To(Succeed())

		list := &corev1.NamespaceList{}
		g.Expect(adminClient.List(ctx, list)).To(Succeed())
		g.Expect(list.Items).To(ContainElement(HaveField("Name", ns.Name)))

		user, err := testEnv.AddUser(envtest.User{Name: username}, nil)
		g.Expect(err).NotTo(HaveOccurred())
		cs, err := kubernetes.NewForConfig(user.Config())
		g.Expect(err).NotTo(HaveOccurred())

		res, err := NewChecker().FilterAccessibleNamespaces(ctx, cs.AuthorizationV1(), list.Items)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res).To(ConsistOf(HaveField("Name", ns.Name)))
	})
}
