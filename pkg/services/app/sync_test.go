package app

import (
	"errors"
	"runtime"
	"time"

	"github.com/benbjohnson/clock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

var (
	syncParams SyncParams
	appClock   *clock.Mock
)

var _ = Describe("Sync", func() {
	var _ = BeforeEach(func() {
		appClock = clock.NewMock()
		appSrv.(*AppSvc).Clock = appClock

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

	XIt("sets proper annotation tag to the resource", func() {
		ready := make(chan bool)

		go func() {
			ready <- true
			err := appSrv.Sync(syncParams)
			Expect(err).ToNot(HaveOccurred())
		}()
		runtime.Gosched()
		<-ready

		appClock.Add(10 * time.Second)

		_, resource := kubeClient.SetResourceArgsForCall(0)
		Expect(resource.GetAnnotations()).To(Equal(map[string]string{"reconcile.fluxcd.io/requestedAt": appClock.Now().Format(time.RFC3339Nano)}))
	})
})
