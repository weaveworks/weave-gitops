package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/logger"
)

// CleanupFunc is a function supposed to be called while a GitOps Run session
// is terminating. Each component creating resources on the cluster should
// return such a function that is then added to the CleanupFuncs stack by
// the orchestrating code and removed from it and executed during shutdown.
type CleanupFunc func(ctx context.Context, log logger.Logger) error

// CleanupFuncs is a stack holding CleanupFunc references that are used
// to roll up all resources created during an GitOps Run session as soon
// as the session is terminated.
type CleanupFuncs struct {
	fns []CleanupFunc
}

// Push implements the stack's Push operation, adding the given CleanupFunc
// to the top of the stack.
func (c *CleanupFuncs) Push(f CleanupFunc) {
	if f == nil {
		return
	}

	c.fns = append(c.fns, f)
}

var ErrEmptyStack = errors.New("stack is empty")

// Pop implements the stack's Pop operation, returning and removing
// the top CleanupFunc from the stack.
func (c *CleanupFuncs) Pop() (CleanupFunc, error) {
	if len(c.fns) == 0 {
		return nil, ErrEmptyStack
	}
	var res CleanupFunc
	res, c.fns = c.fns[len(c.fns)-1], c.fns[:len(c.fns)-1]
	return res, nil
}

func CleanupCluster(ctx context.Context, log logger.Logger, fns CleanupFuncs) error {
	for {
		fn, err := fns.Pop()
		if err != nil {
			if err != ErrEmptyStack {
				return fmt.Errorf("failed fetching next cleanup function: %w", err)
			}
			break
		}
		if err := fn(ctx, log); err != nil {
			log.Failuref("failed cleaning up: %s", err)
		}
	}

	return nil
}
