package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

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

type taskListModel struct {
	tasks       []*models.Task
	selected    int
	viewing     bool
	viewContent string
	err         error
	repo        TaskRepository
	opts        TaskListOptions
	showAll     bool
	// filter      string
}

type tasksLoadedMsg []*models.Task
type taskViewMsg string
type errorTaskMsg error

func (m taskListModel) Init() tea.Cmd {
	return m.loadTasks()
}

func (m taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.viewing {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.viewing = false
				m.viewContent = ""
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.tasks)-1 {
				m.selected++
			}
		case "enter", "v":
			if len(m.tasks) > 0 && m.selected < len(m.tasks) {
				return m, m.viewTask(m.tasks[m.selected])
			}
		case "r":
			return m, m.loadTasks()
		case "a":
			m.showAll = !m.showAll
			return m, m.loadTasks()
		case "d":
			if len(m.tasks) > 0 && m.selected < len(m.tasks) {
				return m, m.markDone(m.tasks[m.selected])
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if idx := int(msg.String()[0] - '1'); idx < len(m.tasks) {
				m.selected = idx
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

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	priorityHighStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	priorityMediumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	priorityLowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("28"))

	if m.viewing {
		s.WriteString(m.viewContent)
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press q/esc/backspace to return to list"))
		return s.String()
	}

	s.WriteString(titleStyle.Render("Tasks"))
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

	headerLine := fmt.Sprintf("%-3s %-4s %-40s %-10s %-10s %-15s", "", "ID", "Description", "Status", "Priority", "Project")
	s.WriteString(headerStyle.Render(headerLine))
	s.WriteString("\n")
	s.WriteString(headerStyle.Render(strings.Repeat("─", 80)))
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

		line := fmt.Sprintf("%s%-4d %-40s %-10s %-10s %-15s",
			prefix, task.ID, description, status, priority, project)

		if i == m.selected {
			s.WriteString(selectedStyle.Render(line))
		} else {
			// Color based on priority
			switch strings.ToLower(task.Priority) {
			case "high", "urgent":
				s.WriteString(priorityHighStyle.Render(line))
			case "medium":
				s.WriteString(priorityMediumStyle.Render(line))
			case "low":
				s.WriteString(priorityLowStyle.Render(line))
			default:
				if task.Status == "completed" {
					s.WriteString(statusStyle.Render(line))
				} else {
					s.WriteString(style.Render(line))
				}
			}
		}

		// Add tags if any
		if len(task.Tags) > 0 && i == m.selected {
			s.WriteString(" @" + strings.Join(task.Tags, " @"))
		}

		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(style.Render("Controls: ↑/↓/k/j to navigate, Enter/v to view, d to mark done, a to toggle all/pending"))
	s.WriteString("\n")
	s.WriteString(style.Render("r to refresh, q to quit, 1-9 to jump to task"))

	return s.String()
}

func (m taskListModel) loadTasks() tea.Cmd {
	return func() tea.Msg {
		opts := repo.TaskListOptions{}

		// Set status filter
		if m.showAll || m.opts.ShowAll {
			// Show all tasks - no status filter
		} else {
			opts.Status = "pending"
		}

		// Apply other filters from options
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

		// Reload tasks after marking done
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
	}

	program := tea.NewProgram(model, tea.WithInput(tl.opts.Input), tea.WithOutput(tl.opts.Output))

	_, err := program.Run()
	return err
}

func (tl *TaskList) staticList(ctx context.Context) error {
	opts := repo.TaskListOptions{}

	if tl.opts.ShowAll {
		// Show all tasks - no status filter
	} else {
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
