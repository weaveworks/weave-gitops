package fluxexec

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func envSlice(environ map[string]string) []string {
	env := []string{}
	for k, v := range environ {
		env = append(env, k+"="+v)
	}

	return env
}

// TODO merge good, or filter out some forbidden env vars
func (flux *Flux) buildEnv(mergeEnv map[string]string) []string {
	return envSlice(mergeEnv)
}

func (flux *Flux) buildFluxCmd(ctx context.Context, mergeEnv map[string]string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, flux.execPath, args...)

	// Use the inherited environment if mergeEnv is nil
	if mergeEnv == nil {
		cmd.Env = os.Environ()
	} else {
		cmd.Env = append(os.Environ(), flux.buildEnv(mergeEnv)...)
	}

	cmd.Dir = flux.workingDir

	if flux.logger != nil {
		flux.logger.Printf("[INFO] running Flux command: %s", cmd.String())
	}

	return cmd
}

func mergeWriters(writers ...io.Writer) io.Writer {
	compact := []io.Writer{}

	for _, w := range writers {
		if w != nil {
			compact = append(compact, w)
		}
	}

	if len(compact) == 0 {
		return ioutil.Discard
	}

	if len(compact) == 1 {
		return compact[0]
	}

	return io.MultiWriter(compact...)
}

func writeOutput(ctx context.Context, r io.ReadCloser, w io.Writer) error {
	// ReadBytes will block until bytes are read, which can cause a delay in
	// returning even if the command's context has been canceled. Use a separate
	// goroutine to prompt ReadBytes to return on cancel
	closeCtx, closeCancel := context.WithCancel(ctx)
	defer closeCancel()

	go func() {
		select {
		case <-ctx.Done():
			r.Close()
		case <-closeCtx.Done():
			return
		}
	}()

	buf := bufio.NewReader(r)

	for {
		line, err := buf.ReadBytes('\n')
		if len(line) > 0 {
			if _, err := w.Write(line); err != nil {
				return err
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}
	}
}
