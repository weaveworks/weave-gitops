package app

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

var syncParams SyncParams

var _ = Describe("Sync", func() {
	var _ = BeforeEach(func() {
		syncParams = SyncParams{
			Name:      "my-app",
			Namespace: "my-namespace",
		}

		kubeClient.GetApplicationReturns(&wego.Application{
			Spec: wego.ApplicationSpec{DeploymentType: wego.DeploymentTypeKustomize, SourceType: wego.SourceTypeGit},
		}, nil)
	})

	It("errors out when cant get application", func() {
		kubeClient.GetApplicationReturns(nil, errors.New("error"))

		err := appSrv.Sync(syncParams)

		Expect(err.Error()).To(HavePrefix("failed getting application"))
	})

	It("sets proper annotation tag to the resource", func() {
		err := appSrv.Sync(syncParams)
		Expect(err).ToNot(HaveOccurred())

		_, resource := kubeClient.SetResourceArgsForCall(0)
		Expect(resource.GetAnnotations()).To(Equal(map[string]string{"reconcile.fluxcd.io/requestedAt": "foo"}))
	})
})
