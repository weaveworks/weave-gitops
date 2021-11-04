package automation

import (
	"context"
	"testing"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
)

var (
	fluxClient   *fluxfakes.FakeFlux
	gitProviders *gitprovidersfakes.FakeGitProvider
	log          *loggerfakes.FakeLogger

	automationSvc AutomationService
)

var _ = BeforeEach(func() {
	fluxClient = &fluxfakes.FakeFlux{}
	gitProviders = &gitprovidersfakes.FakeGitProvider{
		GetRepoVisibilityStub: func(ctx context.Context, _ gitproviders.RepoURL) (*gitprovider.RepositoryVisibility, error) {
			vis := gitprovider.RepositoryVisibilityPrivate
			return &vis, nil
		},
	}

	log = &loggerfakes.FakeLogger{}
	automationSvc = NewAutomationService(gitProviders, fluxClient, log)
})

func TestAutomation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Automation Suite")
}
