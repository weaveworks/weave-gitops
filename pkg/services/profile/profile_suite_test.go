package profile_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/profile"
)

var (
	gitClient    *gitfakes.FakeGit
	osysClient   *osysfakes.FakeOsys
	gitProviders *gitprovidersfakes.FakeGitProvider
	log          *loggerfakes.FakeLogger
	profileSvc   *profile.ProfileSvc
)

var _ = BeforeSuite(func() {
	gitClient = &gitfakes.FakeGit{}
	osysClient = &osysfakes.FakeOsys{}
	log = &loggerfakes.FakeLogger{}
	gitProviders = &gitprovidersfakes.FakeGitProvider{}
	profileSvc = profile.NewService(context.TODO(), log, osysClient)

})

func TestProfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profile Suite")
}
