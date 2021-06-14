package app

import "fmt"

type UninstallParams struct{}

func (a *App) Uninstall(params UninstallParams) error {
	return fmt.Errorf("app.Uninstall not implemented")
}
