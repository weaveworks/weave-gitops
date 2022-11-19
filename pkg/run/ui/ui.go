package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UILogger struct {
	Program *tea.Program
}

type uiActionType int32

const (
	runPrompt uiActionType = 0
)

type uiPrompt struct {
	prompt      string
	placeholder string
	value       string
}

type uiAction struct {
	actionType uiActionType
	prompt     uiPrompt
}

type RunActionType int32

const (
	RunActionTypeInstallDashboard RunActionType = 0
)

type RunAction struct {
	actionType          RunActionType
	shouldPerformAction bool
}

func (log *UILogger) Write(p []byte) (n int, err error) {
	log.Program.Send(logMsg{msg: string(p)})

	return len(p), nil
}

type logMsg struct{ msg string }

type PortForwardMsg struct{ Msg string }

type UIModel struct {
	// actions
	uiActions  []*uiAction     // prompts
	runActions chan *RunAction // actions which should be performed by GitOps Run

	// system
	windowIsReady bool
	width         int
	height        int

	// viewports
	rootViewport  viewport.Model
	logViewport   viewport.Model
	inputViewport viewport.Model

	// logs
	Logs            []string
	portForwardLogs []string
}

// UI styling
var (
	// viewports
	rootViewportStyle = lipgloss.NewStyle()
	logViewportStyle  = lipgloss.NewStyle().
				Padding(1).
				BorderStyle(lipgloss.NormalBorder()).
				Align(lipgloss.Center, lipgloss.Top)
	inputViewportStyle = lipgloss.NewStyle().
				Padding(1).
				MarginTop(1).
				BorderStyle(lipgloss.NormalBorder()).
				Align(lipgloss.Center, lipgloss.Bottom)
)

func makeViewport(width int, height int, style lipgloss.Style) viewport.Model {
	vp := viewport.New(width, height)
	vp.Style = style

	return vp
}

func updateViewportSize(vp viewport.Model, width int, height int) viewport.Model {
	vp.Width = width
	vp.Height = height

	return vp
}

func InitialUIModel(runActions chan *RunAction) UIModel {
	uiActions := []*uiAction{
		{
			actionType: runPrompt,
			prompt: uiPrompt{
				prompt:      "Do you want to install the Weave GitOps Dashboard?",
				placeholder: "",
				value:       "Y",
			},
		},
	}

	return UIModel{
		uiActions:  uiActions,
		runActions: runActions,
	}
}

func (m UIModel) Init() tea.Cmd {
	return nil
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			return m, nil
			// m.logViewport.LineUp(1)
		case tea.KeyDown:
			return m, nil
			// m.logViewport.LineDown(1)
		case tea.KeyCtrlE:
			go func() {
				action := &RunAction{
					actionType:          RunActionTypeInstallDashboard,
					shouldPerformAction: true,
				}

				m.runActions <- action
			}()
		}
	case tea.WindowSizeMsg:
		w := msg.Width
		h := msg.Height

		m.width = w
		m.height = h

		logHeight := int(float64(h) * 0.70)
		inputHeight := int(float64(h) * 0.30)

		if !m.windowIsReady {
			m.logViewport = makeViewport(w, logHeight, logViewportStyle)
			m.logViewport.SetContent(m.getLogViewportContent())

			m.inputViewport = makeViewport(w, inputHeight, inputViewportStyle)
			m.inputViewport.SetContent(m.getInputViewportContent())

			m.rootViewport = makeViewport(w, h, rootViewportStyle)
			m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())

			m.windowIsReady = true
		} else {
			m.logViewport = updateViewportSize(m.logViewport, w, logHeight)
			m.logViewport.SetContent(m.getLogViewportContent())

			m.inputViewport = updateViewportSize(m.inputViewport, w, inputHeight)
			m.inputViewport.SetContent(m.getInputViewportContent())

			m.rootViewport = updateViewportSize(m.rootViewport, w, h)
			m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())
		}
	case logMsg:
		m.Logs = append(m.Logs, msg.msg)

		m.logViewport.SetContent(m.getLogViewportContent())

		m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())
	case PortForwardMsg:
		m.portForwardLogs = append(m.portForwardLogs, msg.Msg)

		m.inputViewport.SetContent(m.getInputViewportContent())

		m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())
	}

	var (
		cmdRootViewport  tea.Cmd
		cmdLogViewport   tea.Cmd
		cmdInputViewport tea.Cmd
	)

	m.rootViewport, cmdRootViewport = m.rootViewport.Update(msg)
	m.logViewport, cmdLogViewport = m.logViewport.Update(msg)
	m.inputViewport, cmdInputViewport = m.inputViewport.Update(msg)

	cmds = append(cmds, cmdRootViewport, cmdLogViewport, cmdInputViewport)

	return m, tea.Batch(cmds...)
}

func (m UIModel) View() string {
	if !m.windowIsReady {
		return "\n  Initializing..."
	}

	return m.rootViewport.View()
}

func (m UIModel) getLogViewportContent() string {
	filler := strings.Repeat(" ", m.width) + "\n\n\n"

	return filler + strings.Join(m.Logs, "\n")
}

func (m UIModel) getInputViewportContent() string {
	filler := strings.Repeat(" ", m.width) + "\n\n"

	return filler + strings.Join(m.portForwardLogs, "\n")
}

const Test = 123
