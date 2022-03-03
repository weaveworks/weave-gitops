package app

import (
	"context"
	"testing"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
)

var (
	gitClient    *gitfakes.FakeGit
	fluxClient   *fluxfakes.FakeFlux
	kubeClient   *kubefakes.FakeKube
	osysClient   *osysfakes.FakeOsys
	gitProviders *gitprovidersfakes.FakeGitProvider
	log          *loggerfakes.FakeLogger
	appSrv       AppService
)

var _ = BeforeEach(func() {
	gitClient = &gitfakes.FakeGit{}
	fluxClient = &fluxfakes.FakeFlux{}
	osysClient = &osysfakes.FakeOsys{}
	kubeClient = &kubefakes.FakeKube{
		GetClusterNameStub: func(ctx context.Context) (string, error) {
			return "test-cluster", nil
		},
		GetClusterStatusStub: func(ctx context.Context) kube.ClusterStatus {
			return kube.GitOpsInstalled
		},
	}

	kubeClient.NamespacePresentReturns(true, nil)

	gitProviders = &gitprovidersfakes.FakeGitProvider{
		GetRepoVisibilityStub: func(ctx context.Context, _ gitproviders.RepoURL) (*gitprovider.RepositoryVisibility, error) {
			vis := gitprovider.RepositoryVisibilityPrivate
			return &vis, nil
		},
	}

	log = &loggerfakes.FakeLogger{}
	appSrv = New(context.Background(), log, fluxClient, kubeClient, osysClient)
})

func TestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Suite")
}
