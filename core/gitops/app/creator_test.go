package app

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	testNamespace = "test-system"
)

type creatorFixture struct {
	*GomegaWithT
	env *testutils.K8sTestEnv
}

func (f creatorFixture) cleanUpFixture(t *testing.T) {
	//f.env.Stop()
}

func setUpCreatorTest(t *testing.T) creatorFixture {
	os.Setenv("KUBEBUILDER_ASSETS", "../../../tools/bin/envtest")
	os.Setenv("USE_EXISTING_CLUSTER", "true")
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			testutils.AppCRDsPath,
			testutils.FluxCRDsPath,
		},
		CRDInstallOptions: envtest.CRDInstallOptions{
			CleanUpAfterUse: false,
		},
	}
	_, err := testEnv.Start()

	if err != nil {
		t.Errorf("could not start testEnv: %w", err)
	}

	return creatorFixture{
		//env:         env,
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestNewKubeCreator(t *testing.T) {
	f := setUpCreatorTest(t)

	f.Expect(true).To((BeTrue()))
}

func TestNewKubeCreator2(t *testing.T) {
	f := setUpCreatorTest(t)

	f.Expect(true).To((BeTrue()))
}
