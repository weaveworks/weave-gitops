package wrapper

import (
	"context"

	gogit "github.com/go-git/go-git/v5"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Git
type Git interface {
	PlainCloneContext(ctx context.Context, path string, isBare bool, o *gogit.CloneOptions) (*gogit.Repository, error)
	PlainOpen(path string) (*gogit.Repository, error)
	PlainInit(path string, isBare bool) (*gogit.Repository, error)
}

type goGit struct{}

var _ Git = new(goGit)

func (g *goGit) PlainCloneContext(ctx context.Context, path string, isBare bool, o *gogit.CloneOptions) (*gogit.Repository, error) {
	return gogit.PlainCloneContext(ctx, path, isBare, o)
}

func (g *goGit) PlainOpen(path string) (*gogit.Repository, error) {
	return gogit.PlainOpen(path)
}

func (g *goGit) PlainInit(path string, isBare bool) (*gogit.Repository, error) {
	return gogit.PlainInit(path, isBare)
}

func NewGoGit() Git {
	return &goGit{}
}
