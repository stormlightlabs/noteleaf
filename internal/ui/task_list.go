package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

var (
	TitleColorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(Guac.Hex())).Bold(true)
	SelectedColorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(Salt.Hex())).Bold(true)
	HeaderColorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(Malibu.Hex())).Bold(true)
	StatusColorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(Julep.Hex()))
)

// Key bindings for task list navigation
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	View      key.Binding
	Refresh   key.Binding
	ToggleAll key.Binding
	MarkDone  key.Binding
	Quit      key.Binding
	Back      key.Binding
	Help      key.Binding
	Numbers   []key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.MarkDone, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.View},
		{k.MarkDone, k.Refresh, k.ToggleAll},
		{k.Help, k.Quit, k.Back},
	}
}

var keys = keyMap{
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view task")),
	View:      key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view task")),
	Refresh:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	ToggleAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle all/pending")),
	MarkDone:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "mark done")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Back:      key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Numbers: []key.Binding{
		key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "jump to task 1")),
		key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "jump to task 2")),
		key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "jump to task 3")),
		key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "jump to task 4")),
		key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "jump to task 5")),
		key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "jump to task 6")),
		key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "jump to task 7")),
		key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "jump to task 8")),
		key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "jump to task 9")),
	},
}

// TaskRepository interface for dependency injection in tests
type TaskRepository interface {
	List(ctx context.Context, opts repo.TaskListOptions) ([]*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
}

// TaskListOptions configures the task list UI behavior
type TaskListOptions struct {
	// Output destination (stdout for interactive, buffer for testing)
	Output io.Writer
	// Input source (stdin for interactive, strings reader for testing)
	Input io.Reader
	// Enable static mode (no interactive components)
	Static   bool
	Status   string
	Priority string
	Project  string
	ShowAll  bool
}

// TaskList handles task browsing and viewing UI
type TaskList struct {
	repo TaskRepository
	opts TaskListOptions
}

// NewTaskList creates a new task list UI component
func NewTaskList(repo TaskRepository, opts TaskListOptions) *TaskList {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	return &TaskList{repo: repo, opts: opts}
}

type (
	tasksLoadedMsg []*models.Task
	taskViewMsg    string
	errorTaskMsg   error
)

type taskListModel struct {
	tasks       []*models.Task
	selected    int
	viewing     bool
	viewContent string
	err         error
	repo        TaskRepository
	opts        TaskListOptions
	showAll     bool
	keys        keyMap
	help        help.Model
	showingHelp bool
}

func (m taskListModel) Init() tea.Cmd {
	return m.loadTasks()
}

func (m taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		if m.viewing {
			switch {
			case key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit):
				m.viewing = false
				m.viewContent = ""
				return m, nil
			case key.Matches(msg, m.keys.Help):
				m.showingHelp = true
				return m, nil
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, m.keys.Down):
			if m.selected < len(m.tasks)-1 {
				m.selected++
			}
		case key.Matches(msg, m.keys.Enter) || key.Matches(msg, m.keys.View):
			if len(m.tasks) > 0 && m.selected < len(m.tasks) {
				return m, m.viewTask(m.tasks[m.selected])
			}
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadTasks()
		case key.Matches(msg, m.keys.ToggleAll):
			m.showAll = !m.showAll
			return m, m.loadTasks()
		case key.Matches(msg, m.keys.MarkDone):
			if len(m.tasks) > 0 && m.selected < len(m.tasks) {
				return m, m.markDone(m.tasks[m.selected])
			}
		case key.Matches(msg, m.keys.Help):
			m.showingHelp = true
			return m, nil
		default:
			for i, numKey := range m.keys.Numbers {
				if key.Matches(msg, numKey) && i < len(m.tasks) {
					m.selected = i
					break
				}
			}
		}
	case tasksLoadedMsg:
		m.tasks = []*models.Task(msg)
		if m.selected >= len(m.tasks) && len(m.tasks) > 0 {
			m.selected = len(m.tasks) - 1
		}
	case taskViewMsg:
		m.viewContent = string(msg)
		m.viewing = true
	case errorTaskMsg:
		m.err = error(msg)
	}
	return m, nil
}

func (m taskListModel) View() string {
	var s strings.Builder

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(Squid.Hex()))

	if m.showingHelp {
		return m.help.View(m.keys)
	}

	if m.viewing {
		s.WriteString(m.viewContent)
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press q/esc/backspace to return to list, ? for help"))
		return s.String()
	}

	s.WriteString(TitleColorStyle.Render("Tasks"))
	if m.showAll {
		s.WriteString(" (showing all)")
	} else {
		s.WriteString(" (pending only)")
	}
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.tasks) == 0 {
		s.WriteString("No tasks found")
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press r to refresh, q to quit"))
		return s.String()
	}

	headerLine := fmt.Sprintf("   %-4s %-40s %-10s %-10s %-15s", "ID", "Description", "Status", "Priority", "Project")
	s.WriteString(HeaderColorStyle.Render(headerLine))
	s.WriteString("\n")
	s.WriteString(HeaderColorStyle.Render(strings.Repeat("─", 80)))
	s.WriteString("\n")

	for i, task := range m.tasks {
		prefix := "   "
		if i == m.selected {
			prefix = " > "
		}

		description := task.Description
		if len(description) > 38 {
			description = description[:35] + "..."
		}

		status := task.Status
		if len(status) > 8 {
			status = status[:8]
		}

		priority := task.Priority
		if priority == "" {
			priority = "-"
		}
		if len(priority) > 8 {
			priority = priority[:8]
		}

		project := task.Project
		if project == "" {
			project = "-"
		}
		if len(project) > 13 {
			project = project[:10] + "..."
		}

		priority = utils.Titlecase(priority)
		padded := fmt.Sprintf("%-10s", priority)
		var colored string
		switch strings.ToLower(task.Priority) {
		case "high", "urgent":
			colored = PriorityHigh.Render(padded)
		case "medium":
			colored = PriorityMedium.Render(padded)
		case "low":
			colored = PriorityLow.Render(padded)
		default:
			colored = padded
		}

		line := fmt.Sprintf("%s%-4d %-40s %-10s %s %-15s", prefix, task.ID, description, status, colored, project)

		if i == m.selected {
			s.WriteString(SelectedColorStyle.Render(line))
		} else {
			if task.Status == "completed" {
				s.WriteString(StatusColorStyle.Render(line))
			} else {
				s.WriteString(style.Render(line))
			}
		}

		if len(task.Tags) > 0 && i == m.selected {
			s.WriteString(" @" + strings.Join(task.Tags, " @"))
		}

		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))

	return s.String()
}

func (m taskListModel) loadTasks() tea.Cmd {
	return func() tea.Msg {
		opts := repo.TaskListOptions{}
		showAll := m.showAll || m.opts.ShowAll

		if !showAll {
			opts.Status = "pending"
		}

		if m.opts.Status != "" {
			opts.Status = m.opts.Status
		}
		if m.opts.Priority != "" {
			opts.Priority = m.opts.Priority
		}
		if m.opts.Project != "" {
			opts.Project = m.opts.Project
		}

		opts.SortBy = "modified"
		opts.SortOrder = "DESC"
		opts.Limit = 50

		tasks, err := m.repo.List(context.Background(), opts)
		if err != nil {
			return errorTaskMsg(err)
		}

		return tasksLoadedMsg(tasks)
	}
}

func (m taskListModel) viewTask(task *models.Task) tea.Cmd {
	return func() tea.Msg {
		var content strings.Builder
		content.WriteString(fmt.Sprintf("# Task %d\n\n", task.ID))
		content.WriteString(fmt.Sprintf("**UUID:** %s\n", task.UUID))
		content.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))
		content.WriteString(fmt.Sprintf("**Status:** %s\n", task.Status))

		if task.Priority != "" {
			content.WriteString(fmt.Sprintf("**Priority:** %s\n", task.Priority))
		}

		if task.Project != "" {
			content.WriteString(fmt.Sprintf("**Project:** %s\n", task.Project))
		}

		if len(task.Tags) > 0 {
			content.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(task.Tags, ", ")))
		}

		if task.Due != nil {
			content.WriteString(fmt.Sprintf("**Due:** %s\n", task.Due.Format("2006-01-02 15:04")))
		}

		content.WriteString(fmt.Sprintf("**Created:** %s\n", task.Entry.Format("2006-01-02 15:04")))
		content.WriteString(fmt.Sprintf("**Modified:** %s\n", task.Modified.Format("2006-01-02 15:04")))

		if task.Start != nil {
			content.WriteString(fmt.Sprintf("**Started:** %s\n", task.Start.Format("2006-01-02 15:04")))
		}

		if task.End != nil {
			content.WriteString(fmt.Sprintf("**Completed:** %s\n", task.End.Format("2006-01-02 15:04")))
		}

		if len(task.Annotations) > 0 {
			content.WriteString("\n**Annotations:**\n")
			for _, annotation := range task.Annotations {
				content.WriteString(fmt.Sprintf("- %s\n", annotation))
			}
		}

		return taskViewMsg(content.String())
	}
}

func (m taskListModel) markDone(task *models.Task) tea.Cmd {
	return func() tea.Msg {
		if task.Status == "completed" {
			return errorTaskMsg(fmt.Errorf("task already completed"))
		}

		task.Status = "completed"
		err := m.repo.Update(context.Background(), task)
		if err != nil {
			return errorTaskMsg(fmt.Errorf("failed to mark task done: %w", err))
		}

		return m.loadTasks()()
	}
}

// Browse opens an interactive TUI for navigating and viewing tasks
func (tl *TaskList) Browse(ctx context.Context) error {
	if tl.opts.Static {
		return tl.staticList(ctx)
	}

	model := taskListModel{
		repo:    tl.repo,
		opts:    tl.opts,
		showAll: tl.opts.ShowAll,
		keys:    keys,
		help:    help.New(),
	}

	program := tea.NewProgram(model, tea.WithInput(tl.opts.Input), tea.WithOutput(tl.opts.Output))

	_, err := program.Run()
	return err
}

func (tl *TaskList) staticList(ctx context.Context) error {
	opts := repo.TaskListOptions{}

	if !tl.opts.ShowAll {
		opts.Status = "pending"
	}

	if tl.opts.Status != "" {
		opts.Status = tl.opts.Status
	}
	if tl.opts.Priority != "" {
		opts.Priority = tl.opts.Priority
	}
	if tl.opts.Project != "" {
		opts.Project = tl.opts.Project
	}

	opts.SortBy = "modified"
	opts.SortOrder = "DESC"

	tasks, err := tl.repo.List(ctx, opts)
	if err != nil {
		fmt.Fprintf(tl.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(tl.opts.Output, "Tasks")
	if tl.opts.ShowAll {
		fmt.Fprintf(tl.opts.Output, " (showing all)")
	} else {
		fmt.Fprintf(tl.opts.Output, " (pending only)")
	}
	fmt.Fprintf(tl.opts.Output, "\n\n")

	if len(tasks) == 0 {
		fmt.Fprintf(tl.opts.Output, "No tasks found\n")
		return nil
	}

	fmt.Fprintf(tl.opts.Output, "%-4s %-40s %-10s %-10s %-15s\n", "ID", "Description", "Status", "Priority", "Project")
	fmt.Fprintf(tl.opts.Output, "%s\n", strings.Repeat("─", 80))

	for _, task := range tasks {
		description := task.Description
		if len(description) > 38 {
			description = description[:35] + "..."
		}

		status := task.Status
		if len(status) > 8 {
			status = status[:8]
		}

		priority := task.Priority
		if priority == "" {
			priority = "-"
		}
		if len(priority) > 8 {
			priority = priority[:8]
		}

		project := task.Project
		if project == "" {
			project = "-"
		}
		if len(project) > 13 {
			project = project[:10] + "..."
		}

		fmt.Fprintf(tl.opts.Output, "%-4d %-40s %-10s %-10s %-15s", task.ID, description, status, priority, project)

		if len(task.Tags) > 0 {
			fmt.Fprintf(tl.opts.Output, " @%s", strings.Join(task.Tags, " @"))
		}

		fmt.Fprintf(tl.opts.Output, "\n")
	}

	return nil
}
