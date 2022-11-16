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

func (log *UILogger) Write(p []byte) (n int, err error) {
	log.Program.Send(logMsg{msg: string(p)})

	//fmt.Println(">>> " + string(p))

	// os.Stdout.Write(p)

	return len(p), nil
}

type logMsg struct{ msg string }

// type inputLogMsg struct{ msg string }

// type logErrMsg struct {
// 	err        error
// 	shouldExit bool
// }

type UIModel struct {
	uiEvents      chan string
	windowIsReady bool

	// viewports
	rootViewport  viewport.Model
	logViewport   viewport.Model
	inputViewport viewport.Model

	// logs
	log []string
}

// UI styling
var (
	// viewports
	rootViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#d3d8de")).
				Foreground(lipgloss.Color("#ffffff")).
				BorderForeground(lipgloss.Color("#191919"))
	logViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3492e5")).
				Foreground(lipgloss.Color("#050e16")).
				BorderForeground(lipgloss.Color("#191919"))
	inputViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#28aec2")).
				Foreground(lipgloss.Color("#050e16")).
				BorderForeground(lipgloss.Color("#191919"))
)

func makeViewport(width int, height int, content string, style lipgloss.Style) viewport.Model {
	vp := viewport.New(width, height)
	vp.YPosition = 0
	vp.Style = style
	vp.SetContent(content)

	return vp
}

func InitialUIModel(uiEvents chan string) UIModel {
	return UIModel{
		uiEvents: uiEvents,
	}
}

func (m UIModel) Init() tea.Cmd {
	return nil
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		go func() { m.uiEvents <- "test event 2" }()

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlE:
			go func() { m.uiEvents <- "test event 1" }()
		}
	case tea.WindowSizeMsg:
		if !m.windowIsReady {
			line := strings.Repeat(" ", msg.Width)

			m.rootViewport = makeViewport(msg.Width, msg.Height, line+"\n\n\n", rootViewportStyle)

			logHeight := int(float64(msg.Height) * 0.6)

			m.logViewport = makeViewport(msg.Width, logHeight, line+"\n\n\n", logViewportStyle)
			m.logViewport.YPosition = 0

			m.inputViewport = makeViewport(msg.Width, logHeight, line+"\n\n", inputViewportStyle)
			m.inputViewport.YPosition = logHeight

			m.rootViewport.SetContent(m.logViewport.View() + m.inputViewport.View())

			m.windowIsReady = true
		} else {
			m.rootViewport.Width = msg.Width
			m.rootViewport.Height = msg.Height
		}
	case logMsg:
		m.log = append(m.log, msg.msg)

		m.logViewport.SetContent(m.getContent())

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

	// cmds = append(cmds, cmdRootViewport)
	cmds = append(cmds, cmdRootViewport, cmdLogViewport, cmdInputViewport)

	return m, tea.Batch(cmds...)
}

func (m UIModel) View() string {
	if !m.windowIsReady {
		return "\n  Initializing..."
	}

	return m.rootViewport.View()
}

func (m UIModel) getContent() string {
	return strings.Join(m.log, "\n")
}

const Test = 123
