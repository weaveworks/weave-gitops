package bootstrap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type preWizardModel struct {
	windowIsReady bool
	viewport      viewport.Model
	table         table.Model
	textInput     textinput.Model
	msgChan       chan GitProvider
}

type wizardModel struct {
	windowIsReady bool
	viewport      viewport.Model
	inputs        []*bootstrapWizardInput
	msgChan       chan BootstrapCmdOptions
	cursorMode    textinput.CursorMode
	focusIndex    int
	errorMsg      string
}

type checkbox struct {
	checked bool
}

type bootstrapWizardInputType int32

const (
	bootstrapWizardInputTypeTextInput bootstrapWizardInputType = 0
	bootstrapWizardInputTypeCheckbox  bootstrapWizardInputType = 1
)

type bootstrapWizardInput struct {
	inputType     bootstrapWizardInputType
	flagName      string
	prompt        string
	textInput     textinput.Model
	checkboxInput *checkbox
}

const (
	flagSeparator = " - "
	buttonText    = "Submit"
)

// UI styling
var (
	// table
	baseTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	// text inputs
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	// button
	focusedButton = focusedStyle.Render(fmt.Sprintf("[ %s ]", buttonText))
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render(buttonText))
)

func makeViewport(width int, height int, content string) viewport.Model {
	vp := viewport.New(width, height)
	vp.YPosition = 0
	vp.SetContent(content)

	return vp
}

func initialPreWizardModel(msgChan chan GitProvider) preWizardModel {
	columns := []table.Column{
		{Title: "Index", Width: 6},
		{Title: "Git Provider", Width: 20},
	}

	rows := []table.Row{}

	for i, name := range allGitProviderNames {
		rows = append(rows, []string{
			fmt.Sprint(i + 1), name,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(allGitProviders)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.Bold(false).Foreground(lipgloss.NoColor{})
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "Please enter your Git provider index or name from the table"
	ti.Focus()
	ti.CharLimit = 120
	ti.Width = 40
	ti.PromptStyle = focusedStyle
	ti.TextStyle = focusedStyle

	return preWizardModel{
		table:     t,
		textInput: ti,
		msgChan:   msgChan,
	}
}

func (m preWizardModel) Init() tea.Cmd { return textinput.Blink }

func (m preWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			provider := GitProviderUnknown

			indexOrName := strings.ToLower(strings.TrimSpace(m.textInput.Value()))

			for key, value := range allGitProviders {
				strValue := fmt.Sprint(value)

				if indexOrName == strings.ToLower(strValue) || indexOrName == strings.ToLower(key) {
					provider = allGitProviders[key]
					break
				}
			}

			if provider != GitProviderUnknown {
				go func() { m.msgChan <- provider }()

				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		if !m.windowIsReady {
			m.viewport = makeViewport(msg.Width, msg.Height, m.getContent())

			m.windowIsReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	var (
		cmdViewport  tea.Cmd
		cmdTable     tea.Cmd
		cmdTextInput tea.Cmd
	)

	m.table, cmdTable = m.table.Update(msg)
	m.textInput, cmdTextInput = m.textInput.Update(msg)

	m.viewport.SetContent(m.getContent())

	m.viewport, cmdViewport = m.viewport.Update(msg)
	cmds := []tea.Cmd{cmdViewport, cmdTable, cmdTextInput}

	return m, tea.Batch(cmds...)
}

func (m preWizardModel) View() string {
	return m.viewport.View()
}

func (m preWizardModel) getContent() string {
	return fmt.Sprintf(
		"Please enter Git provider index or name and press Enter"+"\n"+
			"(up and down arrows to scroll the view,"+"\n"+
			"Ctrl+C twice to quit):"+"\n%s",
		baseTableStyle.Render(m.table.View())+"\n",
	) + m.textInput.View()
}

func makeInput(task *BootstrapWizardTask, isFocused bool) *bootstrapWizardInput {
	var inputType bootstrapWizardInputType

	if task.isBoolean {
		inputType = bootstrapWizardInputTypeCheckbox
	} else {
		inputType = bootstrapWizardInputTypeTextInput
	}

	flagName := task.flagName

	prompt := task.flagName + flagSeparator + task.flagDescription

	ti := textinput.Model{}

	var cb *checkbox

	if inputType == bootstrapWizardInputTypeCheckbox {
		cb = &checkbox{
			checked: task.flagValue == "true",
		}
	} else {
		ti = textinput.New()
		ti.CursorStyle = cursorStyle
		ti.CharLimit = 100

		ti.SetValue(task.flagValue)
		ti.Placeholder = task.flagDescription

		if task.isPassword {
			ti.EchoMode = textinput.EchoPassword
		}

		if isFocused {
			ti.Focus()
			ti.PromptStyle = focusedStyle
			ti.TextStyle = focusedStyle
		}
	}

	return &bootstrapWizardInput{
		inputType:     inputType,
		flagName:      flagName,
		prompt:        prompt,
		textInput:     ti,
		checkboxInput: cb,
	}
}

func (input *bootstrapWizardInput) getView(isFocused bool) string {
	if input.inputType == bootstrapWizardInputTypeTextInput {
		return input.textInput.View()
	}

	var checkmark string

	if input.checkboxInput.checked {
		checkmark = "x"

		if isFocused {
			checkmark = focusedStyle.Render(checkmark)
		}
	} else {
		checkmark = blurredStyle.Render("_")
	}

	open := "["
	close := "]"

	if isFocused {
		open = focusedStyle.Render(open)
		close = focusedStyle.Render(close)
	}

	return fmt.Sprintf("%s%s%s %s", open, checkmark, close, input.flagName)
}

func initialWizardModel(tasks []*BootstrapWizardTask, msgChan chan BootstrapCmdOptions) wizardModel {
	numInputs := len(tasks)

	inputs := make([]*bootstrapWizardInput, numInputs)

	for i := range inputs {
		inputs[i] = makeInput(tasks[i], i == 0)
	}

	return wizardModel{
		inputs:   inputs,
		errorMsg: "",
		msgChan:  msgChan,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		// Change cursor mode
		case tea.KeyCtrlR:
			m.cursorMode++
			if m.cursorMode > textinput.CursorHide {
				m.cursorMode = textinput.CursorBlink
			}

			cmdsTextInputs := make([]tea.Cmd, len(m.inputs))

			for i, input := range m.inputs {
				if input.inputType == bootstrapWizardInputTypeTextInput {
					cmdsTextInputs[i] = input.textInput.SetCursorMode(m.cursorMode)
				}
			}

			cmds = append(cmds, cmdsTextInputs...)
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter:
			t := msg.Type

			if t == tea.KeyEnter && m.focusIndex == len(m.inputs) {
				options := make(BootstrapCmdOptions)

				for _, input := range m.inputs {
					var value string

					if input.inputType == bootstrapWizardInputTypeTextInput {
						value = strings.TrimSpace(input.textInput.Value())

						if value == "" {
							m.errorMsg = "Missing value in " + input.textInput.Placeholder

							m.viewport.SetContent(m.getContent())

							var cmdViewport tea.Cmd

							m.viewport, cmdViewport = m.viewport.Update(msg)

							return m, cmdViewport
						}
					} else {
						value = strconv.FormatBool(input.checkboxInput.checked)
					}

					options[input.flagName] = value
				}

				go func() { m.msgChan <- options }()

				return m, tea.Quit
			}

			if t == tea.KeyShiftTab {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmdsTextInputs := []tea.Cmd{}

			for i, input := range m.inputs[:len(m.inputs)] {
				if input.inputType == bootstrapWizardInputTypeCheckbox {
					continue
				}

				if i == m.focusIndex {
					cmdsTextInputs = append(cmdsTextInputs, input.textInput.Focus())
					input.textInput.PromptStyle = focusedStyle
					input.textInput.TextStyle = focusedStyle

					continue
				}

				input.textInput.Blur()
				input.textInput.PromptStyle = noStyle
				input.textInput.TextStyle = noStyle
			}

			cmds = append(cmds, cmdsTextInputs...)
		case tea.KeySpace:
			if m.focusIndex == len(m.inputs) {
				return m, nil
			}

			input := m.inputs[m.focusIndex]

			if input.inputType != bootstrapWizardInputTypeCheckbox {
				break
			}

			input.checkboxInput.checked = !input.checkboxInput.checked
		}
	case tea.WindowSizeMsg:
		if !m.windowIsReady {
			m.viewport = makeViewport(msg.Width, msg.Height, m.getContent())

			m.windowIsReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	cmdsTextInputs := m.updateInputs(msg)

	m.viewport.SetContent(m.getContent())

	var cmdViewport tea.Cmd

	m.viewport, cmdViewport = m.viewport.Update(msg)

	cmds = append(cmds, cmdsTextInputs, cmdViewport)

	return m, tea.Batch(cmds...)
}

func (m wizardModel) View() string {
	return m.viewport.View()
}

func (m *wizardModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i, input := range m.inputs {
		if input.inputType == bootstrapWizardInputTypeCheckbox {
			continue
		}

		input.textInput, cmds[i] = input.textInput.Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m wizardModel) getContent() string {
	var b strings.Builder

	b.WriteString("Please enter the following values" + "\n" +
		"(Tab and Shift+Tab to move input selection," + "\n" +
		"(Space to toggle the currently focused checkbox," + "\n" +
		"Enter to move to the next input or submit the form, " + "\n" +
		"up and down arrows to scroll the view, Ctrl+C twice to quit):" + "\n\n\n")

	for i, input := range m.inputs {
		b.WriteString(input.prompt)
		b.WriteRune('\n')
		b.WriteString(input.getView(i == m.focusIndex))

		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}

	fmt.Fprintf(&b, "\n\n%s  %s\n\n", *button, m.errorMsg)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (Ctrl+R to change style)"))

	return b.String()
}
