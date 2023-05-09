package run_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta/run"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

func TestCleanupWorksWithNil(t *testing.T) {
	g := NewWithT(t)

	var s run.CleanupFuncs
	f1 := func(ctx context.Context, log logger.Logger) error { return nil }
	f2 := func(ctx context.Context, log logger.Logger) error { return nil }

	s.Push(f1)
	s.Push(nil) // should be ignored
	s.Push(f2)

	fn, err := s.Pop()
	g.Expect(err).NotTo(HaveOccurred(), "pop returned an unexpected error")
	g.Expect(
		reflect.ValueOf(fn).Pointer()).
		To(Equal(reflect.ValueOf(f2).Pointer()), "value returned from stack is not f2")

	fn, err = s.Pop()
	g.Expect(err).NotTo(HaveOccurred(), "pop returned an unexpected error")
	g.Expect(
		reflect.ValueOf(fn).Pointer()).
		To(Equal(reflect.ValueOf(f1).Pointer()), "value returned from stack is not f1")

	fn, err = s.Pop()
	g.Expect(err).To(Equal(run.ErrEmptyStack), "pop returned an unexpected error")
	g.Expect(fn).To(BeNil(), "unexpected value returned from stack")
}

func TestCleanupFailsGracefullyOnConsecutiveCallsToPop(t *testing.T) {
	g := NewWithT(t)

	var s run.CleanupFuncs
	fn, err := s.Pop()
	g.Expect(err).To(Equal(run.ErrEmptyStack), "unexpected error returned from pop")
	g.Expect(fn).To(BeNil(), "unexpected value returned from pop")

	fn, err = s.Pop()
	g.Expect(err).To(Equal(run.ErrEmptyStack), "unexpected error returned from pop")
	g.Expect(fn).To(BeNil(), "unexpected value returned from pop")
}

func TestCleanupClusterRunsAllFunctionsFromStackInCorrectOrder(t *testing.T) {
	g := NewWithT(t)

	cnt := 0
	var s run.CleanupFuncs

	s.Push(func(ctx context.Context, log logger.Logger) error { cnt *= 2; return nil })
	s.Push(func(ctx context.Context, log logger.Logger) error { cnt += 4; return nil })
	s.Push(func(ctx context.Context, log logger.Logger) error { cnt = 3; return nil })

	err := run.CleanupCluster(context.Background(), nil, s)
	g.Expect(err).NotTo(HaveOccurred(), "unexpected error returned")
	g.Expect(cnt).To(Equal(14), "unexpected execution order")
}

func TestCleanupClusterLogsAllErrors(t *testing.T) {
	g := NewWithT(t)

	cnt := 0
	var s run.CleanupFuncs

	s.Push(func(ctx context.Context, log logger.Logger) error { cnt *= 2; return nil })
	s.Push(func(ctx context.Context, log logger.Logger) error { cnt += 4; return nil })
	s.Push(func(ctx context.Context, log logger.Logger) error { return fmt.Errorf("foo") })
	s.Push(func(ctx context.Context, log logger.Logger) error { cnt = 3; return nil })

	var buf strings.Builder
	err := run.CleanupCluster(context.Background(), logger.NewCLILogger(&buf), s)
	g.Expect(err).NotTo(HaveOccurred(), "function should not have returned an error")
	g.Expect(cnt).To(Equal(14), "unexpected execution order")
	g.Expect(buf.String()).To(Equal("✗ failed cleaning up: foo\n"), "unexpected log output")
}
