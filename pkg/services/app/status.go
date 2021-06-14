package app

import "fmt"

type StatusParams struct{}

func (a *App) Status(params StatusParams) error {
	return fmt.Errorf("app.Status not implemented")
}
