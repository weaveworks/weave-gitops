package app

import (
	"context"
	"testing"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
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

	appSrv AppService
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

	gitProviders = &gitprovidersfakes.FakeGitProvider{
		GetRepoVisibilityStub: func(url string) (*gitprovider.RepositoryVisibility, error) {
			vis := gitprovider.RepositoryVisibilityPrivate
			return &vis, nil
		},

		GetAccountTypeStub: func(owner string) (gitproviders.ProviderAccountType, error) {
			return gitproviders.AccountTypeUser, nil
		},
	}

	appSrv = New(context.Background(), &loggerfakes.FakeLogger{}, gitClient, gitClient, gitProviders, fluxClient, kubeClient, osysClient)
})

func TestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Suite")
}
