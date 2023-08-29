package watch

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/apis/meta"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
)

// mock controller-runtime client
type mockClientForFindConditionMessages struct {
	client.Client
}

// mock client.List
func (c *mockClientForFindConditionMessages) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error { // m
	list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "deployment",
					"namespace": "default",
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "Ready",
							"status":  "False",
							"message": "This is message",
						},
						map[string]interface{}{
							"type":    "Healthy",
							"status":  "True",
							"message": "no error",
						},
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "app2",
					"namespace": "default",
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "Ready",
							"status":  "True",
							"message": "no error",
						},
						map[string]interface{}{
							"type":    "Healthy",
							"status":  "True",
							"message": "no error",
						},
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "app3",
					"namespace": "default",
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":    "Ready",
							"status":  "False",
							"message": "app 3 error",
						},
						map[string]interface{}{
							"type":    "Healthy",
							"status":  "False",
							"message": "time out",
						},
					},
				},
			},
		},
	}

	return nil
}

func deleteObjects(ctx context.Context, o ...client.Object) {
	for _, obj := range o {
		Expect(
			k8sClient.Delete(ctx, obj),
		).To(
			Succeed(), "failed deleting object %s/%s", obj.GetNamespace(), obj.GetName())

		if _, isNamespace := obj.(*corev1.Namespace); isNamespace {
			// Namespaces can't be deleted in envtest:
			// https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
			return
		}

		Eventually(k8sClient.Get, "10s").
			WithArguments(ctx, client.ObjectKeyFromObject(obj), obj).
			Should(WithTransform(apierrors.IsNotFound, BeTrue()),
				"object %s/%s not deleted", obj.GetNamespace(), obj.GetName())
	}
}

var _ = Describe("SetupBucketSourceAndKS", func() {
	log := logger.NewCLILogger(GinkgoWriter)

	It("fails when Namespace isn't set", func() {
		Expect(
			SetupBucketSourceAndKS(context.Background(), log, k8sClient, SetupRunObjectParams{}),
		).To(
			HaveOccurred(), "expected function call to fail")
	})

	It("fails when user lacks permission in Namespace", func() {
		// create a user that has no permission on the cluster
		authUser, err := k8sEnv.Env.AddUser(envtest.User{
			Name: "no-permission",
		}, nil)
		Expect(err).NotTo(HaveOccurred(), "failed creating unauthenticated user")

		scheme, err := kube.CreateScheme()
		Expect(err).NotTo(HaveOccurred())

		unauthClient, err := client.New(authUser.Config(), client.Options{
			Scheme: scheme,
		})
		Expect(err).NotTo(HaveOccurred())

		testNS := corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "setupbucketsourceandks-",
			},
		}
		Expect(k8sClient.Create(context.Background(), &testNS)).To(Succeed(), "failed creating test Namespace")
		defer deleteObjects(context.Background(), &testNS)

		Expect(
			SetupBucketSourceAndKS(context.Background(), log, unauthClient, SetupRunObjectParams{
				Namespace: testNS.Name,
			}),
		).To(
			HaveOccurred(), "expected function call to fail")
	})

	Describe("decryption setup", func() {
		It("fails with non-existent file name", func() {
			testNS := corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					GenerateName: "setupbucketsourceandks-",
				},
			}
			Expect(k8sClient.Create(context.Background(), &testNS)).To(Succeed(), "failed creating test Namespace")
			defer deleteObjects(context.Background(), &testNS)

			Expect(SetupBucketSourceAndKS(context.Background(), log, k8sClient, SetupRunObjectParams{
				Namespace:         testNS.Name,
				DecryptionKeyFile: "/does/not/exist",
			})).To(MatchError(HavePrefix("failed setting up decryption")), "expected a failure")
		})

		It("fails with unknown file extension", func() {
			testNS := corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					GenerateName: "setupbucketsourceandks-",
				},
			}
			Expect(k8sClient.Create(context.Background(), &testNS)).To(Succeed(), "failed creating test Namespace")
			defer deleteObjects(context.Background(), &testNS)

			Expect(SetupBucketSourceAndKS(context.Background(), log, k8sClient, SetupRunObjectParams{
				Namespace:         testNS.Name,
				DecryptionKeyFile: "./testdata/emptyfile",
			})).To(MatchError(ContainSubstring("failed determining decryption key type from filename")), "expected a failure")
		})

		DescribeTable("creates a Secret and configures the Kustomization",
			func(filename, secretKey string) {
				testNS := corev1.Namespace{
					ObjectMeta: v1.ObjectMeta{
						GenerateName: "setupbucketsourceandks-",
					},
				}
				Expect(k8sClient.Create(context.Background(), &testNS)).To(Succeed(), "failed creating test Namespace")
				defer deleteObjects(context.Background(), &testNS)

				Expect(SetupBucketSourceAndKS(context.Background(), log, k8sClient, SetupRunObjectParams{
					Namespace:         testNS.Name,
					DecryptionKeyFile: filename,
				}),
				).To(Succeed(), "function failed unexpectedly")

				var decSecret corev1.Secret
				Expect(
					k8sClient.Get(context.Background(), client.ObjectKey{
						Namespace: testNS.Name,
						Name:      constants.RunDevKsDecryption,
					}, &decSecret),
				).To(Succeed(), "failed to retrieve decryption Secret")

				Expect(decSecret.Data).To(HaveKeyWithValue(secretKey, []byte{}), "unexpected data in decryption Secret")

				var ks kustomizev1.Kustomization
				Expect(
					k8sClient.Get(context.Background(), client.ObjectKey{
						Namespace: testNS.Name,
						Name:      constants.RunDevKsName,
					}, &ks),
				).To(Succeed(), "failed to retrieve Kustomization")

				Expect(ks.Spec.Decryption).ToNot(BeNil(), "decryption spec not set")
				Expect(ks.Spec.Decryption.Provider).To(Equal("sops"), "unexpected decryption provider set")
				Expect(ks.Spec.Decryption.SecretRef).ToNot(BeNil(), "decryption spec missing Secret reference")
				Expect(ks.Spec.Decryption.SecretRef).To(Equal(&meta.LocalObjectReference{
					Name: decSecret.Name,
				}), "decryption spec has invalid Secret reference")
			},
			Entry("with GPG file", "./testdata/emptyfile.asc", "identity.asc"),
			Entry("with age file", "./testdata/emptyfile.agekey", "age.agekey"),
		)
	})

	It("returns no error", func() {
		testNS := &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "setupbucketsourceandks-",
			},
		}
		Expect(k8sClient.Create(context.Background(), testNS)).To(Succeed(), "failed creating test Namespace")
		defer deleteObjects(context.Background(), testNS)

		err := SetupBucketSourceAndKS(context.Background(), log, k8sClient, SetupRunObjectParams{
			Namespace: testNS.Name,
		})
		Expect(err).ToNot(HaveOccurred(), "setting up dev bucket and Kustomization failed")
	})
})

var _ = Describe("findConditionMessages", func() {
	It("returns the condition messages", func() {
		client := &mockClientForFindConditionMessages{}
		ks := &kustomizev1.Kustomization{
			Spec: kustomizev1.KustomizationSpec{},
			Status: kustomizev1.KustomizationStatus{
				Inventory: &kustomizev1.ResourceInventory{
					Entries: []kustomizev1.ResourceRef{
						{
							ID:      "default_deployment_apps_Deployment",
							Version: "v1",
						},
						{
							ID:      "default_app2_apps_Deployment",
							Version: "v1",
						},
						{
							ID:      "default_app3_apps_Deployment",
							Version: "v1",
						},
					},
				},
			},
		}
		messages, err := findConditionMessages(context.Background(), client, ks)
		Expect(err).ToNot(HaveOccurred())
		Expect(messages).To(Equal([]string{
			"Deployment default/deployment: This is message",
			"Deployment default/app3: app 3 error",
			"Deployment default/app3: time out",
		}))
	})
})

var _ = Describe("InitializeTargetDir", func() {
	It("creates a file in an empty directory", func() {
		dir, err := os.MkdirTemp("", "target-dir")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)

		kustomizationPath := filepath.Join(dir, "kustomization.yaml")

		_, err = os.Stat(kustomizationPath)
		Expect(err).To(HaveOccurred()) // File not created yet

		err = InitializeTargetDir(dir)
		Expect(err).ToNot(HaveOccurred())

		fi, err := os.Stat(kustomizationPath)
		Expect(err).ToNot(HaveOccurred())

		err = InitializeTargetDir(dir)
		Expect(err).ToNot(HaveOccurred())

		fi2, err := os.Stat(kustomizationPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(fi2.ModTime()).To(Equal(fi.ModTime())) // File not updated
	})

	It("creates a file in nonexistent directory", func() {
		dir, err := os.MkdirTemp("", "target-dir")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)

		childDir := filepath.Join(dir, "subdirectory", "subsubdirectory")
		_, err = os.Stat(childDir)
		Expect(err).To(HaveOccurred()) // Directory not created yet

		kustomizationPath := filepath.Join(childDir, "kustomization.yaml")

		err = InitializeTargetDir(childDir)
		Expect(err).ToNot(HaveOccurred())

		_, err = os.Stat(kustomizationPath)
		Expect(err).ToNot(HaveOccurred())
	})

	It("throws an error if pointed at a file", func() {
		dir, err := os.MkdirTemp("", "target-dir")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)

		kustomizationPath := filepath.Join(dir, "kustomization.yaml")
		err = InitializeTargetDir(dir)
		Expect(err).ToNot(HaveOccurred())

		err = InitializeTargetDir(kustomizationPath)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("CreateIgnorer", func() {
	It("finds and parses existing gitignore", func() {
		str, err := filepath.Abs("../../..")
		Expect(err).ToNot(HaveOccurred())
		ignorer := CreateIgnorer(str)
		Expect(ignorer.MatchesPath("pkg/server")).To(Equal(false))
		Expect(ignorer.MatchesPath("temp~")).To(Equal(true))
		Expect(ignorer.MatchesPath("bin/gitops")).To(Equal(true))
	})
	It("doesn't mind no gitignore", func() {
		str, err := filepath.Abs(".")
		Expect(err).ToNot(HaveOccurred())
		ignorer := CreateIgnorer(str)
		Expect(ignorer.MatchesPath("bin/gitops")).To(Equal(false))
	})
})
