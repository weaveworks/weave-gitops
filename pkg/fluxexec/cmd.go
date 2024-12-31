package fluxexec

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"

	"github.com/weaveworks/weave-gitops/core/logger"
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

	flux.logger.V(logger.LogLevelInfo).Info(fmt.Sprintf("Running Flux command: %s", cmd.String()))

	return cmd
}

func writeOutput(ctx context.Context, r io.ReadCloser, log logr.Logger) error {
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
			log.Info(strings.TrimSuffix(string(line), "\n"))
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}
	}
}
