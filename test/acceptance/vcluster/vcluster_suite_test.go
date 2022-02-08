package vcluster_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/test/vcluster"
)

func TestVcluster(t *testing.T) {
	if err := vcluster.InstallNginxIngressController(); err != nil {
		t.Errorf("failed installing ingress controller: %w", err)
		t.FailNow()
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Vcluster Suite")
}
