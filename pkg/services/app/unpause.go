package app

type UnpauseParams struct {
	Name      string
	Namespace string
}

func (a *App) Unpause(params UnpauseParams) error {
	return a.pauseOrUnpause(false, params.Name, params.Namespace)
}
