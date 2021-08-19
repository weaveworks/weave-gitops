package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

// BlockingCLIAuthHandler takes over the terminal experience and returns a token when the user completes the flow.
type BlockingCLIAuthHandler func(context.Context, io.Writer) (string, error)

type AuthProvider interface {
	DoCLIAuth(ctx context.Context, stdout io.Writer) (string, error)
}

func NewAuthProvider(name gitproviders.GitProviderName) (BlockingCLIAuthHandler, error) {
	switch name {
	case gitproviders.GitProviderGitHub:
		return NewGithubDeviceFlowHandler(http.DefaultClient), nil

	}

	return nil, fmt.Errorf("unsupported auth provider \"%s\"", name)
}
