package run

import (
	"context"
	"errors"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	"github.com/fluxcd/pkg/ssa"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cli-utils/pkg/object"
)

// mock ssa.ResourceManager
type mockResourceManagerForApply struct {
	ResourceManagerForApply
	state stateApply
}

type stateApply string

const (
	stateApplyAllReturnErr   stateApply = "apply-all-return-err"
	stateWaitForSetReturnErr stateApply = "wait-for-set-return-err"
	applyAllErrorMsg                    = "apply all error"
	waitForSetErrorMsg                  = "wait for set error"
)

func (man *mockResourceManagerForApply) ApplyAll(ctx context.Context, objects []*unstructured.Unstructured, opts ssa.ApplyOptions) (*ssa.ChangeSet, error) {
	switch man.state {
	case stateApplyAllReturnErr:
		return nil, errors.New(applyAllErrorMsg)

	default:
		changeSet := ssa.NewChangeSet()

		return changeSet, nil
	}
}

func (man *mockResourceManagerForApply) WaitForSet(set object.ObjMetadataSet, opts ssa.WaitOptions) error {
	switch man.state {
	case stateWaitForSetReturnErr:
		return errors.New(waitForSetErrorMsg)

	default:
		return nil
	}
}

var _ = Describe("apply", func() {
	var fakeLogger *loggerfakes.FakeLogger
	var fakeContext context.Context
	var fakeInstallOptions install.Options
	var fakeManifestsContent []byte

	BeforeEach(func() {
		fakeLogger = &loggerfakes.FakeLogger{}
		fakeContext = context.Background()
		fakeInstallOptions = install.MakeDefaultOptions()

		fakeManifests, err := install.Generate(fakeInstallOptions, "")
		Expect(err).NotTo(HaveOccurred())
		fakeManifestsContent = []byte(fakeManifests.Content)
	})

	It("should apply manifests successfully", func() {
		man := &mockResourceManagerForApply{}

		_, err := apply(fakeLogger, fakeContext, man, fakeManifestsContent)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return an apply all error if the resource manager returns an apply all error", func() {
		man := &mockResourceManagerForApply{state: stateApplyAllReturnErr}

		_, err := apply(fakeLogger, fakeContext, man, fakeManifestsContent)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(applyAllErrorMsg))
	})

	It("should return a wait for set error if the resource manager returns a wait for set error", func() {
		man := &mockResourceManagerForApply{state: stateWaitForSetReturnErr}

		_, err := apply(fakeLogger, fakeContext, man, fakeManifestsContent)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(waitForSetErrorMsg))
	})
})
