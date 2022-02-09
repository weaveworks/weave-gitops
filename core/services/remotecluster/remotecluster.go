package remotecluster

import (
	"context"
	"errors"

	"k8s.io/client-go/rest"
)

type ConfigGetter interface {
	GetByName(ctx context.Context, name string) (*rest.Config, error)
}

type Configs map[string]*rest.Config

func NewConfigGetter(vals Configs) ConfigGetter {
	return defaultConfigGetter{values: vals}
}

type defaultConfigGetter struct {
	values Configs
}

func (cg defaultConfigGetter) GetByName(ctx context.Context, name string) (*rest.Config, error) {
	cfg, ok := cg.values[name]
	if !ok {
		return nil, errors.New("config for %q not found")
	}

	return cfg, nil
}
