package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

// Actions

type uiActionType int32

const uiActionTypeShowDashboardPrompt uiActionType = 0

type uiPrompt struct {
	textInput textinput.Model
}

// Messages

type logMsg struct {
	msg string
}

type portForwardMsg struct {
	msg string
}

type uiActionMsg struct {
	actionType uiActionType
}

type runActionMsg struct {
	actionType RunActionType
}

// UI Model

type UIModel struct {
	// actions
	runActions chan *RunAction

	// rendering
	windowIsReady bool
	maxWidth      int
	width         int
	height        int

	// viewports
	rootViewport  viewport.Model
	logViewport   viewport.Model
	inputViewport viewport.Model

	// logs
	logs            []string
	portForwardLogs []string

	// prompt
	prompt *uiPrompt
}

// UI styling

const viewportPadding = 1

var (
	// viewports
	rootViewportStyle = lipgloss.NewStyle()
	logViewportStyle  = lipgloss.NewStyle().
				Padding(0, viewportPadding, viewportPadding, viewportPadding).
				BorderStyle(lipgloss.NormalBorder()).
				Align(lipgloss.Center, lipgloss.Top)
	inputViewportStyle = lipgloss.NewStyle().
				Padding(0, viewportPadding, viewportPadding, viewportPadding).
				MarginTop(viewportPadding).
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
	return UIModel{
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
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			m.logViewport.LineUp(1)
		case tea.KeyDown:
			m.logViewport.LineDown(1)
		case tea.KeyEnter:
			if m.prompt != nil {
				value := strings.ToLower(strings.TrimSpace(m.prompt.textInput.Value()))

				if value == "y" || value == "yes" {
					m.installDashboard()
				} else if value == "n" || value == "no" {
					m.prompt = nil
				}
			}
		case tea.KeyEsc:
			if m.prompt != nil {
				m.prompt = nil
			}
		}
	case tea.WindowSizeMsg:
		w := msg.Width
		h := msg.Height

		m.width = w
		m.height = h
		m.maxWidth = w - viewportPadding*4

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
		m.logs = append(m.logs, msg.msg)

		m.logViewport.SetContent(m.getLogViewportContent())

		m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())
	case portForwardMsg:
		m.portForwardLogs = append(m.portForwardLogs, msg.msg)

		m.inputViewport.SetContent(m.getInputViewportContent())

		m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())
	case uiActionMsg:
		switch msg.actionType {
		case uiActionTypeShowDashboardPrompt:
			m.prompt = makeDashboardPrompt()

			m.inputViewport.SetContent(m.getInputViewportContent())

			m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())
		}
	case runActionMsg:
		switch msg.actionType {
		case RunActionTypeInstallDashboard:
			m.installDashboard()
		}
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

func (m UIModel) wrapContentString(content string) string {
	return wrap.String(wordwrap.String(content, m.maxWidth), m.maxWidth)
}

func (m UIModel) getLogViewportContent() string {
	content := strings.Repeat(" ", m.width) + "\n"

	// This wrapping method can be used in conjunction with word-wrapping
	// when word-wrapping is preferred but a line limit has to be enforced.
	content += m.wrapContentString(strings.Join(m.logs, ""))

	return content
}

func (m UIModel) getInputViewportContent() string {
	content := strings.Repeat(" ", m.width) + "\n"

	numPortForwardLogs := len(m.portForwardLogs)

	if numPortForwardLogs > 0 {
		pLogs := make([]string, numPortForwardLogs)

		for i, log := range m.portForwardLogs {
			pLogs[i] = fmt.Sprintf("%d: %s", i+1, log)
		}

		content += "Port forwards for Applications\n" + m.wrapContentString(strings.Join(pLogs, "\n"))
	}

	if m.prompt != nil {
		content += m.prompt.textInput.View()
	}

	return content
}

func (m UIModel) installDashboard() {
	go func() {
		action := &RunAction{
			actionType: RunActionTypeInstallDashboard,
		}

		m.runActions <- action
	}()

	m.prompt = nil
}

func makeDashboardPrompt() *uiPrompt {
	ti := textinput.New()

	ti.CharLimit = 3

	ti.Prompt = "Do you want to install the Weave GitOps Dashboard?"
	ti.SetValue("Y")
	ti.Placeholder = ""

	ti.Focus()

	return &uiPrompt{
		textInput: ti,
	}
}
