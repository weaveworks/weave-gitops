package kube_test

import (
	"context"
	"errors"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
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

	It("GetApplications", func() {
		ctx := context.Background()
		name := "my-app"
		app := &wego.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace.Name,
			},
			Spec: wego.ApplicationSpec{
				DeploymentType: wego.DeploymentTypeKustomize,
				SourceType:     wego.SourceTypeGit,
			},
		}

		Expect(k8sClient.Create(ctx, app)).Should(Succeed())

		list, err := k.GetApplications(ctx, namespace.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
		Expect(list[0].Name).To(Equal(name))
	})

	Describe("Apply", func() {
		It("applies a namespaced manifest", func() {
			ctx := context.Background()
			name := "my-app"

			kust := fmt.Sprintf(`
apiVersion: kustomize.toolkit.fluxcd.io/v1beta1
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

			kustObj := &kustomizev1.Kustomization{}
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
})
