package internal

import (
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
)

type GetAuthHandler func(gitproviders.GitProviderName) (auth.BlockingCLIAuthHandler, error)
