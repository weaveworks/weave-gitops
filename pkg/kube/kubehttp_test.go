package kube_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
	kyaml "sigs.k8s.io/yaml"
)

var _ = Describe("KubeHTTP", func() {
	var (
		namespace *corev1.Namespace
		err       error
	)

	var _ = BeforeEach(func() {
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		err = k8sClient.Create(context.Background(), namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")

		k = &kube.KubeHTTP{
			Client:      k8sTestEnv.Client,
			DynClient:   k8sTestEnv.DynClient,
			RestMapper:  k8sTestEnv.RestMapper,
			ClusterName: testClustername,
		}
	})

	AfterEach(func() {
		err = k8sClient.Delete(context.Background(), namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
	})

	It("GetClusterName", func() {
		name, err := k.GetClusterName(context.Background())
		Expect(err).NotTo(HaveOccurred())

		Expect(name).To(Equal(testClustername))
	})

	It("GetClusterStatus", func() {
		ctx := context.Background()
		status := k.GetClusterStatus(ctx)

		// To determine cluster status, we check for the wego CRD.
		// We cannot remove that CRD for tests, so we can only test this
		// cluster state.
		Expect(status.String()).To(Equal(kube.GitOpsInstalled.String()))
	})

	It("FluxPresent", func() {
		ctx := context.Background()

		exists1, err := k.FluxPresent(ctx)
		Expect(err).NotTo(HaveOccurred())

		// Flux doesn't exist yet
		Expect(exists1).To(BeFalse())

		fluxNs := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: kube.FluxNamespace,
			},
		}

		// Create the namespace
		err = k8sClient.Create(ctx, &fluxNs)
		Expect(err).NotTo(HaveOccurred())

		exists2, err := k.FluxPresent(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(exists2).To(BeTrue())
	})

	It("NamespacePresent", func() {
		ctx := context.Background()
		namespace := "wego-system"

		exists1, err := k.NamespacePresent(ctx, namespace)
		Expect(err).NotTo(HaveOccurred())

		// Namespace doesn't exist yet
		Expect(exists1).To(BeFalse())

		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		// Create the namespace
		err = k8sClient.Create(ctx, &ns)
		Expect(err).NotTo(HaveOccurred())

		exists2, err := k.NamespacePresent(ctx, namespace)
		Expect(err).NotTo(HaveOccurred())

		Expect(exists2).To(BeTrue())
	})

	It("GetApplication", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace.Name,
			},
			Spec: wego.ApplicationSpec{
				SourceType:     wego.SourceTypeGit,
				DeploymentType: wego.DeploymentTypeKustomize,
			},
		}

		Expect(k8sClient.Create(ctx, app)).Should(Succeed())

		a, err := k.GetApplication(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name})
		Expect(err).NotTo(HaveOccurred())
		Expect(a.Name).To(Equal(name))
	})

	It("SecretPresent", func() {
		name := "my-secret"
		ctx := context.Background()
		secret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace.Name},
		}

		err = k8sClient.Create(ctx, &secret)
		Expect(err).NotTo(HaveOccurred())

		exists, err := k.SecretPresent(ctx, name, namespace.Name)
		Expect(err).NotTo(HaveOccurred())

		Expect(exists).To(BeTrue())
	})

	Describe("Apply", func() {
		It("applies a namespaced manifest", func() {
			ctx := context.Background()
			name := "my-app"

			kust := fmt.Sprintf(`
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  name: %s
  namespace: %s
spec:
  interval: 1m0s
  prune: true
  validation: client
  sourceRef:
    name: foo
    kind: GitRepository
`, name, namespace.Name)
			Expect(k.Apply(ctx, []byte(kust), namespace.Name)).Should(Succeed())

			kustObj := &kustomizev2.Kustomization{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name}, kustObj)
			Expect(err).NotTo(HaveOccurred())

			Expect(kustObj.Name).To(Equal(name))
		})

		It("applies a cluster wide manifest", func() {
			ctx := context.Background()
			namespace := `
apiVersion: v1
kind: Namespace
metadata:
  name: foo
`

			Expect(k.Apply(ctx, []byte(namespace), "")).Should(Succeed())
		})
		It("fails to apply invalid manifest", func() {
			ctx := context.Background()

			kust := "invalid yaml"

			err := k.Apply(ctx, []byte(kust), namespace.Name)

			Expect(errors.Unwrap(err).Error()).Should(ContainSubstring("failed decoding manifest"))
		})
	})

	Describe("Delete", func() {
		It("delete a manifest", func() {
			ctx := context.Background()
			name := "my-app"

			app := &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					Branch:         "master",
					Path:           "/.kustomize",
					DeploymentType: wego.DeploymentTypeKustomize,
					SourceType:     wego.SourceTypeGit,
				},
			}
			appYaml := fmt.Sprintf(`
apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: %s
  spec:
    branch: master
    deployment_type: kustomize
    path: ./kustomize
    source_type: git
`, name, namespace.Name)

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			Expect(k.Delete(ctx, []byte(appYaml))).Should(Succeed())
		})

		It("delete an invalid manifest", func() {
			ctx := context.Background()
			appYaml := "invalid"

			err := k.Delete(ctx, []byte(appYaml))
			Expect(errors.Unwrap(err).Error()).Should(ContainSubstring("failed decoding manifest"))
		})
	})

	It("DeleteByName", func() {
		ctx := context.Background()
		name := "my-app"

		app := &wego.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace.Name,
			},
			Spec: wego.ApplicationSpec{
				Branch:         "master",
				Path:           "/.kustomize",
				DeploymentType: wego.DeploymentTypeKustomize,
				SourceType:     wego.SourceTypeGit,
			},
		}

		Expect(k8sClient.Create(ctx, app)).Should(Succeed())

		Expect(k.DeleteByName(ctx, name, kube.GVRApp, namespace.Name)).Should(Succeed())

		a, err := k.GetApplication(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name})
		Expect(err).ToNot(HaveOccurred())
		Expect(a.DeletionTimestamp).ToNot(BeNil())
	})
	Describe("Getting client with override kubeconfig", func() {
		It("valid kubeconfig", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
			origkc := os.Getenv("KUBECONFIG")
			defer os.Setenv("KUBECONFIG", origkc)
			dir, err := ioutil.TempDir("", "kube-http-test-")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)
			tests := []struct {
				name        string
				expName     string
				clusterName string
			}{
				{"foo", "foo", "foo"},
				{"weave-upa-admin@weave-upa", "weave-upa", "weave-upa"},
				{"user@weave.works@market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo", "user@weave.works@market.eu-west-2.eksctl.io-podinfo"},
				{"user@market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo", "user@market.eu-west-2.eksctl.io-podinfo"},
				{"user@market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo", "market.eu-west-2.eksctl.io-podinfo"},
				{"cluster_name", "cluster-name", "cluster_name"},
			}
			for _, test := range tests {
				createKubeconfig(test.name, test.clusterName, dir, true)
				_, cname, err := kube.RestConfig()
				Expect(err).ToNot(HaveOccurred(), "Failed to get a kube config")

				Expect(cname).To(Equal(test.expName))
			}

		})
		It("errors when pointing at a missing kubeconfig file", func() {
			t, err := ioutil.TempFile("", "not_a_kubeconfig")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(t.Name())
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
			origkc := os.Getenv("KUBECONFIG")
			defer os.Setenv("KUBECONFIG", origkc)
			os.Setenv("KUBECONFIG", t.Name())
			_, _, err = kube.RestConfig()
			Expect(err).To(HaveOccurred(), "Should receive an error about invalid kubeconfig ")

		})
		It("errors when no current_context set in kubeconfig file", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
			origkc := os.Getenv("KUBECONFIG")
			defer os.Setenv("KUBECONFIG", origkc)
			dir, err := ioutil.TempDir("", "kube-http-test-")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)
			createKubeconfig("foo", "bar", dir, false)
			_, _, err = kube.RestConfig()
			Expect(err).To(HaveOccurred(), "Should receive an error about no current context ")
		})
		It("returns a sensisble clusterName inCluster", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, nil }
			_, clusterName, err := kube.RestConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterName).To(Equal("default"))
		})
		It("derives the clusterName from the env inCluster", func() {
			kube.InClusterConfig = func() (*rest.Config, error) { return nil, nil }
			origcn := os.Getenv("CLUSTER_NAME")
			defer os.Setenv("CLUSTER_NAME", origcn)
			os.Setenv("CLUSTER_NAME", "foo")
			_, clusterName, err := kube.RestConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterName).To(Equal("foo"))
		})
	})

	Describe("SetResouce", func() {
		It("sets a k8s resource", func() {
			ctx := context.Background()
			name := "my-app"
			app := &wego.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace.Name,
				},
				Spec: wego.ApplicationSpec{
					SourceType:     wego.SourceTypeGit,
					DeploymentType: wego.DeploymentTypeKustomize,
				},
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			resource := &wego.Application{}

			err := k.GetResource(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name}, resource)
			Expect(err).NotTo(HaveOccurred())

			resource.SetAnnotations(map[string]string{
				"my-annotation": "note",
			})

			err = k.SetResource(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			newResource := &wego.Application{}

			err = k.GetResource(ctx, types.NamespacedName{Name: name, Namespace: namespace.Name}, newResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(newResource.GetAnnotations()["my-annotation"]).To(Equal("note"))
		})
	})

	Describe("GetWegoConfig", func() {
		It("get a wego config in a namespace", func() {
			ctx := context.Background()
			name := types.NamespacedName{Name: "weave-gitops-config", Namespace: namespace.Name}
			cm := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name.Name,
					Namespace: name.Namespace,
				},
				Data: map[string]string{
					"config": "FluxNamespace: flux-system",
				},
			}

			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())

			wegoConfig, err := k.GetWegoConfig(ctx, name.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(wegoConfig.FluxNamespace).To(Equal("flux-system"))
		})

		It("get the first wego config in all namespacee", func() {
			ctx := context.Background()

			name := types.NamespacedName{Name: "weave-gitops-config", Namespace: namespace.Name}
			cm := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name.Name,
					Namespace: name.Namespace,
				},
				Data: map[string]string{
					"config": "FluxNamespace: flux-system",
				},
			}

			Expect(k8sClient.Create(ctx, cm)).Should(Succeed())

			wegoConfig, err := k.GetWegoConfig(ctx, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(wegoConfig.FluxNamespace).To(Equal("flux-system"))
		})

		It("fails getting config map", func() {
			_, err := k.GetWegoConfig(context.Background(), "foo")
			Expect(err.Error()).To(ContainSubstring("Wego Config not found"))
		})
	})

	Describe("SetWegoConfig", func() {
		It("set a wego config in a namespace", func() {
			ctx := context.Background()

			cm, err := k.SetWegoConfig(ctx, kube.WegoConfig{FluxNamespace: "foo"}, namespace.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(cm).ToNot(BeNil())

			wegoConfig, err := k.GetWegoConfig(ctx, namespace.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(wegoConfig.FluxNamespace).To(Equal("foo"))
		})

		It("fails setting a wego config", func() {
			ctx := context.Background()

			cm, err := k.SetWegoConfig(ctx, kube.WegoConfig{FluxNamespace: "foo"}, "")
			Expect(err.Error()).To(ContainSubstring("failed getting weave-gitops configmap"))
			Expect(cm).To(BeNil())
		})
	})

})

func createKubeconfig(name, clusterName, dir string, setCurContext bool) {
	f, err := ioutil.TempFile(dir, "test.kubeconfig")
	Expect(err).ToNot(HaveOccurred())

	c := clientcmdapi.Config{}

	if setCurContext {
		c.CurrentContext = name
	}

	c.APIVersion = "v1"
	c.Kind = "Config"
	c.Contexts = append(c.Contexts, clientcmdapi.NamedContext{Name: name, Context: clientcmdapi.Context{Cluster: clusterName}})
	c.Clusters = append(c.Clusters, clientcmdapi.NamedCluster{Name: clusterName, Cluster: clientcmdapi.Cluster{Server: "127.0.0.1"}})
	kubeconfig, err := kyaml.Marshal(&c)
	Expect(err).ToNot(HaveOccurred())
	_, err = f.Write(kubeconfig)
	Expect(err).ToNot(HaveOccurred())
	f.Close()
	os.Setenv("KUBECONFIG", f.Name())
}
