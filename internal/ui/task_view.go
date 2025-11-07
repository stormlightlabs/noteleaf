package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// TaskViewOptions configures the task view UI behavior
type TaskViewOptions struct {
	// Output destination (stdout for interactive, buffer for testing)
	Output io.Writer
	// Input source (stdin for interactive, strings reader for testing)
	Input io.Reader
	// Enable static mode (no interactive components)
	Static bool
	// Width and height for viewport sizing
	Width  int
	Height int
}

// TaskView handles task detail viewing UI
type TaskView struct {
	task *models.Task
	opts TaskViewOptions
}

// NewTaskView creates a new task view UI component
func NewTaskView(task *models.Task, opts TaskViewOptions) *TaskView {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	// Set default dimensions if not provided
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.Height == 0 {
		opts.Height = 24
	}
	return &TaskView{task: task, opts: opts}
}

// Task view specific key bindings
type taskViewKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Quit     key.Binding
	Back     key.Binding
	Help     key.Binding
}

func (k taskViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Back, k.Help, k.Quit}
}

func (k taskViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Top, k.Bottom},
		{k.Help, k.Back, k.Quit},
	}
}

var taskViewKeys = taskViewKeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
	PageUp:   key.NewBinding(key.WithKeys("pgup", "b"), key.WithHelp("pgup/b", "page up")),
	PageDown: key.NewBinding(key.WithKeys("pgdown", "f"), key.WithHelp("pgdown/f", "page down")),
	Top:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("home/g", "go to top")),
	Bottom:   key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("end/G", "go to bottom")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Back:     key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

type taskViewModel struct {
	task        *models.Task
	viewport    viewport.Model
	keys        taskViewKeyMap
	help        help.Model
	showingHelp bool
	opts        TaskViewOptions
}

func (m taskViewModel) Init() tea.Cmd {
	return nil
}

func (m taskViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showingHelp {
			switch {
			case key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Help):
				m.showingHelp = false
				return m, nil
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.Back):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showingHelp = true
			return m, nil
		case key.Matches(msg, m.keys.Up):
			m.viewport.ScrollUp(1)
		case key.Matches(msg, m.keys.Down):
			m.viewport.ScrollDown(1)
		case key.Matches(msg, m.keys.PageUp):
			m.viewport.HalfPageUp()
		case key.Matches(msg, m.keys.PageDown):
			m.viewport.HalfPageDown()
		case key.Matches(msg, m.keys.Top):
			m.viewport.GotoTop()
		case key.Matches(msg, m.keys.Bottom):
			m.viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		headerHeight := 3 // Title + spacing
		footerHeight := 3 // Help + spacing
		verticalMarginHeight := headerHeight + footerHeight

		if !m.opts.Static {
			m.viewport.Width = msg.Width - 2 // Account for padding
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m taskViewModel) View() string {
	if m.showingHelp {
		return m.help.View(m.keys)
	}

	title := TableTitleStyle.Render(fmt.Sprintf("Task %d", m.task.ID))
	content := m.viewport.View()
	help := MutedStyle.Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", help)
}

func formatTaskContent(task *models.Task) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("UUID: %s\n", task.UUID))
	content.WriteString(fmt.Sprintf("Description: %s\n", task.Description))
	content.WriteString(fmt.Sprintf("Status: %s\n", utils.Titlecase(task.Status)))

	if task.Priority != "" {
		content.WriteString(fmt.Sprintf("Priority: %s\n", utils.Titlecase(task.Priority)))
	}

	if task.Project != "" {
		content.WriteString(fmt.Sprintf("Project: %s\n", task.Project))
	}

	if len(task.Tags) > 0 {
		content.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(task.Tags, ", ")))
	}

	content.WriteString("\nDates:\n")
	content.WriteString(fmt.Sprintf("- Created: %s\n", task.Entry.Format("2006-01-02 15:04")))
	content.WriteString(fmt.Sprintf("- Modified: %s\n", task.Modified.Format("2006-01-02 15:04")))

	if task.Due != nil {
		content.WriteString(fmt.Sprintf("- Due: %s\n", task.Due.Format("2006-01-02 15:04")))
	}

	if task.Start != nil {
		content.WriteString(fmt.Sprintf("- Started: %s\n", task.Start.Format("2006-01-02 15:04")))
	}

	if task.End != nil {
		content.WriteString(fmt.Sprintf("- Completed: %s\n", task.End.Format("2006-01-02 15:04")))
	}

	if len(task.Annotations) > 0 {
		content.WriteString("\nAnnotations:\n")
		for i, annotation := range task.Annotations {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, annotation))
		}
	}

	return content.String()
}

// Show displays the task in interactive mode
func (tv *TaskView) Show(ctx context.Context) error {
	if tv.opts.Static {
		return tv.staticShow(ctx)
	}

	vp := viewport.New(tv.opts.Width-2, tv.opts.Height-6)
	vp.SetContent(formatTaskContent(tv.task))

	model := taskViewModel{
		task:     tv.task,
		viewport: vp,
		keys:     taskViewKeys,
		help:     help.New(),
		opts:     tv.opts,
	}

	program := tea.NewProgram(model, tea.WithInput(tv.opts.Input), tea.WithOutput(tv.opts.Output))

	_, err := program.Run()
	return err
}

func (tv *TaskView) staticShow(context.Context) error {
	content := formatTaskContent(tv.task)

	title := fmt.Sprintf("Task %d\n\n", tv.task.ID)

	fmt.Fprint(tv.opts.Output, title+content)
	return nil
}
