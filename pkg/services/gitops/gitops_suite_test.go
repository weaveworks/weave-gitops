package gitops_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var (
	fluxClient *fluxfakes.FakeFlux
	kubeClient *kubefakes.FakeKube
	logger     *loggerfakes.FakeLogger

	gitopsSrv gitops.GitopsService
)

func TestGitops(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gitops Suite")
}
