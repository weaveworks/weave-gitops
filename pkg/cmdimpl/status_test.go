package cmdimpl

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run Command Status Test", func() {
	It("Verify status info for kustomize app", func() {

		// Setup
		appName := "kustomize-app"
		clusterName := "mycluster"

		// create temporary directory to use as HOME
		homePath, err := ioutil.TempDir(os.TempDir(), "status-dir")
		Expect(err).Should(Succeed())
		defer Expect(os.RemoveAll(homePath)).Should(Succeed())
		Expect(os.Setenv("HOME", homePath)).Should(Succeed())

		Expect(
			os.MkdirAll(
				filepath.Join(homePath, ".wego", "repositories", clusterName+"-wego", "apps", appName),
				0755,
			),
		).Should(Succeed())

		// flux mocks
		case0 := "get all -n wego-system"
		output0 := `NAME                     	READY	MESSAGE                                                        	REVISION                                     SUSPENDED
gitrepository/wego       	True 	Fetched revision: main/00b92bf6606e040c59404a7257508d65d300bc91	main/00b92bf6606e040c59404a7257508d65d300bc91False
gitrepository/kustomize-app	True 	Fetched revision: main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	main/a2b5b8c0919f405e52619bfc52b5304240d9ef76False

NAME                     	READY	MESSAGE                                                        	REVISION                                     SUSPENDED
kustomization/wego       	True 	Applied revision: main/00b92bf6606e040c59404a7257508d65d300bc91	main/00b92bf6606e040c59404a7257508d65d300bc91False
kustomization/kustomize-app	True 	Applied revision: main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	main/a2b5b8c0919f405e52619bfc52b5304240d9ef76False`

		case1 := "get all -A kustomize-app"
		output1 := `NAMESPACE   	NAME                     	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED
wego-system	gitrepository/kustomize-app	True 	Fetched revision: main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	False

NAMESPACE   	NAME                     	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED
wego-system	kustomization/kustomize-app	True 	Applied revision: main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	False
`

		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				Expect(args).Should(BeElementOf(case0, case1))
				switch args {
				case case0:
					return []byte(output0), nil
				case case1:
					return []byte(output1), nil
				default:
					return nil, fmt.Errorf("arguments not expected %s", args)
				}
			},
		}

		fluxops.SetFluxHandler(fakeHandler)

		// kubectl mocks
		case0Kubectl := `kubectl config current-context`
		case0KubectlOutput := clusterName

		case1KubectlOutput := `status:
  conditions:
  - lastTransitionTime: "2021-05-24T22:48:28Z"`
		case1Kubectl := `kubectl \
			-n wego-system \
			get kustomization/kustomize-app -oyaml`

		_ = override.WithOverrides(func() override.Result {

			expectedOutput := `Latest successful deployment time: 2021-05-24T22:48:28Z
NAMESPACE   	NAME                     	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED
wego-system	gitrepository/kustomize-app	True 	Fetched revision: main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	False

NAMESPACE   	NAME                     	READY	MESSAGE                                                        	REVISION                                     	SUSPENDED
wego-system	kustomization/kustomize-app	True 	Applied revision: main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	main/a2b5b8c0919f405e52619bfc52b5304240d9ef76	False

`

			reader, writer, _ := os.Pipe()
			backup := os.Stdout
			os.Stdout = writer

			// Command parameters
			params := AddParamSet{}
			params.DeploymentType = "kustomize"
			params.Namespace = "wego-system"
			params.Name = appName
			Expect(Status(params)).Should(Succeed())

			os.Stdout = backup

			Expect(writer.Close()).Should(Succeed())
			bts, err := ioutil.ReadAll(reader)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(reader.Close()).Should(Succeed())

			Expect(string(bts)).Should(Equal(expectedOutput))

			fmt.Println(string(bts))

			return override.Result{} //CallCommandSeparatingOutputStreams
		}, utils.OverrideBehavior(utils.CallCommandSeparatingOutputStreamsOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				Expect(args[0].(string)).Should(BeElementOf(case0Kubectl, case1Kubectl))
				switch (args[0]).(string) {
				case case0Kubectl:
					return []byte(case0KubectlOutput), []byte(""), nil
				case case1Kubectl:
					return []byte(case1KubectlOutput), []byte(""), nil
				default:
					return nil, nil, fmt.Errorf("arguments not expected %s", args)
				}

			}),
		)

	})
})

var _ = Describe("GetDeployment tests", func() {
	It("Fail when getting error from flux", func() {

		myAppName := "my-app-name"
		ns := "wego-system"

		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				Expect(args).Should(Equal("get all -n wego-system"))
				return nil, fmt.Errorf("some error")
			},
		}

		fluxops.SetFluxHandler(fakeHandler)

		deploymentType, err := getDeploymentType(ns, myAppName)
		Expect(err).Should(MatchError("some error"))
		Expect(deploymentType).To(BeEmpty())

	})

	It("Fail when flux returns invalid info", func() {

		myAppName := "my-app-name"
		ns := "wego-system"

		fakeHandler := &fluxopsfakes.FakeFluxHandler{
			HandleStub: func(args string) ([]byte, error) {
				Expect(args).Should(Equal("get all -n wego-system"))
				return []byte("wronginfo"), nil
			},
		}

		fluxops.SetFluxHandler(fakeHandler)

		deploymentType, err := getDeploymentType(ns, myAppName)
		Expect(err).Should(MatchError("some error"))
		Expect(deploymentType).To(BeEmpty())

	})
})
