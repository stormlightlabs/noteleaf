package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

type TaskEditOptions struct {
	Output io.Writer
	Input  io.Reader
	Width  int
	Height int
}

type TaskEditor struct {
	task *models.Task
	repo TaskRepository
	opts TaskEditOptions
}

func NewTaskEditor(task *models.Task, repo TaskRepository, opts TaskEditOptions) *TaskEditor {
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
	return &TaskEditor{
		task: task,
		repo: repo,
		opts: opts,
	}
}

type (
	editMode     int
	priorityMode int
)

const (
	fieldNavigation editMode = iota
	statusPicker
	priorityPicker
	textInput
)

const (
	priorityModeText priorityMode = iota
	priorityModeNumeric
	priorityModeLegacy
)

var (
	statusOptions = []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
		models.StatusAbandoned,
	}

	textPriorityOptions = []string{
		"",
		models.PriorityLow,
		models.PriorityMedium,
		models.PriorityHigh,
	}

	numericPriorityOptions = []string{"", "1", "2", "3", "4", "5"}
	legacyPriorityOptions  = []string{"", "A", "B", "C", "D", "E"}
)

type taskEditKeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Enter        key.Binding
	Tab          key.Binding
	ShiftTab     key.Binding
	Escape       key.Binding
	Save         key.Binding
	Cancel       key.Binding
	Help         key.Binding
	StatusEdit   key.Binding
	Priority     key.Binding
	PriorityMode key.Binding
}

func (k taskEditKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Save, k.Cancel, k.Help}
}

func (k taskEditKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.ShiftTab, k.Enter},
		{k.StatusEdit, k.Priority, k.PriorityMode},
		{k.Save, k.Cancel, k.Help},
	}
}

var taskEditKeys = taskEditKeyMap{
	Up:           key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:         key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Left:         key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "previous")),
	Right:        key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next")),
	Enter:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select/edit")),
	Tab:          key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	ShiftTab:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	Escape:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel/back")),
	Save:         key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Cancel:       key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
	Help:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	StatusEdit:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "edit status")),
	Priority:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "edit priority")),
	PriorityMode: key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "priority mode")),
}

type taskEditModel struct {
	task         *models.Task
	originalTask *models.Task
	repo         TaskRepository
	opts         TaskEditOptions
	keys         taskEditKeyMap
	help         help.Model

	mode          editMode
	currentField  int
	statusIndex   int
	priorityIndex int
	priorityMode  priorityMode

	descInput    textinput.Model
	projectInput textinput.Model

	showingHelp bool
	saved       bool
	cancelled   bool

	fields []string
}

func (m taskEditModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m taskEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showingHelp {
			switch {
			case key.Matches(msg, m.keys.Escape) || key.Matches(msg, m.keys.Help):
				m.showingHelp = false
				return m, nil
			}
			return m, nil
		}

		switch m.mode {
		case fieldNavigation:
			return m.updateFieldNavigation(msg)
		case statusPicker:
			return m.updateStatusPicker(msg)
		case priorityPicker:
			return m.updatePriorityPicker(msg)
		case textInput:
			return m.updateTextInput(msg)
		}

	case tea.WindowSizeMsg:
		m.opts.Width = msg.Width
		m.opts.Height = msg.Height
		m.descInput.Width = msg.Width - 20
		m.projectInput.Width = msg.Width - 20
	}

	return m, tea.Batch(cmds...)
}

func (m taskEditModel) updateFieldNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help):
		m.showingHelp = true
		return m, nil
	case key.Matches(msg, m.keys.Cancel):
		m.cancelled = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.Save):
		return m.saveTask()
	case key.Matches(msg, m.keys.Up) || key.Matches(msg, m.keys.ShiftTab):
		m.currentField = (m.currentField - 1 + len(m.fields)) % len(m.fields)
	case key.Matches(msg, m.keys.Down) || key.Matches(msg, m.keys.Tab):
		m.currentField = (m.currentField + 1) % len(m.fields)
	case key.Matches(msg, m.keys.Enter):
		return m.enterField()
	case key.Matches(msg, m.keys.StatusEdit):
		m.mode = statusPicker
		return m, nil
	case key.Matches(msg, m.keys.Priority):
		m.mode = priorityPicker
		return m, nil
	case key.Matches(msg, m.keys.PriorityMode):
		m.priorityMode = (m.priorityMode + 1) % 3
		m.updatePriorityIndex()
		return m, nil
	}
	return m, nil
}

func (m taskEditModel) updateStatusPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = fieldNavigation
	case key.Matches(msg, m.keys.Up) || key.Matches(msg, m.keys.Left):
		m.statusIndex = (m.statusIndex - 1 + len(statusOptions)) % len(statusOptions)
	case key.Matches(msg, m.keys.Down) || key.Matches(msg, m.keys.Right):
		m.statusIndex = (m.statusIndex + 1) % len(statusOptions)
	case key.Matches(msg, m.keys.Enter):
		m.task.Status = statusOptions[m.statusIndex]
		m.mode = fieldNavigation
	}
	return m, nil
}

func (m taskEditModel) updatePriorityPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var options []string
	switch m.priorityMode {
	case priorityModeText:
		options = textPriorityOptions
	case priorityModeNumeric:
		options = numericPriorityOptions
	case priorityModeLegacy:
		options = legacyPriorityOptions
	}

	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = fieldNavigation
	case key.Matches(msg, m.keys.Up) || key.Matches(msg, m.keys.Left):
		m.priorityIndex = (m.priorityIndex - 1 + len(options)) % len(options)
	case key.Matches(msg, m.keys.Down) || key.Matches(msg, m.keys.Right):
		m.priorityIndex = (m.priorityIndex + 1) % len(options)
	case key.Matches(msg, m.keys.Enter):
		m.task.Priority = options[m.priorityIndex]
		m.mode = fieldNavigation
	case key.Matches(msg, m.keys.PriorityMode):
		m.priorityMode = (m.priorityMode + 1) % 3
		m.updatePriorityIndex()
	}
	return m, nil
}

func (m taskEditModel) updateTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = fieldNavigation
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		switch m.fields[m.currentField] {
		case "Description":
			m.task.Description = m.descInput.Value()
		case "Project":
			m.task.Project = m.projectInput.Value()
		}
		m.mode = fieldNavigation
		return m, nil
	}

	switch m.fields[m.currentField] {
	case "Description":
		m.descInput, cmd = m.descInput.Update(msg)
	case "Project":
		m.projectInput, cmd = m.projectInput.Update(msg)
	}

	return m, cmd
}

func (m taskEditModel) enterField() (tea.Model, tea.Cmd) {
	switch m.fields[m.currentField] {
	case "Description":
		m.mode = textInput
		m.descInput.Focus()
		return m, textinput.Blink
	case "Status":
		m.mode = statusPicker
		return m, nil
	case "Priority":
		m.mode = priorityPicker
		return m, nil
	case "Project":
		m.mode = textInput
		m.projectInput.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m *taskEditModel) updatePriorityIndex() {
	var options []string
	switch m.priorityMode {
	case priorityModeText:
		options = textPriorityOptions
	case priorityModeNumeric:
		options = numericPriorityOptions
	case priorityModeLegacy:
		options = legacyPriorityOptions
	}

	for i, opt := range options {
		if opt == m.task.Priority {
			m.priorityIndex = i
			return
		}
	}
	m.priorityIndex = 0
}

func (m taskEditModel) saveTask() (tea.Model, tea.Cmd) {
	m.saved = true
	return m, tea.Quit
}

func (m taskEditModel) View() string {
	if m.showingHelp {
		return m.help.View(m.keys)
	}

	var content strings.Builder

	title := TitleColorStyle.Render("Edit Task")
	content.WriteString(title + "\n\n")

	for i, field := range m.fields {
		fieldStyle := lipgloss.NewStyle()
		if i == m.currentField && m.mode == fieldNavigation {
			fieldStyle = SelectedColorStyle
		}

		switch field {
		case "Description":
			value := m.task.Description
			if m.mode == textInput && i == m.currentField {
				value = m.descInput.View()
			}
			content.WriteString(fieldStyle.Render(fmt.Sprintf("Description: %s", value)) + "\n")

		case "Status":
			statusStr := m.renderStatusField()
			content.WriteString(fieldStyle.Render(fmt.Sprintf("Status: %s", statusStr)) + "\n")

		case "Priority":
			priorityStr := m.renderPriorityField()
			content.WriteString(fieldStyle.Render(fmt.Sprintf("Priority: %s", priorityStr)) + "\n")

		case "Project":
			value := m.task.Project
			if m.mode == textInput && i == m.currentField {
				value = m.projectInput.View()
			}
			content.WriteString(fieldStyle.Render(fmt.Sprintf("Project: %s", value)) + "\n")
		}
		content.WriteString("\n")
	}

	switch m.mode {
	case statusPicker:
		content.WriteString(m.renderStatusPicker())
	case priorityPicker:
		content.WriteString(m.renderPriorityPicker())
	}

	help := m.help.View(m.keys)

	return lipgloss.JoinVertical(lipgloss.Left, content.String(), help)
}

func (m taskEditModel) renderStatusField() string {
	if m.mode == statusPicker {
		return StatusLegend()
	}
	return FormatStatusWithText(m.task.Status)
}

func (m taskEditModel) renderPriorityField() string {
	if m.mode == priorityPicker {
		modeStr := ""
		switch m.priorityMode {
		case priorityModeText:
			modeStr = "Text"
		case priorityModeNumeric:
			modeStr = "Numeric"
		case priorityModeLegacy:
			modeStr = "Legacy"
		}
		return fmt.Sprintf("%s (Mode: %s)", PriorityLegend(), modeStr)
	}
	return FormatPriorityWithText(m.task.Priority)
}

func (m taskEditModel) renderStatusPicker() string {
	var content strings.Builder
	content.WriteString("Select Status:\n")

	for i, status := range statusOptions {
		style := lipgloss.NewStyle()
		if i == m.statusIndex {
			style = SelectedColorStyle
		}

		line := fmt.Sprintf("%s %s", FormatStatusIndicator(status), status)
		content.WriteString(style.Render(line) + "\n")
	}

	return content.String()
}

func (m taskEditModel) renderPriorityPicker() string {
	var content strings.Builder

	modeStr := ""
	var options []string

	switch m.priorityMode {
	case priorityModeText:
		modeStr = "Text"
		options = textPriorityOptions
	case priorityModeNumeric:
		modeStr = "Numeric (1=Low, 5=High)"
		options = numericPriorityOptions
	case priorityModeLegacy:
		modeStr = "Legacy (A=High, E=Low)"
		options = legacyPriorityOptions
	}

	content.WriteString(fmt.Sprintf("Select Priority (%s - Press 'm' to switch modes):\n", modeStr))

	for i, priority := range options {
		style := lipgloss.NewStyle()
		if i == m.priorityIndex {
			style = SelectedColorStyle
		}

		var line string
		if priority == "" {
			line = fmt.Sprintf("%s None", FormatPriorityIndicator(priority))
		} else {
			line = fmt.Sprintf("%s %s - %s", FormatPriorityIndicator(priority), priority, GetPriorityDescription(priority))
		}
		content.WriteString(style.Render(line) + "\n")
	}

	return content.String()
}

func (te *TaskEditor) Edit(ctx context.Context) (*models.Task, error) {
	descInput := textinput.New()
	descInput.SetValue(te.task.Description)
	descInput.Width = te.opts.Width - 20

	projectInput := textinput.New()
	projectInput.SetValue(te.task.Project)
	projectInput.Width = te.opts.Width - 20

	originalTask := *te.task

	statusIndex := 0
	for i, status := range statusOptions {
		if status == te.task.Status {
			statusIndex = i
			break
		}
	}

	priorityMode := priorityModeText
	if te.task.Priority != "" {
		switch GetPriorityDisplayType(te.task.Priority) {
		case "numeric":
			priorityMode = priorityModeNumeric
		case "legacy":
			priorityMode = priorityModeLegacy
		}
	}

	model := taskEditModel{
		task:         te.task,
		originalTask: &originalTask,
		repo:         te.repo,
		opts:         te.opts,
		keys:         taskEditKeys,
		help:         help.New(),

		mode:         fieldNavigation,
		currentField: 0,
		statusIndex:  statusIndex,
		priorityMode: priorityMode,

		descInput:    descInput,
		projectInput: projectInput,

		fields: []string{"Description", "Status", "Priority", "Project"},
	}

	model.updatePriorityIndex()

	program := tea.NewProgram(model, tea.WithInput(te.opts.Input), tea.WithOutput(te.opts.Output))

	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run task editor: %w", err)
	}

	editModel := finalModel.(taskEditModel)

	if editModel.cancelled {
		return nil, fmt.Errorf("edit cancelled")
	}

	if editModel.saved {
		err := te.repo.Update(ctx, te.task)
		if err != nil {
			return nil, fmt.Errorf("failed to save task: %w", err)
		}
	}

	return te.task, nil
}
