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
)

// ListItem represents a single item in a list
type ListItem interface {
	models.Model
	GetTitle() string
	GetDescription() string
	GetFilterValue() string
}

// ListSource provides data for the list
type ListSource interface {
	Load(ctx context.Context, opts ListOptions) ([]ListItem, error)
	Count(ctx context.Context, opts ListOptions) (int, error)
	Search(ctx context.Context, query string, opts ListOptions) ([]ListItem, error)
}

// ListOptions configures data loading for lists
type ListOptions struct {
	Filters   map[string]any
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
	Search    string
}

// ListAction defines an action that can be performed on a list item
type ListAction struct {
	Key         string
	Description string
	Handler     func(item ListItem) tea.Cmd
}

// DataListKeyMap defines key bindings for list navigation
type DataListKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	View    key.Binding
	Search  key.Binding
	Refresh key.Binding
	Quit    key.Binding
	Back    key.Binding
	Help    key.Binding
	Numbers []key.Binding
	Actions map[string]key.Binding
}

func (k DataListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Search, k.Help, k.Quit}
}

func (k DataListKeyMap) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.View},
		{k.Search, k.Refresh, k.Help, k.Quit, k.Back},
	}

	if len(k.Actions) > 0 {
		actionBindings := make([]key.Binding, 0, len(k.Actions))
		for _, binding := range k.Actions {
			actionBindings = append(actionBindings, binding)
		}
		bindings = append(bindings, actionBindings)
	}

	return bindings
}

// DefaultDataListKeys returns the default key bindings for lists
func DefaultDataListKeys() DataListKeyMap {
	return DataListKeyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		View:    key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view")),
		Search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Back:    key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("esc", "back")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Numbers: []key.Binding{
			key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "jump to 1")),
			key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "jump to 2")),
			key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "jump to 3")),
			key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "jump to 4")),
			key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "jump to 5")),
			key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "jump to 6")),
			key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "jump to 7")),
			key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "jump to 8")),
			key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "jump to 9")),
		},
		Actions: make(map[string]key.Binding),
	}
}

// DataListOptions configures list behavior
type DataListOptions struct {
	Output       io.Writer
	Input        io.Reader
	Static       bool
	Title        string
	Actions      []ListAction
	ViewHandler  func(item ListItem) string
	ItemRenderer func(item ListItem, selected bool) string
	ShowSearch   bool
	Searchable   bool
}

// DataList handles list display and interaction
type DataList struct {
	source ListSource
	opts   DataListOptions
}

// NewDataList creates a new data list
func NewDataList(source ListSource, opts DataListOptions) *DataList {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	if opts.Title == "" {
		opts.Title = "Items"
	}
	if opts.ItemRenderer == nil {
		opts.ItemRenderer = defaultItemRenderer
	}

	return &DataList{
		source: source,
		opts:   opts,
	}
}

type (
	listLoadedMsg []ListItem
	listViewMsg   string
	listErrorMsg  error
	listCountMsg  int
)

type dataListModel struct {
	items        []ListItem
	selected     int
	viewing      bool
	viewContent  string
	viewViewport viewport.Model
	searching    bool
	searchQuery  string
	err          error
	loading      bool
	source       ListSource
	opts         DataListOptions
	keys         DataListKeyMap
	help         help.Model
	showingHelp  bool
	totalCount   int
	listOpts     ListOptions
}

func (m dataListModel) Init() tea.Cmd {
	return tea.Batch(m.loadItems(), m.loadCount())
}

func (m dataListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case key.Matches(msg, m.keys.Up):
				m.viewViewport.ScrollUp(1)
			case key.Matches(msg, m.keys.Down):
				m.viewViewport.ScrollDown(1)
			case msg.String() == "pgup", msg.String() == "b":
				m.viewViewport.HalfPageUp()
			case msg.String() == "pgdown", msg.String() == "f":
				m.viewViewport.HalfPageDown()
			case msg.String() == "g", msg.String() == "home":
				m.viewViewport.GotoTop()
			case msg.String() == "G", msg.String() == "end":
				m.viewViewport.GotoBottom()
			}
			return m, nil
		}

		if m.searching {
			switch msg.String() {
			case "esc", "enter":
				m.searching = false
				if msg.String() == "enter" && m.opts.Searchable {
					m.loading = true
					return m, m.searchItems(m.searchQuery)
				}
				return m, nil
			case "backspace", "ctrl+h":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
				return m, nil
			default:
				if len(msg.Runes) > 0 && msg.Runes[0] >= 32 {
					m.searchQuery += string(msg.Runes)
				}
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, m.keys.Down):
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case key.Matches(msg, m.keys.Enter) || key.Matches(msg, m.keys.View):
			if len(m.items) > 0 && m.selected < len(m.items) && m.opts.ViewHandler != nil {
				return m, m.viewItem(m.items[m.selected])
			}
		case key.Matches(msg, m.keys.Search):
			if m.opts.ShowSearch {
				m.searching = true
				m.searchQuery = ""
				return m, nil
			}
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, tea.Batch(m.loadItems(), m.loadCount())
		case key.Matches(msg, m.keys.Help):
			m.showingHelp = true
			return m, nil
		default:
			for i, numKey := range m.keys.Numbers {
				if key.Matches(msg, numKey) && i < len(m.items) {
					m.selected = i
					break
				}
			}

			for actionKey, binding := range m.keys.Actions {
				if key.Matches(msg, binding) && len(m.items) > 0 && m.selected < len(m.items) {
					for _, action := range m.opts.Actions {
						if action.Key == actionKey {
							return m, action.Handler(m.items[m.selected])
						}
					}
				}
			}
		}
	case listLoadedMsg:
		m.items = []ListItem(msg)
		m.loading = false
		if m.selected >= len(m.items) && len(m.items) > 0 {
			m.selected = len(m.items) - 1
		}
	case listViewMsg:
		m.viewContent = string(msg)
		m.viewing = true
		m.viewViewport = viewport.New(80, 20)
		m.viewViewport.SetContent(m.viewContent)
	case listErrorMsg:
		m.err = error(msg)
		m.loading = false
	case listCountMsg:
		m.totalCount = int(msg)
	case tea.WindowSizeMsg:
		if m.viewing {
			headerHeight := 2
			footerHeight := 3
			verticalMarginHeight := headerHeight + footerHeight
			m.viewViewport.Width = msg.Width
			m.viewViewport.Height = msg.Height - verticalMarginHeight
		}
	}
	return m, nil
}

func (m dataListModel) View() string {
	var s strings.Builder

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(Squid.Hex()))

	if m.showingHelp {
		return m.help.View(m.keys)
	}

	if m.viewing {
		s.WriteString(m.viewViewport.View())
		s.WriteString("\n\n")
		s.WriteString(style.Render("↑/↓/pgup/pgdn: scroll | g/G: top/bottom | q/esc: back | ?: help"))
		return s.String()
	}

	s.WriteString(TitleColorStyle.Render(m.opts.Title))
	if m.totalCount > 0 {
		s.WriteString(fmt.Sprintf(" (%d total)", m.totalCount))
	}
	if m.searchQuery != "" {
		s.WriteString(fmt.Sprintf(" - Search: %s", m.searchQuery))
	}
	s.WriteString("\n\n")

	if m.searching {
		s.WriteString("Search: " + m.searchQuery + "▎")
		s.WriteString("\n")
		s.WriteString(style.Render("Press Enter to search, Esc to cancel"))
		return s.String()
	}

	if m.loading {
		s.WriteString("Loading...")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.items) == 0 {
		message := "No items found"
		if m.searchQuery != "" {
			message = "No items found for search: " + m.searchQuery
		}
		s.WriteString(message)
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press r to refresh, q to quit"))
		if m.opts.ShowSearch {
			s.WriteString(style.Render(", / to search"))
		}
		return s.String()
	}

	for i, item := range m.items {
		selected := i == m.selected
		itemView := m.opts.ItemRenderer(item, selected)
		s.WriteString(itemView)
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))

	return s.String()
}

func (m dataListModel) loadItems() tea.Cmd {
	return func() tea.Msg {
		items, err := m.source.Load(context.Background(), m.listOpts)
		if err != nil {
			return listErrorMsg(err)
		}
		return listLoadedMsg(items)
	}
}

func (m dataListModel) loadCount() tea.Cmd {
	return func() tea.Msg {
		count, err := m.source.Count(context.Background(), m.listOpts)
		if err != nil {
			return listCountMsg(0)
		}
		return listCountMsg(count)
	}
}

func (m dataListModel) searchItems(query string) tea.Cmd {
	return func() tea.Msg {
		items, err := m.source.Search(context.Background(), query, m.listOpts)
		if err != nil {
			return listErrorMsg(err)
		}
		return listLoadedMsg(items)
	}
}

func (m dataListModel) viewItem(item ListItem) tea.Cmd {
	return func() tea.Msg {
		content := m.opts.ViewHandler(item)
		return listViewMsg(content)
	}
}

// defaultItemRenderer provides a default rendering for list items
func defaultItemRenderer(item ListItem, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}

	title := item.GetTitle()
	description := item.GetDescription()

	line := fmt.Sprintf("%s%s", prefix, title)
	if description != "" {
		line += fmt.Sprintf(" - %s", description)
	}

	if selected {
		return SelectedColorStyle.Render(line)
	}
	return line
}

// Browse opens an interactive list interface
func (dl *DataList) Browse(ctx context.Context) error {
	return dl.BrowseWithOptions(ctx, ListOptions{})
}

// BrowseWithOptions opens an interactive list with custom options
func (dl *DataList) BrowseWithOptions(ctx context.Context, listOpts ListOptions) error {
	if dl.opts.Static {
		return dl.staticDisplay(ctx, listOpts)
	}

	keys := DefaultDataListKeys()
	for _, action := range dl.opts.Actions {
		keys.Actions[action.Key] = key.NewBinding(
			key.WithKeys(action.Key),
			key.WithHelp(action.Key, action.Description),
		)
	}

	model := dataListModel{
		source:   dl.source,
		opts:     dl.opts,
		keys:     keys,
		help:     help.New(),
		listOpts: listOpts,
		loading:  true,
	}

	program := tea.NewProgram(model, tea.WithInput(dl.opts.Input), tea.WithOutput(dl.opts.Output))
	_, err := program.Run()
	return err
}

func (dl *DataList) staticDisplay(ctx context.Context, listOpts ListOptions) error {
	items, err := dl.source.Load(ctx, listOpts)
	if err != nil {
		fmt.Fprintf(dl.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(dl.opts.Output, "%s\n\n", dl.opts.Title)

	if len(items) == 0 {
		fmt.Fprintf(dl.opts.Output, "No items found\n")
		return nil
	}

	for _, item := range items {
		itemView := dl.opts.ItemRenderer(item, false)
		fmt.Fprintf(dl.opts.Output, "%s\n", itemView)
	}

	return nil
}
