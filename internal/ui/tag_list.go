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

// Tag list key bindings
type tagKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Refresh key.Binding
	Quit    key.Binding
	Back    key.Binding
	Help    key.Binding
	Numbers []key.Binding
}

func (k tagKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Help, k.Quit}
}

func (k tagKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Refresh},
		{k.Help, k.Quit, k.Back},
	}
}

var tagKeys = tagKeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select tag")),
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Back:    key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Numbers: []key.Binding{
		key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "jump to tag 1")),
		key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "jump to tag 2")),
		key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "jump to tag 3")),
		key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "jump to tag 4")),
		key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "jump to tag 5")),
		key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "jump to tag 6")),
		key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "jump to tag 7")),
		key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "jump to tag 8")),
		key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "jump to tag 9")),
	},
}

type (
	// TagRepository interface for dependency injection in tests
	TagRepository interface {
		GetTags(ctx context.Context) ([]repo.TagSummary, error)
	}

	// TagListOptions configures the tag list UI behavior
	TagListOptions struct {
		// Output destination (stdout for interactive, buffer for testing)
		Output io.Writer
		// Input source (stdin for interactive, strings reader for testing)
		Input io.Reader
		// Enable static mode (no interactive components)
		Static bool
	}

	// TagList handles tag browsing UI
	TagList struct {
		repo TagRepository
		opts TagListOptions
	}
)

// NewTagList creates a new tag list UI component
func NewTagList(repo TagRepository, opts TagListOptions) *TagList {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	return &TagList{repo: repo, opts: opts}
}

type (
	tagsLoadedMsg []repo.TagSummary
	errorTagMsg   error
	tagListModel  struct {
		tags        []repo.TagSummary
		selected    int
		err         error
		repo        TagRepository
		opts        TagListOptions
		keys        tagKeyMap
		help        help.Model
		showingHelp bool
	}
)

func (m tagListModel) Init() tea.Cmd {
	return m.loadTags()
}

func (m tagListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.selected < len(m.tags)-1 {
				m.selected++
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.tags) > 0 && m.selected < len(m.tags) {
				// For now, just show the selected tag name
				// In a real implementation, this might navigate to tasks with that tag
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadTags()
		case key.Matches(msg, m.keys.Help):
			m.showingHelp = true
			return m, nil
		default:
			// Handle number keys for quick selection
			for i, numKey := range m.keys.Numbers {
				if key.Matches(msg, numKey) && i < len(m.tags) {
					m.selected = i
					break
				}
			}
		}
	case tagsLoadedMsg:
		m.tags = []repo.TagSummary(msg)
		if m.selected >= len(m.tags) && len(m.tags) > 0 {
			m.selected = len(m.tags) - 1
		}
	case errorTagMsg:
		m.err = error(msg)
	}
	return m, nil
}

func (m tagListModel) View() string {
	var s strings.Builder

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(Squid.Hex()))

	if m.showingHelp {
		return m.help.View(m.keys)
	}

	s.WriteString(TitleColorStyle.Render("Tags"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.tags) == 0 {
		s.WriteString("No tags found")
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press r to refresh, q to quit"))
		return s.String()
	}

	headerLine := fmt.Sprintf("   %-25s %-15s", "Tag Name", "Task Count")
	s.WriteString(HeaderColorStyle.Render(headerLine))
	s.WriteString("\n")
	s.WriteString(HeaderColorStyle.Render(strings.Repeat("─", 45)))
	s.WriteString("\n")

	for i, tag := range m.tags {
		prefix := "   "
		if i == m.selected {
			prefix = " > "
		}

		tagName := tag.Name
		if len(tagName) > 23 {
			tagName = tagName[:20] + "..."
		}

		taskCountStr := fmt.Sprintf("%d task%s", tag.TaskCount, pluralizeCount(tag.TaskCount))

		line := fmt.Sprintf("%s%-25s %-15s", prefix, tagName, taskCountStr)

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

func (m tagListModel) loadTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.repo.GetTags(context.Background())
		if err != nil {
			return errorTagMsg(err)
		}

		return tagsLoadedMsg(tags)
	}
}

// Browse opens an interactive TUI for navigating tags
func (tl *TagList) Browse(ctx context.Context) error {
	if tl.opts.Static {
		return tl.staticList(ctx)
	}

	model := tagListModel{
		repo: tl.repo,
		opts: tl.opts,
		keys: tagKeys,
		help: help.New(),
	}

	program := tea.NewProgram(model, tea.WithInput(tl.opts.Input), tea.WithOutput(tl.opts.Output))

	_, err := program.Run()
	return err
}

func (tl *TagList) staticList(ctx context.Context) error {
	tags, err := tl.repo.GetTags(ctx)
	if err != nil {
		fmt.Fprintf(tl.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(tl.opts.Output, "Tags\n\n")

	if len(tags) == 0 {
		fmt.Fprintf(tl.opts.Output, "No tags found\n")
		return nil
	}

	fmt.Fprintf(tl.opts.Output, "%-25s %-15s\n", "Tag Name", "Task Count")
	fmt.Fprintf(tl.opts.Output, "%s\n", strings.Repeat("─", 45))

	for _, tag := range tags {
		tagName := tag.Name
		if len(tagName) > 23 {
			tagName = tagName[:20] + "..."
		}

		taskCountStr := fmt.Sprintf("%d task%s", tag.TaskCount, pluralizeCount(tag.TaskCount))

		fmt.Fprintf(tl.opts.Output, "%-25s %-15s\n", tagName, taskCountStr)
	}

	return nil
}