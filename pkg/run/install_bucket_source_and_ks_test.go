package run

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		messages, err := findConditionMessages(client, ks)
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

		childDir := filepath.Join(dir, "subdirectory")
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
		str, err := filepath.Abs("../..")
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
