package app

import "fmt"

type InstallParams struct{}

func (a *App) Install(params InstallParams) error {
	return fmt.Errorf("app.Install not implemented")
}
