package vcluster_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/test/vcluster"
)

func TestVcluster(t *testing.T) {
	if err := vcluster.UpdateHostKubeconfig(); err != nil {
		t.Errorf("failed updating host kubeconfig: %w", err)
		t.FailNow()
	}

	if err := vcluster.WaitClusterConnectivity(); err != nil {
		t.Errorf("failed waiting cluster to be ready: %w", err)
		t.FailNow()
	}

	if err := vcluster.InstallNginxIngressController(); err != nil {
		t.Errorf("failed installing ingress controller: %w", err)
		t.FailNow()
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Vcluster Suite")
}
