package app

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
)

var (
	gitClient    *gitfakes.FakeGit
	fluxClient   *fluxfakes.FakeFlux
	kubeClient   *kubefakes.FakeKube
	osysClient   osys.Osys
	gitProviders *gitprovidersfakes.FakeGitProvider

	appSrv AppService
)

var _ = BeforeEach(func() {
	gitClient = &gitfakes.FakeGit{}
	fluxClient = &fluxfakes.FakeFlux{}
	osysClient = osys.New()
	kubeClient = &kubefakes.FakeKube{
		GetClusterNameStub: func(ctx context.Context) (string, error) {
			return "test-cluster", nil
		},
		GetClusterStatusStub: func(ctx context.Context) kube.ClusterStatus {
			return kube.WeGOInstalled
		},
	}

	gitProviders = &gitprovidersfakes.FakeGitProvider{}

	appSrv = New(logger.NewCLILogger(os.Stderr), gitClient, fluxClient, kubeClient, osysClient)

	appSrv.(*App).gitProviderFactory = func(token string) (gitproviders.GitProvider, error) {
		return gitProviders, nil
	}
})

func TestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Suite")
}
