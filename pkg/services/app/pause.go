package app

type PauseParams struct {
	Name      string
	Namespace string
}

func (a *App) Pause(params PauseParams) error {
	return a.pauseOrUnpause(true, params.Name, params.Namespace)
}
