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
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// Project list key bindings
type projectKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Refresh key.Binding
	Quit    key.Binding
	Back    key.Binding
	Help    key.Binding
	Numbers []key.Binding
}

func (k projectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Help, k.Quit}
}

func (k projectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Refresh},
		{k.Help, k.Quit, k.Back},
	}
}

var projectKeys = projectKeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select project")),
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Back:    key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Numbers: []key.Binding{
		key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "jump to project 1")),
		key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "jump to project 2")),
		key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "jump to project 3")),
		key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "jump to project 4")),
		key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "jump to project 5")),
		key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "jump to project 6")),
		key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "jump to project 7")),
		key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "jump to project 8")),
		key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "jump to project 9")),
	},
}

// ProjectRepository interface for dependency injection in tests
type ProjectRepository interface {
	GetProjects(ctx context.Context) ([]repo.ProjectSummary, error)
}

// ProjectListOptions configures the project list UI behavior
type ProjectListOptions struct {
	// Output destination (stdout for interactive, buffer for testing)
	Output io.Writer
	// Input source (stdin for interactive, strings reader for testing)
	Input io.Reader
	// Enable static mode (no interactive components)
	Static bool
}

// ProjectList handles project browsing UI
type ProjectList struct {
	repo ProjectRepository
	opts ProjectListOptions
}

// NewProjectList creates a new project list UI component
func NewProjectList(repo ProjectRepository, opts ProjectListOptions) *ProjectList {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	return &ProjectList{repo: repo, opts: opts}
}

type (
	projectsLoadedMsg []repo.ProjectSummary
	errorProjectMsg   error
	projectListModel  struct {
		projects    []repo.ProjectSummary
		selected    int
		err         error
		repo        ProjectRepository
		opts        ProjectListOptions
		keys        projectKeyMap
		help        help.Model
		showingHelp bool
	}
)

func (m projectListModel) Init() tea.Cmd {
	return m.loadProjects()
}

func (m projectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, m.keys.Down):
			if m.selected < len(m.projects)-1 {
				m.selected++
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.projects) > 0 && m.selected < len(m.projects) {
				// TODO: navigate to tasks for that project
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadProjects()
		case key.Matches(msg, m.keys.Help):
			m.showingHelp = true
			return m, nil
		default:
			for i, numKey := range m.keys.Numbers {
				if key.Matches(msg, numKey) && i < len(m.projects) {
					m.selected = i
					break
				}
			}
		}
	case projectsLoadedMsg:
		m.projects = []repo.ProjectSummary(msg)
		if m.selected >= len(m.projects) && len(m.projects) > 0 {
			m.selected = len(m.projects) - 1
		}
	case errorProjectMsg:
		m.err = error(msg)
	}
	return m, nil
}

func (m projectListModel) View() string {
	var s strings.Builder

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(Squid.Hex()))

	if m.showingHelp {
		return m.help.View(m.keys)
	}

	s.WriteString(TitleColorStyle.Render("Projects"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.projects) == 0 {
		s.WriteString("No projects found")
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press r to refresh, q to quit"))
		return s.String()
	}

	headerLine := fmt.Sprintf("   %-30s %-15s", "Project Name", "Task Count")
	s.WriteString(HeaderColorStyle.Render(headerLine))
	s.WriteString("\n")
	s.WriteString(HeaderColorStyle.Render(strings.Repeat("─", 50)))
	s.WriteString("\n")

	for i, project := range m.projects {
		prefix := "   "
		if i == m.selected {
			prefix = " > "
		}

		projectName := project.Name
		if len(projectName) > 28 {
			projectName = projectName[:25] + "..."
		}

		taskCountStr := fmt.Sprintf("%d task%s", project.TaskCount, pluralizeCount(project.TaskCount))

		line := fmt.Sprintf("%s%-30s %-15s", prefix, projectName, taskCountStr)

		if i == m.selected {
			s.WriteString(SelectedColorStyle.Render(line))
		} else {
			s.WriteString(style.Render(line))
		}

		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))

	return s.String()
}

func (m projectListModel) loadProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.repo.GetProjects(context.Background())
		if err != nil {
			return errorProjectMsg(err)
		}

		return projectsLoadedMsg(projects)
	}
}

// Browse opens an interactive TUI for navigating projects
func (pl *ProjectList) Browse(ctx context.Context) error {
	if pl.opts.Static {
		return pl.staticList(ctx)
	}

	model := projectListModel{
		repo: pl.repo,
		opts: pl.opts,
		keys: projectKeys,
		help: help.New(),
	}

	program := tea.NewProgram(model, tea.WithInput(pl.opts.Input), tea.WithOutput(pl.opts.Output))

	_, err := program.Run()
	return err
}

func (pl *ProjectList) staticList(ctx context.Context) error {
	projects, err := pl.repo.GetProjects(ctx)
	if err != nil {
		fmt.Fprintf(pl.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(pl.opts.Output, "Projects\n\n")

	if len(projects) == 0 {
		fmt.Fprintf(pl.opts.Output, "No projects found\n")
		return nil
	}

	fmt.Fprintf(pl.opts.Output, "%-30s %-15s\n", "Project Name", "Task Count")
	fmt.Fprintf(pl.opts.Output, "%s\n", strings.Repeat("─", 50))

	for _, project := range projects {
		projectName := project.Name
		if len(projectName) > 28 {
			projectName = projectName[:25] + "..."
		}

		taskCountStr := fmt.Sprintf("%d task%s", project.TaskCount, pluralizeCount(project.TaskCount))

		fmt.Fprintf(pl.opts.Output, "%-30s %-15s\n", projectName, taskCountStr)
	}

	return nil
}

func pluralizeCount(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
