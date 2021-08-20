package app

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys"
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
			return kube.WeGOInstalled
		},
	}

	gitProviders = &gitprovidersfakes.FakeGitProvider{}
	appSrv = New(&loggerfakes.FakeLogger{}, gitClient, fluxClient, kubeClient, osysClient)

	appSrv.(*App).gitProviderFactory = func(token string) (gitproviders.GitProvider, error) {
		return gitProviders, nil
	}

	appSrv.(*App).temporaryGitClientFactory = func(osysClient osys.Osys, privKeypath string) (git.Git, error) {
		return gitClient, nil
	}
})

func TestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Suite")
}
