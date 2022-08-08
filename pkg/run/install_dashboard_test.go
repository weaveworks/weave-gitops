package run_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

const (
	secret    = "some-secret"
	namespace = "some-namespace"
)

var _ = Describe("InstallDashboard", func() {
	var fakeLogger *loggerfakes.FakeLogger
	var fakeContext context.Context

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
		fakeContext = context.Background()
	})

	It("should install dashboard successfully", func() {
		man := &mockResourceManagerForApply{}

		err := run.InstallDashboard(fakeLogger, fakeContext, man, namespace, secret)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return an apply all error if the resource manager returns an apply all error", func() {
		man := &mockResourceManagerForApply{state: stateApplyAllReturnErr}

		err := run.InstallDashboard(fakeLogger, fakeContext, man, namespace, secret)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(applyAllErrorMsg))
	})
})
