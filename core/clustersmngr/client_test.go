package clustersmngr_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestClientGet(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"

	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}
	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	k := &kustomizev1.Kustomization{}

	g.Expect(clustersClient.Get(ctx, clusterName, types.NamespacedName{Name: appName, Namespace: ns.Name}, k)).To(Succeed())
	g.Expect(k.Name).To(Equal(appName))
}

func TestClientClusteredList(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)
	namespaced := true

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}
	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	cklist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})

	g.Expect(clustersClient.ClusteredList(ctx, cklist, namespaced)).To(Succeed())

	klist := cklist.Lists()[clusterName][0].(*kustomizev1.KustomizationList)

	g.Expect(klist.Items).To(HaveLen(1))
	g.Expect(klist.Items[0].Name).To(Equal(appName))

	gitRepo := &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: "https://example.com/repo",
			SecretRef: &meta.LocalObjectReference{
				Name: "somesecret",
			},
		},
	}

	g.Expect(k8sEnv.Client.Create(ctx, gitRepo)).To(Succeed())

	cgrlist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.GitRepositoryList{}
	})

	g.Expect(clustersClient.ClusteredList(ctx, cgrlist, namespaced)).To(Succeed())

	glist := cgrlist.Lists()[clusterName][0].(*sourcev1.GitRepositoryList)
	g.Expect(glist.Items).To(HaveLen(1))
	g.Expect(glist.Items[0].Name).To(Equal(appName))
}

func TestClientClusteredListPagination(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	ns1 := createNamespace(g)
	ns2 := createNamespace(g)
	namespaced := true

	clusterName := "mycluster"

	createKust := func(name string, nsName string) {
		kust := &kustomizev1.Kustomization{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsName,
			},
			Spec: kustomizev1.KustomizationSpec{
				SourceRef: kustomizev1.CrossNamespaceSourceReference{
					Kind: sourcev1.GitRepositoryKind,
				},
			},
		}
		ctx := context.Background()
		g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())
	}

	// Create 2 kustomizations in 2 namespaces
	for i := 0; i < 2; i++ {
		appName := "myapp-" + strconv.Itoa(i)
		createKust(appName, ns1.Name)
	}

	for i := 0; i < 1; i++ {
		appName := "myapp-" + strconv.Itoa(i)
		createKust(appName, ns2.Name)
	}

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns1, *ns2},
	}
	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	// First request comes with no continue token
	cklist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})
	g.Expect(clustersClient.ClusteredList(ctx, cklist, namespaced, client.Limit(1), client.Continue(""))).To(Succeed())
	g.Expect(cklist.Lists()[clusterName]).To(HaveLen(2))
	klist := cklist.Lists()[clusterName][0].(*kustomizev1.KustomizationList)
	g.Expect(klist.Items).To(HaveLen(1))

	continueToken := cklist.GetContinue()

	// Second request comes with the continue token
	cklist = clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})
	g.Expect(clustersClient.ClusteredList(ctx, cklist, namespaced, client.Limit(1), client.Continue(continueToken))).To(Succeed())
	g.Expect(cklist.Lists()[clusterName]).To(HaveLen(1))
	klist0 := cklist.Lists()[clusterName][0].(*kustomizev1.KustomizationList)
	g.Expect(klist0.Items).To(HaveLen(1))

	continueToken = cklist.GetContinue()

	// Third request comes with an empty namespaces continue token
	cklist = clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})
	g.Expect(clustersClient.ClusteredList(ctx, cklist, namespaced, client.Limit(1), client.Continue(continueToken))).To(Succeed())
	g.Expect(cklist.Lists()[clusterName]).To(HaveLen(0))
}

func TestClientClusteredListClusterScoped(t *testing.T) {
	g := NewGomegaWithT(t)

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)
	clusterRole := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
			Labels: map[string]string{
				"name": appName,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs:     []string{"get"},
				Resources: []string{"pods"},
			},
		},
	}
	opts := []client.ListOption{
		client.MatchingLabelsSelector{
			Selector: labels.Set(
				map[string]string{
					"name": appName,
				},
			).AsSelector(),
		},
	}

	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, &clusterRole)).To(Succeed())

	cklist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &rbacv1.ClusterRoleList{}
	})

	g.Expect(clustersClient.ClusteredList(ctx, cklist, false, opts...)).To(Succeed())

	klist := cklist.Lists()[clusterName][0].(*rbacv1.ClusterRoleList)

	g.Expect(klist.Items).To(HaveLen(1))
	g.Expect(klist.Items[0].Name).To(Equal(appName))
}

func TestClientCLusteredListErrors(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	cklist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})

	labels := client.MatchingLabels{
		"foo": "@invalid",
	}

	cerr := clustersClient.ClusteredList(context.Background(), cklist, true, labels)
	g.Expect(cerr).ToNot(BeNil())

	var errs clustersmngr.ClusteredListError

	g.Expect(errors.As(cerr, &errs)).To(BeTrue())
}

func TestClientList(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}
	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	list := &kustomizev1.KustomizationList{}

	g.Expect(clustersClient.List(ctx, clusterName, list, client.InNamespace(ns.Name))).To(Succeed())

	g.Expect(list.Items).To(HaveLen(1))
	g.Expect(list.Items[0].Name).To(Equal(appName))
}

func TestClientCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}
	ctx := context.Background()

	g.Expect(clustersClient.Create(ctx, clusterName, kust)).To(Succeed())

	k := &kustomizev1.Kustomization{}
	g.Expect(clustersClient.Get(ctx, clusterName, types.NamespacedName{Name: appName, Namespace: ns.Name}, k)).To(Succeed())
	g.Expect(k.Name).To(Equal(appName))
}

func TestClientDelete(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}
	ctx := context.Background()

	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	g.Expect(clustersClient.Delete(ctx, clusterName, kust)).To(Succeed())
}

func TestClientUpdate(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "/foo",
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}
	ctx := context.Background()
	g.Expect(k8sEnv.Client.Create(ctx, kust)).To(Succeed())

	kust.Spec.Path = "/bar"
	g.Expect(clustersClient.Update(ctx, clusterName, kust)).To(Succeed())

	k := &kustomizev1.Kustomization{}
	g.Expect(k8sEnv.Client.Get(ctx, types.NamespacedName{Name: appName, Namespace: ns.Name}, k)).To(Succeed())
	g.Expect(k.Spec.Path).To(Equal("/bar"))
}

func TestClientPatch(t *testing.T) {
	g := NewGomegaWithT(t)
	ns := createNamespace(g)

	clusterName := "mycluster"
	appName := "myapp" + rand.String(5)

	clientsPool := createClusterClientsPool(g, clusterName)

	nsMap := map[string][]corev1.Namespace{
		clusterName: {*ns},
	}

	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	kust := &kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       kustomizev1.KustomizationKind,
			APIVersion: kustomizev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ns.Name,
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "/foo",
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
			},
		},
	}

	ctx := context.Background()
	opt := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("test"),
	}
	g.Expect(clustersClient.Patch(ctx, clusterName, kust, client.Apply, opt...)).To(Succeed())

	k := &kustomizev1.Kustomization{}
	g.Expect(k8sEnv.Client.Get(ctx, types.NamespacedName{Name: appName, Namespace: ns.Name}, k)).To(Succeed())
	g.Expect(k.Spec.Path).To(Equal("/foo"))
}

func createNamespace(g *GomegaWithT) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	g.Expect(k8sEnv.Client.Create(context.Background(), ns)).To(Succeed())

	return ns
}

func createClusterRoleBinding(g *GomegaWithT, user string) *rbacv1.ClusterRoleBinding {
	crb := &rbacv1.ClusterRoleBinding{}
	crb.Name = "kube-test-" + rand.String(5)
	crb.Subjects = []rbacv1.Subject{
		{
			Kind: "User",
			Name: user,
		},
	}
	crb.RoleRef = rbacv1.RoleRef{
		Kind: "ClusterRole",
		Name: "cluster-admin",
	}

	g.Expect(k8sEnv.Client.Create(context.Background(), crb)).To(Succeed())

	return crb
}

func createClusterClientsPool(g *GomegaWithT, clusterName string) clustersmngr.ClientsPool {
	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsPool := clustersmngr.NewClustersClientsPool()

	config := *k8sEnv.Rest
	config.Timeout = 1 * time.Second
	config.Impersonate = rest.ImpersonationConfig{
		UserName: "anne",
		// Put the user in the `system:masters` group to avoid auth errors
		Groups: []string{"system:masters"},
	}
	client, err := client.New(&config, client.Options{
		Scheme: k8sEnv.Client.Scheme(),
	})
	g.Expect(err).To(BeNil())

	cluster, err := cluster.NewSingleCluster(clusterName, k8sEnv.Rest, scheme)
	g.Expect(err).To(BeNil())
	err = clientsPool.Add(
		client,
		cluster,
	)

	g.Expect(err).To(BeNil())

	return clientsPool
}
