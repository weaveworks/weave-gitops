package fluxexec

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func (flux *Flux) wrapExitError(ctx context.Context, err error, stderr string) error {
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		// not an exit error, short circuit, nothing to wrap
		return err
	}

	ctxErr := ctx.Err()

	// nothing to parse, return early
	errString := strings.TrimSpace(stderr)
	if errString == "" {
		return &unwrapper{exitErr, ctxErr}
	}

	// TODO parse stderr for error message here

	return fmt.Errorf("%w\n%s", &unwrapper{exitErr, ctxErr}, stderr)
}

type unwrapper struct {
	err    error
	ctxErr error
}

func (u *unwrapper) Unwrap() error {
	return u.err
}

func (u *unwrapper) Is(target error) bool {
	switch target {
	case context.DeadlineExceeded, context.Canceled:
		return u.ctxErr == context.DeadlineExceeded ||
			u.ctxErr == context.Canceled
	}

	return false
}

func (u *unwrapper) Error() string {
	return u.err.Error()
}
