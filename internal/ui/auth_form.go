package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuthFormOptions configures the auth form display
type AuthFormOptions struct {
	Output io.Writer
	Input  io.Reader
	Width  int
	Height int
}

// AuthFormResult holds the submitted credentials
type AuthFormResult struct {
	Handle   string
	Password string
	Canceled bool
}

// AuthForm provides an interactive form for AT Protocol authentication
type AuthForm struct {
	initialHandle string
	opts          AuthFormOptions
}

// NewAuthForm creates a new authentication form
func NewAuthForm(initialHandle string, opts AuthFormOptions) *AuthForm {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.Height == 0 {
		opts.Height = 24
	}

	return &AuthForm{
		initialHandle: initialHandle,
		opts:          opts,
	}
}

type authFormKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Enter    key.Binding
	Submit   key.Binding
	Cancel   key.Binding
}

var authFormKeys = authFormKeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("↑/shift+tab", "previous field")),
	Down:     key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("↓/tab", "next field")),
	Tab:      key.NewBinding(key.WithKeys("tab")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	Submit:   key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "submit")),
	Cancel:   key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc/ctrl+c", "cancel")),
}

type authFormModel struct {
	handleInput   textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	keys          authFormKeyMap
	submitted     bool
	canceled      bool
	handleLocked  bool
}

func (m authFormModel) Init() tea.Cmd {
	if m.handleLocked {
		return m.passwordInput.Focus()
	}
	return m.handleInput.Focus()
}

func (m authFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.canceled = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Submit), key.Matches(msg, m.keys.Enter):
			if m.handleInput.Value() == "" {
				return m, nil
			}
			if m.passwordInput.Value() == "" {
				return m, nil
			}
			m.submitted = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.Tab):
			m.nextInput()
			cmds = append(cmds, m.updateFocus())
		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.ShiftTab):
			m.prevInput()
			cmds = append(cmds, m.updateFocus())
		}
	}

	cmd := m.updateInputs(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *authFormModel) nextInput() {
	if m.handleLocked {
		m.focusIndex = 1
	} else {
		m.focusIndex = (m.focusIndex + 1) % 2
	}
}

func (m *authFormModel) prevInput() {
	if m.handleLocked {
		m.focusIndex = 1
	} else {
		m.focusIndex = (m.focusIndex - 1 + 2) % 2
	}
}

func (m *authFormModel) updateFocus() tea.Cmd {
	if m.focusIndex == 0 && !m.handleLocked {
		m.handleInput.Focus()
		m.passwordInput.Blur()
		return textinput.Blink
	}

	m.handleInput.Blur()
	m.passwordInput.Focus()
	return textinput.Blink
}

func (m *authFormModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if !m.handleLocked {
		m.handleInput, cmd = m.handleInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.passwordInput, cmd = m.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m authFormModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9"))

	b.WriteString(titleStyle.Render("AT Protocol Authentication"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("BlueSky Handle:"))
	b.WriteString("\n")
	if m.handleLocked {
		lockedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
		b.WriteString(lockedStyle.Render(m.handleInput.Value()))
		b.WriteString(lockedStyle.Render(" (locked)"))
	} else {
		b.WriteString(m.handleInput.View())
	}
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("App Password:"))
	b.WriteString("\n")
	b.WriteString(m.passwordInput.View())
	b.WriteString("\n\n")

	if m.handleInput.Value() == "" {
		b.WriteString(errorStyle.Render("Handle is required"))
		b.WriteString("\n")
	}
	if m.passwordInput.Value() == "" {
		b.WriteString(errorStyle.Render("Password is required"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	helpText := "tab/shift+tab: navigate • enter/ctrl+s: submit • esc/ctrl+c: cancel"
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

// Run displays the auth form and returns the entered credentials
func (af *AuthForm) Run() (*AuthFormResult, error) {
	handleInput := textinput.New()
	handleInput.Placeholder = "username.bsky.social"
	handleInput.Width = 40
	handleInput.CharLimit = 253

	passwordInput := textinput.New()
	passwordInput.Placeholder = "App password"
	passwordInput.Width = 40
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'

	handleLocked := false
	focusIndex := 0

	if af.initialHandle != "" {
		handleInput.SetValue(af.initialHandle)
		handleLocked = true
		focusIndex = 1
	}

	model := authFormModel{
		handleInput:   handleInput,
		passwordInput: passwordInput,
		focusIndex:    focusIndex,
		keys:          authFormKeys,
		handleLocked:  handleLocked,
	}

	program := tea.NewProgram(
		model,
		tea.WithInput(af.opts.Input),
		tea.WithOutput(af.opts.Output),
	)

	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run auth form: %w", err)
	}

	result := finalModel.(authFormModel)

	return &AuthFormResult{
		Handle:   result.handleInput.Value(),
		Password: result.passwordInput.Value(),
		Canceled: result.canceled,
	}, nil
}
