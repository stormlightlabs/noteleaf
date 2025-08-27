// FIXME: this module is missing test coverage
package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// NoteListOptions configures the note list UI behavior
type NoteListOptions struct {
	// Output destination (stdout for interactive, buffer for testing)
	Output io.Writer
	// Input source (stdin for interactive, strings reader for testing)
	Input io.Reader
	// Enable static mode (no interactive components)
	Static bool
	// Show archived notes
	ShowArchived bool
	// Filter by tags
	Tags []string
}

// NoteList handles note browsing and viewing UI
type NoteList struct {
	repo *repo.NoteRepository
	opts NoteListOptions
}

// NewNoteList creates a new note list UI component
func NewNoteList(repo *repo.NoteRepository, opts NoteListOptions) *NoteList {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	return &NoteList{repo: repo, opts: opts}
}

type noteListModel struct {
	notes       []*models.Note
	selected    int
	viewing     bool
	viewContent string
	err         error
	repo        *repo.NoteRepository
	opts        NoteListOptions
	currentPage int
}

type notesLoadedMsg []*models.Note
type noteViewMsg string
type errorNoteMsg error

func (m noteListModel) Init() tea.Cmd {
	return m.loadNotes()
}

func (m noteListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.selected < len(m.notes)-1 {
				m.selected++
			}
		case "enter", "v":
			if len(m.notes) > 0 && m.selected < len(m.notes) {
				return m, m.viewNote(m.notes[m.selected])
			}
		case "r":
			return m, m.loadNotes()
		}
	case notesLoadedMsg:
		m.notes = []*models.Note(msg)
	case noteViewMsg:
		m.viewContent = string(msg)
		m.viewing = true
	case errorNoteMsg:
		m.err = error(msg)
	}
	return m, nil
}

func (m noteListModel) View() string {
	var s strings.Builder

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)

	if m.viewing {
		s.WriteString(m.viewContent)
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press q/esc/backspace to return to list"))
		return s.String()
	}

	s.WriteString(titleStyle.Render("Notes"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.notes) == 0 {
		s.WriteString("No notes found")
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press r to refresh, q to quit"))
		return s.String()
	}

	headerLine := fmt.Sprintf("%-4s %-30s %-20s %-15s", "ID", "Title", "Tags", "Modified")
	s.WriteString(headerStyle.Render(headerLine))
	s.WriteString("\n")
	s.WriteString(headerStyle.Render(strings.Repeat("─", 70)))
	s.WriteString("\n")

	for i, note := range m.notes {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}

		title := note.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}

		tags := ""
		if len(note.Tags) > 0 {
			tags = strings.Join(note.Tags, ", ")
			if len(tags) > 18 {
				tags = tags[:15] + "..."
			}
		}

		modified := note.Modified.Format("2006-01-02 15:04")

		line := fmt.Sprintf("%s%-4d %-30s %-20s %-15s",
			prefix, note.ID, title, tags, modified)

		if i == m.selected {
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(style.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(style.Render("Use ↑/↓ to navigate, Enter/v to view, r to refresh, q to quit"))

	return s.String()
}

func (m noteListModel) loadNotes() tea.Cmd {
	return func() tea.Msg {
		opts := repo.NoteListOptions{
			Tags: m.opts.Tags,
		}
		if !m.opts.ShowArchived {
			archived := false
			opts.Archived = &archived
		}

		notes, err := m.repo.List(context.Background(), opts)
		if err != nil {
			return errorNoteMsg(err)
		}

		return notesLoadedMsg(notes)
	}
}

func (m noteListModel) viewNote(note *models.Note) tea.Cmd {
	return func() tea.Msg {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err != nil {
			return errorNoteMsg(fmt.Errorf("failed to create markdown renderer: %w", err))
		}

		content := m.formatNoteForView(note)
		rendered, err := renderer.Render(content)
		if err != nil {
			return errorNoteMsg(fmt.Errorf("failed to render markdown: %w", err))
		}

		return noteViewMsg(rendered)
	}
}

func (m noteListModel) formatNoteForView(note *models.Note) string {
	var content strings.Builder

	content.WriteString("# " + note.Title + "\n\n")

	if len(note.Tags) > 0 {
		content.WriteString("**Tags:** ")
		for i, tag := range note.Tags {
			if i > 0 {
				content.WriteString(", ")
			}
			content.WriteString("`" + tag + "`")
		}
		content.WriteString("\n\n")
	}

	content.WriteString("**Created:** " + note.Created.Format("2006-01-02 15:04") + "\n")
	content.WriteString("**Modified:** " + note.Modified.Format("2006-01-02 15:04") + "\n\n")
	content.WriteString("---\n\n")

	noteContent := strings.TrimSpace(note.Content)
	if !strings.HasPrefix(noteContent, "# ") {
		content.WriteString(noteContent)
	} else {
		lines := strings.Split(noteContent, "\n")
		if len(lines) > 1 {
			content.WriteString(strings.Join(lines[1:], "\n"))
		}
	}

	return content.String()
}

// Browse opens an interactive TUI for navigating and viewing notes
func (nl *NoteList) Browse(ctx context.Context) error {
	if nl.opts.Static {
		return nl.staticList(ctx)
	}

	model := noteListModel{
		repo:        nl.repo,
		opts:        nl.opts,
		currentPage: 1,
	}

	program := tea.NewProgram(model, tea.WithInput(nl.opts.Input), tea.WithOutput(nl.opts.Output))

	_, err := program.Run()
	return err
}

func (nl *NoteList) staticList(ctx context.Context) error {
	opts := repo.NoteListOptions{
		Tags: nl.opts.Tags,
	}
	if !nl.opts.ShowArchived {
		archived := false
		opts.Archived = &archived
	}

	notes, err := nl.repo.List(ctx, opts)
	if err != nil {
		fmt.Fprintf(nl.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(nl.opts.Output, "Notes\n\n")

	if len(notes) == 0 {
		fmt.Fprintf(nl.opts.Output, "No notes found\n")
		return nil
	}

	fmt.Fprintf(nl.opts.Output, "%-4s %-30s %-20s %-15s\n", "ID", "Title", "Tags", "Modified")
	fmt.Fprintf(nl.opts.Output, "%s\n", strings.Repeat("─", 70))

	for _, note := range notes {
		title := note.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}

		tags := ""
		if len(note.Tags) > 0 {
			tags = strings.Join(note.Tags, ", ")
			if len(tags) > 18 {
				tags = tags[:15] + "..."
			}
		}

		modified := note.Modified.Format("2006-01-02 15:04")

		fmt.Fprintf(nl.opts.Output, "%-4d %-30s %-20s %-15s\n", note.ID, title, tags, modified)
	}

	return nil
}
