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
	"github.com/stormlightlabs/noteleaf/internal/models"
)

// DataRecord represents a single row of data in a table
type DataRecord interface {
	models.Model
	GetField(name string) any
}

// DataSource provides data for the table
type DataSource interface {
	Load(ctx context.Context, opts DataOptions) ([]DataRecord, error)
	Count(ctx context.Context, opts DataOptions) (int, error)
}

// Field defines a column in the table
type Field struct {
	Name      string
	Title     string
	Width     int
	Formatter func(value any) string
}

// DataOptions configures data loading
type DataOptions struct {
	Filters   map[string]any
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// Action defines an action that can be performed on a record
type Action struct {
	Key         string
	Description string
	Handler     func(record DataRecord) tea.Cmd
}

// DataTableKeyMap defines key bindings for table navigation
type DataTableKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	View    key.Binding
	Refresh key.Binding
	Quit    key.Binding
	Back    key.Binding
	Help    key.Binding
	Numbers []key.Binding
	Actions map[string]key.Binding
}

func (k DataTableKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Help, k.Quit}
}

func (k DataTableKeyMap) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.View},
		{k.Refresh, k.Help, k.Quit, k.Back},
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

// DefaultDataTableKeys returns the default key bindings
func DefaultDataTableKeys() DataTableKeyMap {
	return DataTableKeyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		View:    key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view")),
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

// DataTableOptions configures table behavior
type DataTableOptions struct {
	Output      io.Writer
	Input       io.Reader
	Static      bool
	Title       string
	Fields      []Field
	Actions     []Action
	ViewHandler func(record DataRecord) string
}

// DataTable handles table display and interaction
type DataTable struct {
	source DataSource
	opts   DataTableOptions
}

// NewDataTable creates a new data table
func NewDataTable(source DataSource, opts DataTableOptions) *DataTable {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	if opts.Title == "" {
		opts.Title = "Data"
	}

	return &DataTable{
		source: source,
		opts:   opts,
	}
}

type (
	dataLoadedMsg []DataRecord
	dataViewMsg   string
	dataErrorMsg  error
	dataCountMsg  int
)

type dataTableModel struct {
	records     []DataRecord
	selected    int
	viewing     bool
	viewContent string
	err         error
	loading     bool
	source      DataSource
	opts        DataTableOptions
	keys        DataTableKeyMap
	help        help.Model
	showingHelp bool
	totalCount  int
	dataOpts    DataOptions
}

func (m dataTableModel) Init() tea.Cmd {
	return tea.Batch(m.loadData(), m.loadCount())
}

func (m dataTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.selected < len(m.records)-1 {
				m.selected++
			}
		case key.Matches(msg, m.keys.Enter) || key.Matches(msg, m.keys.View):
			if len(m.records) > 0 && m.selected < len(m.records) && m.opts.ViewHandler != nil {
				return m, m.viewRecord(m.records[m.selected])
			}
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, tea.Batch(m.loadData(), m.loadCount())
		case key.Matches(msg, m.keys.Help):
			m.showingHelp = true
			return m, nil
		default:
			for i, numKey := range m.keys.Numbers {
				if key.Matches(msg, numKey) && i < len(m.records) {
					m.selected = i
					break
				}
			}

			for actionKey, binding := range m.keys.Actions {
				if key.Matches(msg, binding) && len(m.records) > 0 && m.selected < len(m.records) {
					for _, action := range m.opts.Actions {
						if action.Key == actionKey {
							return m, action.Handler(m.records[m.selected])
						}
					}
				}
			}
		}
	case dataLoadedMsg:
		m.records = []DataRecord(msg)
		m.loading = false
		if m.selected >= len(m.records) && len(m.records) > 0 {
			m.selected = len(m.records) - 1
		}
	case dataViewMsg:
		m.viewContent = string(msg)
		m.viewing = true
	case dataErrorMsg:
		m.err = error(msg)
		m.loading = false
	case dataCountMsg:
		m.totalCount = int(msg)
	}
	return m, nil
}

func (m dataTableModel) View() string {
	var s strings.Builder

	style := MutedStyle

	if m.showingHelp {
		return m.help.View(m.keys)
	}

	if m.viewing {
		s.WriteString(m.viewContent)
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press q/esc/backspace to return to list, ? for help"))
		return s.String()
	}

	s.WriteString(TableTitleStyle.Render(m.opts.Title))
	if m.totalCount > 0 {
		s.WriteString(fmt.Sprintf(" (%d total)", m.totalCount))
	}
	s.WriteString("\n\n")

	if m.loading {
		s.WriteString("Loading...")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.records) == 0 {
		s.WriteString("No records found")
		s.WriteString("\n\n")
		s.WriteString(style.Render("Press r to refresh, q to quit"))
		return s.String()
	}

	headerParts := make([]string, len(m.opts.Fields))
	for i, field := range m.opts.Fields {
		format := fmt.Sprintf("%%-%ds", field.Width)
		headerParts[i] = fmt.Sprintf(format, field.Title)
	}
	headerLine := fmt.Sprintf("   %s", strings.Join(headerParts, " "))
	s.WriteString(TableHeaderStyle.Render(headerLine))
	s.WriteString("\n")

	totalWidth := 3 + len(strings.Join(headerParts, " "))
	s.WriteString(TableHeaderStyle.Render(strings.Repeat("─", totalWidth)))
	s.WriteString("\n")

	for i, record := range m.records {
		prefix := "   "
		if i == m.selected {
			prefix = " > "
		}

		rowParts := make([]string, len(m.opts.Fields))
		for j, field := range m.opts.Fields {
			value := record.GetField(field.Name)

			var displayValue string
			if field.Formatter != nil {
				displayValue = field.Formatter(value)
			} else {
				displayValue = fmt.Sprintf("%v", value)
			}

			if len(displayValue) > field.Width-1 {
				displayValue = displayValue[:field.Width-4] + "..."
			}

			format := fmt.Sprintf("%%-%ds", field.Width)
			rowParts[j] = fmt.Sprintf(format, displayValue)
		}

		line := fmt.Sprintf("%s%s", prefix, strings.Join(rowParts, " "))

		if i == m.selected {
			s.WriteString(TableSelectedStyle.Render(line))
		} else {
			s.WriteString(style.Render(line))
		}

		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))

	return s.String()
}

func (m dataTableModel) loadData() tea.Cmd {
	return func() tea.Msg {
		records, err := m.source.Load(context.Background(), m.dataOpts)
		if err != nil {
			return dataErrorMsg(err)
		}
		return dataLoadedMsg(records)
	}
}

func (m dataTableModel) loadCount() tea.Cmd {
	return func() tea.Msg {
		count, err := m.source.Count(context.Background(), m.dataOpts)
		if err != nil {
			return dataCountMsg(0)
		}
		return dataCountMsg(count)
	}
}

func (m dataTableModel) viewRecord(record DataRecord) tea.Cmd {
	return func() tea.Msg {
		content := m.opts.ViewHandler(record)
		return dataViewMsg(content)
	}
}

// Browse opens an interactive table interface
func (dt *DataTable) Browse(ctx context.Context) error {
	return dt.BrowseWithOptions(ctx, DataOptions{})
}

// BrowseWithOptions opens an interactive table with custom data options
func (dt *DataTable) BrowseWithOptions(ctx context.Context, dataOpts DataOptions) error {
	if dt.opts.Static {
		return dt.staticDisplay(ctx, dataOpts)
	}

	keys := DefaultDataTableKeys()
	for _, action := range dt.opts.Actions {
		keys.Actions[action.Key] = key.NewBinding(
			key.WithKeys(action.Key),
			key.WithHelp(action.Key, action.Description),
		)
	}

	model := dataTableModel{
		source:   dt.source,
		opts:     dt.opts,
		keys:     keys,
		help:     help.New(),
		dataOpts: dataOpts,
		loading:  true,
	}

	program := tea.NewProgram(model, tea.WithInput(dt.opts.Input), tea.WithOutput(dt.opts.Output))
	_, err := program.Run()
	return err
}

func (dt *DataTable) staticDisplay(ctx context.Context, dataOpts DataOptions) error {
	records, err := dt.source.Load(ctx, dataOpts)
	if err != nil {
		fmt.Fprintf(dt.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(dt.opts.Output, "%s\n\n", dt.opts.Title)

	if len(records) == 0 {
		fmt.Fprintf(dt.opts.Output, "No records found\n")
		return nil
	}

	headerParts := make([]string, len(dt.opts.Fields))
	for i, field := range dt.opts.Fields {
		format := fmt.Sprintf("%%-%ds", field.Width)
		headerParts[i] = fmt.Sprintf(format, field.Title)
	}
	fmt.Fprintf(dt.opts.Output, "%s\n", strings.Join(headerParts, " "))

	totalWidth := len(strings.Join(headerParts, " "))
	fmt.Fprintf(dt.opts.Output, "%s\n", strings.Repeat("─", totalWidth))

	for _, record := range records {
		rowParts := make([]string, len(dt.opts.Fields))
		for i, field := range dt.opts.Fields {
			value := record.GetField(field.Name)

			var displayValue string
			if field.Formatter != nil {
				displayValue = field.Formatter(value)
			} else {
				displayValue = fmt.Sprintf("%v", value)
			}

			if len(displayValue) > field.Width-1 {
				displayValue = displayValue[:field.Width-4] + "..."
			}

			format := fmt.Sprintf("%%-%ds", field.Width)
			rowParts[i] = fmt.Sprintf(format, displayValue)
		}

		fmt.Fprintf(dt.opts.Output, "%s\n", strings.Join(rowParts, " "))
	}

	return nil
}
