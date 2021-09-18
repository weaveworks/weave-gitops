package app

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type helmClient struct {
}

func (h helmClient) Add(params AddParams) error {
	panic("implement me")
}

func (h helmClient) Get(name types.NamespacedName) (*wego.Application, error) {
	panic("implement me")
}

func (h helmClient) GetCommits(params CommitParams, application *wego.Application) ([]gitprovider.Commit, error) {
	panic("implement me")
}

func (h helmClient) Remove(params RemoveParams) error {
	panic("implement me")
}

func (h helmClient) Status(params StatusParams) (string, string, error) {
	panic("implement me")
}

func (h helmClient) Pause(params PauseParams) error {
	panic("implement me")
}

func (h helmClient) Unpause(params UnpauseParams) error {
	panic("implement me")
}

