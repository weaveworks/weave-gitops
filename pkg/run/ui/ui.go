package ui

import (
	"fmt"
	"os"

	// "os/signal"

	"strings"

	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/logger"

	"github.com/weaveworks/weave-gitops/pkg/kube"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var Program *tea.Program

type UIModel struct {
	Args           []string
	Flags          RunCommandFlags
	GitopsRunCmd   *cobra.Command
	KubeConfigArgs *genericclioptions.ConfigFlags

	fluxVersion string

	ready bool

	logViewport   viewport.Model
	inputViewport viewport.Model
	textInput     textinput.Model
	err           error

	width  int
	height int

	logs           []string
	additionalLogs []string
}

type RunCommandFlags struct {
	FluxVersion     string
	AllowK8sContext string
	Components      []string
	ComponentsExtra []string
	Timeout         time.Duration
	PortForward     string // port forward specifier, e.g. "port=8080:8080,resource=svc/app"
	DashboardPort   string
	RootDir         string
	// Global flags.
	Namespace  string
	KubeConfig string
	// Flags, created by genericclioptions.
	Context string
}

const (
	dashboardName    = "ww-gitops"
	dashboardPodName = "ww-gitops-weave-gitops"
	adminUsername    = "admin"
)

var watcher *fsnotify.Watcher

// HelmChartversion TODO: update setting HelmChartVersion var everywhere
var HelmChartVersion = "3.0.0"

var (
	kubeClient     *kube.KubeHTTP
	KubeConfigArgs *genericclioptions.ConfigFlags

	rootDir      string
	relativePath string

	log                           logger.Logger
	cancelDevBucketPortForwarding func()
	cancelDashboardPortForwarding func()
	ticker                        *time.Ticker
)

// UI styling
var (
	// add view styles
	logViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3492e5")).
				Foreground(lipgloss.Color("#050e16")).
				BorderForeground(lipgloss.Color("#191919"))
	inputViewportStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#28aec2")).
				Foreground(lipgloss.Color("#050e16")).
				BorderForeground(lipgloss.Color("#191919"))
)

func InitialUIModel() UIModel {
	return UIModel{}
}

type LogMsg struct{ Msg string }
type AdditionalLogMsg struct{ Msg string }

type LogErrMsg struct {
	Err        error
	ShouldExit bool
}

func (m UIModel) Init() tea.Cmd {
	return nil
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd      tea.Cmd
		logVpCmd   tea.Cmd
		inputVpCmd tea.Cmd
	)

	m.textInput, inputVpCmd = m.textInput.Update(msg)
	m.logViewport, logVpCmd = m.logViewport.Update(msg)
	m.inputViewport, inputVpCmd = m.inputViewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlC:
			cmd := func() tea.Msg {
				return tea.Quit()
			}

			return m, tea.Batch(cmd)
		case tea.KeyEnter:
			// enter answer
		}
		// Window size is received when starting up and on every resize
	case tea.WindowSizeMsg:
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.

			logHeight := int(float64(msg.Height) * 0.7)

			m.width = msg.Width
			m.height = msg.Height

			m.logViewport = viewport.New(msg.Width, logHeight)
			m.logViewport.YPosition = 20
			// m.logViewport.HighPerformanceRendering = true

			m.inputViewport = viewport.New(msg.Width, int(float64(msg.Height)*0.3))
			m.inputViewport.YPosition = logHeight
			// m.inputViewport.HighPerformanceRendering = true

			m.logViewport.Style = logViewportStyle
			m.inputViewport.Style = inputViewportStyle

			line := strings.Repeat(" ", msg.Width)

			m.logViewport.SetContent(line + "\n\n\n")
			m.inputViewport.SetContent(line + "\n\n")

			m.ready = true
		} else {
			m.width = msg.Width
			m.height = msg.Height

			m.logViewport.Width = msg.Width
			m.logViewport.Height = int(float64(msg.Height) * 0.7)

			m.inputViewport.Width = msg.Width
			m.inputViewport.Height = int(float64(msg.Height) * 0.3)

			m.logViewport.Style = logViewportStyle
			m.inputViewport.Style = inputViewportStyle

			line := strings.Repeat(" ", msg.Width)

			m.logViewport.SetContent(line + "\n\n\n" + strings.Join(m.logs, "\n"))
			m.inputViewport.SetContent(line + "\n\n" + strings.Join(m.additionalLogs, "\n"))
		}

		// if useHighPerformanceRenderer {
		// Render (or re-render) the whole viewport. Necessary both to
		// initialize the viewport and when the window is resized.
		//
		// This is needed for high-performance rendering only.
		// }
		// return m, nil
		return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd)
		// return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd, viewport.Sync(m.logViewport), viewport.Sync(m.inputViewport))
	case LogMsg:
		tea.Println("LogMsg:")
		tea.Println(msg.Msg)

		m.logs = append(m.logs, msg.Msg)
		line := strings.Repeat(" ", m.width)
		m.logViewport.SetContent(line + "\n\n\n" + strings.Join(m.logs, "\n"))
		m.logViewport.GotoBottom()

		return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd)
		// return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd, viewport.Sync(m.logViewport), viewport.Sync(m.inputViewport))
	case AdditionalLogMsg:
		tea.Println("AdditionalLogMsg:")
		tea.Println(msg.Msg)

		m.additionalLogs = append(m.additionalLogs, msg.Msg)

		line := strings.Repeat(" ", m.width)
		m.inputViewport.SetContent(line + "\n\n" + strings.Join(m.additionalLogs, "\n"))
		m.inputViewport.GotoBottom()

		return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd)
	// return m, tea.Batch(tiCmd, logVpCmd, inputVpCmd, viewport.Sync(m.logViewport), viewport.Sync(m.inputViewport))
	case LogErrMsg:
		fmt.Println(msg.Err.Error())
		if msg.ShouldExit {
			os.Exit(1)
		}
	}

	return m, nil
}

func (m UIModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf(
		"%s\n\n%s",
		m.logViewport.View(),
		m.inputViewport.View(),
	) + "\n\n"

	// return "test string"
}
