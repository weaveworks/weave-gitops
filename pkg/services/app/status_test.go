package app

import (
	"context"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/flux"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

var _ = Describe("Run Command Status Test", func() {

	var fluxClient *fluxfakes.FakeFlux
	var kubeClient *kubefakes.FakeKube

	var appSrv AppService
	var params StatusParams

	BeforeEach(func() {
		fluxClient = &fluxfakes.FakeFlux{
			GetDeploymentTypeStub: func(s string, s2 string) (flux.DeploymentType, error) {
				return "kustomize", nil
			},
			GetAllResourcesStatusStub: func(s string) ([]byte, error) {
				return []byte(`NAMESPACE  	NAME               	READY	MESSAGE                                                   	REVISION                                	SUSPENDED
wego-system	helmrepository/my-kustomize-app	True 	Fetched revision: e8edbcf0370b642e41be4de9ff07133cc72914b8	e8edbcf0370b642e41be4de9ff07133cc72914b8	False

NAMESPACE  	NAME            	READY	MESSAGE                         	REVISION	SUSPENDED
wego-system	helmrelease/my-kustomize-app	True 	Release reconciliation succeeded	1.7.3   	False

NAMESPACE  	NAME              	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED
wego-system	kustomization/my-kustomize-app	True 	Applied revision: main/83b287e769f2d6e602480874231e694a70ec9826	main/83b287e769f2d6e602480874231e694a70ec9826	False`), nil
			},
		}

		kubeClient = &kubefakes.FakeKube{
			GetApplicationStub: func(ctx context.Context, name types.NamespacedName) (*wego.Application, error) {
				return &wego.Application{
					Spec: wego.ApplicationSpec{Path: "bar"},
				}, nil
			},
			LatestSuccessfulDeploymentTimeStub: func(context.Context, types.NamespacedName, string) (string, error) {
				return "2021-07-13T17:04:07Z", nil
			},
		}

		appSrv = New(logger.New(os.Stderr), nil, fluxClient, kubeClient, nil)

		params = StatusParams{
			AppName:   "my-kustomize-app",
			Namespace: "wego-system",
		}
	})

	It("Verify status info for kustomize app", func() {

		expected := `Latest successful deployment time: 2021-07-13T17:04:07Z
NAMESPACE  	NAME               	READY	MESSAGE                                                   	REVISION                                	SUSPENDED
wego-system	helmrepository/my-kustomize-app	True 	Fetched revision: e8edbcf0370b642e41be4de9ff07133cc72914b8	e8edbcf0370b642e41be4de9ff07133cc72914b8	False

NAMESPACE  	NAME            	READY	MESSAGE                         	REVISION	SUSPENDED
wego-system	helmrelease/my-kustomize-app	True 	Release reconciliation succeeded	1.7.3   	False

NAMESPACE  	NAME              	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED
wego-system	kustomization/my-kustomize-app	True 	Applied revision: main/83b287e769f2d6e602480874231e694a70ec9826	main/83b287e769f2d6e602480874231e694a70ec9826	False
`

		output := utils.CaptureStdout(func() {
			err := appSrv.Status(params)
			Expect(err).ToNot(HaveOccurred())
		})
		Expect(output).To(Equal(expected))

	})
})
