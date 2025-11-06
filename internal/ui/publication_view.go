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
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

// PublicationViewOptions configures the publication view UI behavior
type PublicationViewOptions struct {
	Output io.Writer
	Input  io.Reader
	Static bool
	Width  int
	Height int
}

// PublicationView handles publication viewing UI with pager
type PublicationView struct {
	note *models.Note
	opts PublicationViewOptions
}

// NewPublicationView creates a new publication view UI component
func NewPublicationView(note *models.Note, opts PublicationViewOptions) *PublicationView {
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
	return &PublicationView{note: note, opts: opts}
}

// Publication view specific key bindings
type publicationViewKeyMap struct {
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

func (k publicationViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Back, k.Help, k.Quit}
}

func (k publicationViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Top, k.Bottom},
		{k.Help, k.Back, k.Quit},
	}
}

var publicationViewKeys = publicationViewKeyMap{
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

type publicationViewModel struct {
	note        *models.Note
	viewport    viewport.Model
	keys        publicationViewKeyMap
	help        help.Model
	showingHelp bool
	opts        PublicationViewOptions
	ready       bool
}

func (m publicationViewModel) Init() tea.Cmd {
	return nil
}

func (m publicationViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		headerHeight := 3
		footerHeight := 3
		verticalMarginHeight := headerHeight + footerHeight

		if !m.opts.Static {
			m.viewport.Width = msg.Width - 2
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if !m.ready {
			m.ready = true
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m publicationViewModel) View() string {
	if m.showingHelp {
		return m.help.View(m.keys)
	}

	status := "published"
	if m.note.IsDraft {
		status = "draft"
	}

	title := TitleColorStyle.Render(fmt.Sprintf("%s (%s)", m.note.Title, status))
	content := m.viewport.View()
	help := lipgloss.NewStyle().Foreground(lipgloss.Color(Squid.Hex())).Render(m.help.View(m.keys))

	if !m.ready {
		return "\n  Initializing..."
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", help)
}

// buildPublicationMarkdown creates the markdown content for rendering
func buildPublicationMarkdown(note *models.Note) string {
	var content strings.Builder

	content.WriteString("# " + note.Title + "\n\n")

	status := "published"
	if note.IsDraft {
		status = "draft"
	}
	content.WriteString("**Status:** " + status + "\n")

	if note.PublishedAt != nil {
		content.WriteString("**Published:** " + note.PublishedAt.Format("2006-01-02 15:04") + "\n")
	}

	content.WriteString("**Modified:** " + note.Modified.Format("2006-01-02 15:04") + "\n")

	if note.LeafletRKey != nil {
		content.WriteString("**RKey:** `" + *note.LeafletRKey + "`\n")
	}

	if note.LeafletCID != nil {
		content.WriteString("**CID:** `" + *note.LeafletCID + "`\n")
	}

	content.WriteString("\n---\n\n")

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

// formatPublicationContent renders markdown with glamour for viewport display
func formatPublicationContent(note *models.Note) (string, error) {
	markdown := buildPublicationMarkdown(note)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return markdown, fmt.Errorf("failed to create renderer: %w", err)
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		return markdown, fmt.Errorf("failed to render markdown: %w", err)
	}

	return rendered, nil
}

// Show displays the publication in interactive mode with pager
func (pv *PublicationView) Show(ctx context.Context) error {
	if pv.opts.Static {
		return pv.staticShow(ctx)
	}

	content, err := formatPublicationContent(pv.note)
	if err != nil {
		return err
	}

	vp := viewport.New(pv.opts.Width-2, pv.opts.Height-6)
	vp.SetContent(content)

	model := publicationViewModel{
		note:     pv.note,
		viewport: vp,
		keys:     publicationViewKeys,
		help:     help.New(),
		opts:     pv.opts,
	}

	program := tea.NewProgram(
		model,
		tea.WithInput(pv.opts.Input),
		tea.WithOutput(pv.opts.Output),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = program.Run()
	return err
}

func (pv *PublicationView) staticShow(context.Context) error {
	content, err := formatPublicationContent(pv.note)
	if err != nil {
		return err
	}

	fmt.Fprint(pv.opts.Output, content)
	return nil
}
