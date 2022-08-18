package fluxexec

import (
	"context"
	"os/exec"
	"strings"
	"sync"
)

func (flux *Flux) runFluxCmd(ctx context.Context, cmd *exec.Cmd) error {
	var errBuf strings.Builder

	// check for early cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Read stdout / stderr logs from pipe instead of setting cmd.Stdout and
	// cmd.Stderr because it can cause hanging when killing the command
	// https://github.com/golang/go/issues/23019
	stdoutWriter := mergeWriters(cmd.Stdout, flux.stdout)
	stderrWriter := mergeWriters(flux.stderr, &errBuf)

	cmd.Stderr = nil
	cmd.Stdout = nil

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	if err != nil {
		return flux.wrapExitError(ctx, err, "")
	}

	var (
		errStdout, errStderr error
		wg                   sync.WaitGroup
	)

	wg.Add(1)

	go func() {
		defer wg.Done()

		errStdout = writeOutput(ctx, stdoutPipe, stdoutWriter)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		errStderr = writeOutput(ctx, stderrPipe, stderrWriter)
	}()

	// Reads from pipes must be completed before calling cmd.Wait(). Otherwise
	// can cause a race condition
	wg.Wait()

	err = cmd.Wait()
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	if err != nil {
		return flux.wrapExitError(ctx, err, errBuf.String())
	}

	// Return error if there was an issue reading the std out/err
	if errStdout != nil && ctx.Err() != nil {
		return flux.wrapExitError(ctx, errStdout, errBuf.String())
	}

	if errStderr != nil && ctx.Err() != nil {
		return flux.wrapExitError(ctx, errStderr, errBuf.String())
	}

	return nil
}
