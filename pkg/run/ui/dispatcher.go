package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Dispatcher

type UIDispatcher struct {
	Program    *tea.Program
	RunActions chan *RunAction // actions which should be performed by GitOps Run
}

func (disp *UIDispatcher) Write(p []byte) (n int, err error) {
	disp.Program.Send(logMsg{msg: string(p)})

	return len(p), nil
}

func (disp *UIDispatcher) LogPortForwardMessage(msg string) {
	disp.Program.Send(portForwardMsg{msg: msg})
}

func (disp *UIDispatcher) ShowDashboardPrompt() {
	disp.Program.Send(uiActionMsg{actionType: uiActionTypeShowDashboardPrompt})
}

func (disp *UIDispatcher) InstallDashboard() {
	disp.Program.Send(runActionMsg{actionType: RunActionTypeInstallDashboard})
}

func (disp *UIDispatcher) Start() error {
	return disp.Program.Start()
}

func (disp *UIDispatcher) Quit() {
	disp.Program.Send(tea.Quit)
}

// Actions

type RunActionType int32

const RunActionTypeInstallDashboard RunActionType = 0

type RunAction struct {
	actionType RunActionType
}
