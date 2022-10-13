package bootstrap

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type preWizardModel struct {
	table     table.Model
	textInput textinput.Model
	msgChan   chan GitProvider
	err       error
}

const flagSeparator = " - "

// UI styling
var (
	// table
	baseTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	// text inputs style
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

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
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
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
		case tea.KeyEscape, tea.KeyTab:
			if m.table.Focused() {
				m.table.Blur()
				m.textInput.Focus()
			} else {
				m.table.Focus()
				m.textInput.Blur()
			}
		case tea.KeyEnter:
			provider := GitProviderUnknown

			if m.table.Focused() {
				name := m.table.SelectedRow()[1]
				provider = allGitProviders[name]
			} else {
				indexOrName := strings.ToLower(strings.TrimSpace(m.textInput.Value()))

				for key, value := range allGitProviders {
					strValue := fmt.Sprint(value)

					if indexOrName == strings.ToLower(strValue) || indexOrName == strings.ToLower(key) {
						provider = allGitProviders[key]
						break
					}
				}
			}

			if provider != GitProviderUnknown {
				go func() { m.msgChan <- provider }()

				return m, tea.Quit
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	var (
		cmdTable     tea.Cmd
		cmdTextInput tea.Cmd
	)

	m.table, cmdTable = m.table.Update(msg)
	m.textInput, cmdTextInput = m.textInput.Update(msg)

	return m, tea.Batch([]tea.Cmd{cmdTable, cmdTextInput}...)
}

func (m preWizardModel) View() string {
	return fmt.Sprintf(
		"Please select or enter your Git provider"+"\n"+
			"(up and down arrows to move table selection, Enter to select git provider, "+"\n"+
			"Esc to switch between table and text input, Ctrl + C twice to quit)"+"\n"+
			"\n%s",
		baseTableStyle.Render(m.table.View())+"\n",
	) + fmt.Sprintf(
		"Please enter Git provider index or name and press Enter\n%s",
		m.textInput.View(),
	)
}

type wizardModel struct {
	textInputs []textinput.Model
	prompts    []string
	msgChan    chan BootstrapCmdOptions
	cursorMode textinput.CursorMode
	focusIndex int
	errorMsg   string
}

func makeTextInput(task *BootstrapWizardTask, isFocused bool) textinput.Model {
	ti := textinput.New()
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

	return ti
}

func initialWizardModel(tasks []*BootstrapWizardTask, msgChan chan BootstrapCmdOptions) wizardModel {
	numInputs := len(tasks)

	inputs := make([]textinput.Model, numInputs)

	for i := range inputs {
		task := tasks[i]

		ti := makeTextInput(task, i == 0)

		inputs[i] = ti
	}

	prompts := []string{}

	for _, task := range tasks {
		prompts = append(prompts, task.flagName+flagSeparator+task.flagDescription)
	}

	return wizardModel{
		textInputs: inputs,
		errorMsg:   "",
		prompts:    prompts,
		msgChan:    msgChan,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			cmds := make([]tea.Cmd, len(m.textInputs))

			for i := range m.textInputs {
				cmds[i] = m.textInputs[i].SetCursorMode(m.cursorMode)
			}

			return m, tea.Batch(cmds...)

		// Set focus to next input
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			t := msg.Type

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if t == tea.KeyEnter && m.focusIndex == len(m.textInputs) {
				options := make(BootstrapCmdOptions)

				for i, input := range m.textInputs {
					prompt := m.prompts[i]

					value := strings.TrimSpace(input.Value())

					if value == "" {
						m.errorMsg = "Missing value in " + input.Placeholder
						return m, nil
					}

					options[prompt[:strings.Index(prompt, flagSeparator)]] = value
				}

				go func() { m.msgChan <- options }()

				return m, tea.Quit
			}

			// Cycle indexes
			if t == tea.KeyUp || t == tea.KeyShiftTab {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.textInputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.textInputs)
			}

			cmds := make([]tea.Cmd, len(m.textInputs))

			for i := 0; i <= len(m.textInputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.textInputs[i].Focus()
					m.textInputs[i].PromptStyle = focusedStyle
					m.textInputs[i].TextStyle = focusedStyle

					continue
				}
				// Remove focused state
				m.textInputs[i].Blur()
				m.textInputs[i].PromptStyle = noStyle
				m.textInputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *wizardModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m wizardModel) View() string {
	var b strings.Builder

	b.WriteString("Please enter the following values" + "\n" +
		"(Tab and Shift+Tab to move input selection," + "\n" +
		"Enter to move to the next input or submit the form, " + "\n" +
		"Ctrl + C twice to quit)" + "\n\n\n")

	for i := range m.textInputs {
		b.WriteString(m.prompts[i])
		b.WriteRune('\n')
		b.WriteString(m.textInputs[i].View())

		if i < len(m.textInputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.textInputs) {
		button = &focusedButton
	}

	fmt.Fprintf(&b, "\n\n%s  %s\n\n", *button, m.errorMsg)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (Ctrl+R to change style)"))

	return b.String()
}
