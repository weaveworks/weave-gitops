/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wego

import (
	"context"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	wegov1alpha1 "github.com/weaveworks/weave-gitops/controllers/wego-controller/apis/wego/v1alpha1"
)

var _ = Describe("Application controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		AppName      = "test-application"
		AppNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating an Application", func() {
		It("it creates a GitRespository", func() {
			ctx := context.Background()
			app := &wegov1alpha1.Application{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "wego.weave.works/v1alpha1",
					Kind:       "Application",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      AppName,
					Namespace: AppNamespace,
				},
				Spec: wegov1alpha1.ApplicationSpec{
					URL: "https://github.com/weaveworks/sample-app",
					Reference: &wegov1alpha1.GitRepositoryRef{
						Branch: "main",
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			appLookupKey := types.NamespacedName{Name: AppName, Namespace: AppNamespace}
			createdGitRepository := &sourcev1.GitRepository{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, appLookupKey, createdGitRepository)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdGitRepository.Spec.URL).Should(Equal("https://github.com/weaveworks/sample-app"))
			Expect(createdGitRepository.Spec.Reference.Branch).Should(Equal("main"))
		})
	})
})
