package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type MockDataRecord struct {
	fields map[string]any
}

func (m MockDataRecord) GetID() int64             { return 1 }
func (m MockDataRecord) SetID(id int64)           {}
func (m MockDataRecord) GetTableName() string     { return "mock_records" }
func (m MockDataRecord) GetCreatedAt() time.Time  { return time.Time{} }
func (m MockDataRecord) SetCreatedAt(t time.Time) {}
func (m MockDataRecord) GetUpdatedAt() time.Time  { return time.Time{} }
func (m MockDataRecord) SetUpdatedAt(t time.Time) {}
func (m MockDataRecord) GetField(name string) any { return m.fields[name] }

func NewMockRecord(id int64, fields map[string]any) MockDataRecord {
	return MockDataRecord{fields: fields}
}

type MockDataSource struct {
	records    []DataRecord
	loadError  error
	countError error
}

func (m *MockDataSource) Load(ctx context.Context, opts DataOptions) ([]DataRecord, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}

	filtered := make([]DataRecord, 0)
	for _, record := range m.records {
		include := true
		for filterField, filterValue := range opts.Filters {
			if record.GetField(filterField) != filterValue {
				include = false
				break
			}
		}
		if include {
			filtered = append(filtered, record)
		}
	}

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, nil
}

func (m *MockDataSource) Count(ctx context.Context, opts DataOptions) (int, error) {
	if m.countError != nil {
		return 0, m.countError
	}

	count := 0
	for _, record := range m.records {
		include := true
		for filterField, filterValue := range opts.Filters {
			if record.GetField(filterField) != filterValue {
				include = false
				break
			}
		}
		if include {
			count++
		}
	}

	return count, nil
}

func createMockRecords() []DataRecord {
	return []DataRecord{
		NewMockRecord(1, map[string]any{
			"name":     "John Doe",
			"status":   "active",
			"priority": "high",
			"project":  "alpha",
		}),
		NewMockRecord(2, map[string]any{
			"name":     "Jane Smith",
			"status":   "pending",
			"priority": "medium",
			"project":  "beta",
		}),
		NewMockRecord(3, map[string]any{
			"name":     "Bob Johnson",
			"status":   "completed",
			"priority": "low",
			"project":  "alpha",
		}),
	}
}

func createTestFields() []Field {
	return []Field{
		{Name: "name", Title: "Name", Width: 20},
		{Name: "status", Title: "Status", Width: 12},
		{Name: "priority", Title: "Priority", Width: 10, Formatter: func(v any) string {
			return strings.ToUpper(fmt.Sprintf("%v", v))
		}},
		{Name: "project", Title: "Project", Width: 15},
	}
}

func TestDataTable(t *testing.T) {
	t.Run("TestDataTableOptions", func(t *testing.T) {
		t.Run("default options", func(t *testing.T) {
			source := &MockDataSource{records: createMockRecords()}
			opts := DataTableOptions{
				Fields: createTestFields(),
			}

			table := NewDataTable(source, opts)
			if table.opts.Output == nil {
				t.Error("Output should default to os.Stdout")
			}
			if table.opts.Input == nil {
				t.Error("Input should default to os.Stdin")
			}
			if table.opts.Title != "Data" {
				t.Error("Title should default to 'Data'")
			}
		})

		t.Run("custom options", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockDataSource{records: createMockRecords()}
			opts := DataTableOptions{
				Output: &buf,
				Static: true,
				Title:  "Test Table",
				Fields: createTestFields(),
				ViewHandler: func(record DataRecord) string {
					return fmt.Sprintf("Viewing: %v", record.GetField("name"))
				},
			}

			table := NewDataTable(source, opts)
			if table.opts.Output != &buf {
				t.Error("Custom output not set")
			}
			if !table.opts.Static {
				t.Error("Static mode not set")
			}
			if table.opts.Title != "Test Table" {
				t.Error("Custom title not set")
			}
		})
	})

	t.Run("Static Mode", func(t *testing.T) {
		t.Run("successful static display", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockDataSource{records: createMockRecords()}

			table := NewDataTable(source, DataTableOptions{
				Output: &buf,
				Static: true,
				Title:  "Test Table",
				Fields: createTestFields(),
			})

			err := table.Browse(context.Background())
			if err != nil {
				t.Fatalf("Browse failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Test Table") {
				t.Error("Title not displayed")
			}
			if !strings.Contains(output, "John Doe") {
				t.Error("First record not displayed")
			}
			if !strings.Contains(output, "Jane Smith") {
				t.Error("Second record not displayed")
			}
			if !strings.Contains(output, "Name") {
				t.Error("Header not displayed")
			}
		})

		t.Run("static display with no records", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockDataSource{records: []DataRecord{}}

			table := NewDataTable(source, DataTableOptions{
				Output: &buf,
				Static: true,
				Fields: createTestFields(),
			})

			err := table.Browse(context.Background())
			if err != nil {
				t.Fatalf("Browse failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "No records found") {
				t.Error("No records message not displayed")
			}
		})

		t.Run("static display with load error", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockDataSource{
				loadError: errors.New("database error"),
			}

			table := NewDataTable(source, DataTableOptions{
				Output: &buf,
				Static: true,
				Fields: createTestFields(),
			})

			err := table.Browse(context.Background())
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			output := buf.String()
			if !strings.Contains(output, "Error: database error") {
				t.Error("Error message not displayed")
			}
		})

		t.Run("static display with filters", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockDataSource{records: createMockRecords()}

			table := NewDataTable(source, DataTableOptions{
				Output: &buf,
				Static: true,
				Fields: createTestFields(),
			})

			opts := DataOptions{
				Filters: map[string]any{
					"status": "active",
				},
			}

			err := table.BrowseWithOptions(context.Background(), opts)
			if err != nil {
				t.Fatalf("Browse failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "John Doe") {
				t.Error("Active record not displayed")
			}
			if strings.Contains(output, "Jane Smith") {
				t.Error("Pending record should be filtered out")
			}
		})
	})

	t.Run("Model", func(t *testing.T) {
		t.Run("initial model state", func(t *testing.T) {
			model := dataTableModel{
				opts: DataTableOptions{
					Fields: createTestFields(),
				},
				loading: true,
			}

			if model.selected != 0 {
				t.Error("Initial selected should be 0")
			}
			if model.viewing {
				t.Error("Initial viewing should be false")
			}
			if !model.loading {
				t.Error("Initial loading should be true")
			}
		})

		t.Run("load data command", func(t *testing.T) {
			source := &MockDataSource{records: createMockRecords()}

			model := dataTableModel{
				source:   source,
				keys:     DefaultDataTableKeys(),
				dataOpts: DataOptions{},
			}

			cmd := model.loadData()
			if cmd == nil {
				t.Fatal("loadData should return a command")
			}

			msg := cmd()
			switch msg := msg.(type) {
			case dataLoadedMsg:
				records := []DataRecord(msg)
				if len(records) != 3 {
					t.Errorf("Expected 3 records, got %d", len(records))
				}
			case dataErrorMsg:
				t.Fatalf("Unexpected error: %v", error(msg))
			default:
				t.Fatalf("Unexpected message type: %T", msg)
			}
		})

		t.Run("load data with error", func(t *testing.T) {
			source := &MockDataSource{
				loadError: errors.New("connection failed"),
			}

			model := dataTableModel{
				source:   source,
				dataOpts: DataOptions{},
			}

			cmd := model.loadData()
			msg := cmd()

			switch msg := msg.(type) {
			case dataErrorMsg:
				err := error(msg)
				if !strings.Contains(err.Error(), "connection failed") {
					t.Errorf("Expected connection error, got: %v", err)
				}
			default:
				t.Fatalf("Expected dataErrorMsg, got: %T", msg)
			}
		})

		t.Run("load count command", func(t *testing.T) {
			source := &MockDataSource{records: createMockRecords()}

			model := dataTableModel{
				source:   source,
				dataOpts: DataOptions{},
			}

			cmd := model.loadCount()
			msg := cmd()

			switch msg := msg.(type) {
			case dataCountMsg:
				count := int(msg)
				if count != 3 {
					t.Errorf("Expected count 3, got %d", count)
				}
			default:
				t.Fatalf("Expected dataCountMsg, got: %T", msg)
			}
		})

		t.Run("load count with error", func(t *testing.T) {
			source := &MockDataSource{
				records:    createMockRecords(),
				countError: errors.New("count failed"),
			}

			model := dataTableModel{
				source:   source,
				dataOpts: DataOptions{},
			}

			cmd := model.loadCount()
			msg := cmd()

			switch msg := msg.(type) {
			case dataCountMsg:
				count := int(msg)
				if count != 0 {
					t.Errorf("Expected count 0 on error, got %d", count)
				}
			default:
				t.Fatalf("Expected dataCountMsg even on error, got: %T", msg)
			}
		})

		t.Run("view record command", func(t *testing.T) {
			viewHandler := func(record DataRecord) string {
				return fmt.Sprintf("Viewing: %v", record.GetField("name"))
			}

			model := dataTableModel{
				opts: DataTableOptions{
					ViewHandler: viewHandler,
					Fields:      createTestFields(),
				},
			}

			record := createMockRecords()[0]
			cmd := model.viewRecord(record)
			msg := cmd()

			switch msg := msg.(type) {
			case dataViewMsg:
				content := string(msg)
				if !strings.Contains(content, "Viewing: John Doe") {
					t.Error("View content not formatted correctly")
				}
			default:
				t.Fatalf("Expected dataViewMsg, got: %T", msg)
			}
		})
	})

	t.Run("Key Handling", func(t *testing.T) {
		source := &MockDataSource{records: createMockRecords()}

		t.Run("navigation keys", func(t *testing.T) {
			model := dataTableModel{
				source:   source,
				records:  createMockRecords(),
				selected: 1,
				keys:     DefaultDataTableKeys(),
				opts:     DataTableOptions{Fields: createTestFields()},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.selected != 0 {
					t.Errorf("Up key should move selection to 0, got %d", m.selected)
				}
			}

			model.selected = 1
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.selected != 2 {
					t.Errorf("Down key should move selection to 2, got %d", m.selected)
				}
			}
		})

		t.Run("boundary conditions", func(t *testing.T) {
			model := dataTableModel{
				source:   source,
				records:  createMockRecords(),
				selected: 0,
				keys:     DefaultDataTableKeys(),
				opts:     DataTableOptions{Fields: createTestFields()},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.selected != 0 {
					t.Error("Up key at top should not change selection")
				}
			}

			model.selected = 2
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.selected != 2 {
					t.Error("Down key at bottom should not change selection")
				}
			}
		})

		t.Run("number shortcuts", func(t *testing.T) {
			model := dataTableModel{
				source:  source,
				records: createMockRecords(),
				keys:    DefaultDataTableKeys(),
				opts:    DataTableOptions{Fields: createTestFields()},
			}

			for i := 1; i <= 3; i++ {
				key := fmt.Sprintf("%d", i)
				newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
				if m, ok := newModel.(dataTableModel); ok {
					expectedIndex := i - 1
					if m.selected != expectedIndex {
						t.Errorf("Number key %s should select index %d, got %d", key, expectedIndex, m.selected)
					}
				}
			}
		})

		t.Run("view key with handler", func(t *testing.T) {
			viewHandler := func(record DataRecord) string {
				return "test view"
			}

			model := dataTableModel{
				source:  source,
				records: createMockRecords(),
				keys:    DefaultDataTableKeys(),
				opts: DataTableOptions{
					Fields:      createTestFields(),
					ViewHandler: viewHandler,
				},
			}

			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
			if cmd == nil {
				t.Error("View key should return command when handler is set")
			}
		})

		t.Run("view key without handler", func(t *testing.T) {
			model := dataTableModel{
				source:  source,
				records: createMockRecords(),
				keys:    DefaultDataTableKeys(),
				opts:    DataTableOptions{Fields: createTestFields()},
			}

			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
			if cmd != nil {
				t.Error("View key should not return command when no handler is set")
			}
		})

		t.Run("quit key", func(t *testing.T) {
			model := dataTableModel{
				keys: DefaultDataTableKeys(),
				opts: DataTableOptions{Fields: createTestFields()},
			}

			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
			if cmd == nil {
				t.Error("Quit key should return quit command")
			}
		})

		t.Run("refresh key", func(t *testing.T) {
			model := dataTableModel{
				source: source,
				keys:   DefaultDataTableKeys(),
				opts:   DataTableOptions{Fields: createTestFields()},
			}

			newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
			if cmd == nil {
				t.Error("Refresh key should return command")
			}
			if m, ok := newModel.(dataTableModel); ok {
				if !m.loading {
					t.Error("Refresh should set loading to true")
				}
			}
		})

		t.Run("help mode", func(t *testing.T) {
			model := dataTableModel{
				keys:        DefaultDataTableKeys(),
				showingHelp: true,
				opts:        DataTableOptions{Fields: createTestFields()},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.selected != 0 {
					t.Error("Navigation should be ignored in help mode")
				}
			}

			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.showingHelp {
					t.Error("Help key should exit help mode")
				}
			}
		})

		t.Run("viewing mode", func(t *testing.T) {
			model := dataTableModel{
				keys:        DefaultDataTableKeys(),
				viewing:     true,
				viewContent: "test content",
				opts:        DataTableOptions{Fields: createTestFields()},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
			if m, ok := newModel.(dataTableModel); ok {
				if m.viewing {
					t.Error("Quit should exit viewing mode")
				}
				if m.viewContent != "" {
					t.Error("Quit should clear view content")
				}
			}
		})
	})

	t.Run("View", func(t *testing.T) {
		source := &MockDataSource{records: createMockRecords()}

		t.Run("normal view", func(t *testing.T) {
			model := dataTableModel{
				source:  source,
				records: createMockRecords(),
				keys:    DefaultDataTableKeys(),
				help:    help.New(),
				opts:    DataTableOptions{Title: "Test", Fields: createTestFields()},
			}

			view := model.View()
			if !strings.Contains(view, "Test") {
				t.Error("Title not displayed")
			}
			if !strings.Contains(view, "John Doe") {
				t.Error("Record data not displayed")
			}
			if !strings.Contains(view, "Name") {
				t.Error("Headers not displayed")
			}
			if !strings.Contains(view, " > ") {
				t.Error("Selection indicator not displayed")
			}
		})

		t.Run("loading view", func(t *testing.T) {
			model := dataTableModel{
				loading: true,
				opts:    DataTableOptions{Title: "Test", Fields: createTestFields()},
			}

			view := model.View()
			if !strings.Contains(view, "Loading...") {
				t.Error("Loading message not displayed")
			}
		})

		t.Run("error view", func(t *testing.T) {
			model := dataTableModel{
				err:  errors.New("test error"),
				opts: DataTableOptions{Title: "Test", Fields: createTestFields()},
			}

			view := model.View()
			if !strings.Contains(view, "Error: test error") {
				t.Error("Error message not displayed")
			}
		})

		t.Run("empty records view", func(t *testing.T) {
			model := dataTableModel{
				records: []DataRecord{},
				opts:    DataTableOptions{Title: "Test", Fields: createTestFields()},
			}

			view := model.View()
			if !strings.Contains(view, "No records found") {
				t.Error("Empty message not displayed")
			}
		})

		t.Run("viewing mode", func(t *testing.T) {
			model := dataTableModel{
				viewing:     true,
				viewContent: "# Test Content\nDetails here",
				opts:        DataTableOptions{Fields: createTestFields()},
			}

			view := model.View()
			if !strings.Contains(view, "# Test Content") {
				t.Error("View content not displayed")
			}
			if !strings.Contains(view, "Press q/esc/backspace to return") {
				t.Error("Return instructions not displayed")
			}
		})

		t.Run("help mode", func(t *testing.T) {
			model := dataTableModel{
				showingHelp: true,
				keys:        DefaultDataTableKeys(),
				help:        help.New(),
				opts:        DataTableOptions{Fields: createTestFields()},
			}

			view := model.View()
			if view == "" {
				t.Error("Help view should not be empty")
			}
		})

		t.Run("field formatters", func(t *testing.T) {
			fields := []Field{
				{Name: "priority", Title: "Priority", Width: 10, Formatter: func(v any) string {
					return strings.ToUpper(fmt.Sprintf("%v", v))
				}},
			}

			model := dataTableModel{
				records: createMockRecords(),
				opts:    DataTableOptions{Fields: fields},
			}

			view := model.View()
			if !strings.Contains(view, "HIGH") {
				t.Error("Field formatter not applied")
			}
		})

		t.Run("long field truncation", func(t *testing.T) {
			longRecord := NewMockRecord(1, map[string]any{
				"name": "This is a very long name that should be truncated",
			})

			fields := []Field{
				{Name: "name", Title: "Name", Width: 10},
			}

			model := dataTableModel{
				records: []DataRecord{longRecord},
				opts:    DataTableOptions{Fields: fields},
			}

			view := model.View()
			if !strings.Contains(view, "...") {
				t.Error("Long field should be truncated with ellipsis")
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		source := &MockDataSource{records: createMockRecords()}

		t.Run("data loaded message", func(t *testing.T) {
			model := dataTableModel{
				source:  source,
				loading: true,
				opts:    DataTableOptions{Fields: createTestFields()},
			}

			records := createMockRecords()[:2]
			newModel, _ := model.Update(dataLoadedMsg(records))

			if m, ok := newModel.(dataTableModel); ok {
				if len(m.records) != 2 {
					t.Errorf("Expected 2 records, got %d", len(m.records))
				}
				if m.loading {
					t.Error("Loading should be set to false")
				}
			}
		})

		t.Run("selected index adjustment", func(t *testing.T) {
			model := dataTableModel{
				selected: 5,
				opts:     DataTableOptions{Fields: createTestFields()},
			}

			records := createMockRecords()[:2]
			newModel, _ := model.Update(dataLoadedMsg(records))

			if m, ok := newModel.(dataTableModel); ok {
				if m.selected != 1 {
					t.Errorf("Selected should be adjusted to 1, got %d", m.selected)
				}
			}
		})

		t.Run("data view message", func(t *testing.T) {
			model := dataTableModel{
				opts: DataTableOptions{Fields: createTestFields()},
			}

			content := "Test view content"
			newModel, _ := model.Update(dataViewMsg(content))

			if m, ok := newModel.(dataTableModel); ok {
				if !m.viewing {
					t.Error("Viewing mode should be activated")
				}
				if m.viewContent != content {
					t.Error("View content not set correctly")
				}
			}
		})

		t.Run("data error message", func(t *testing.T) {
			model := dataTableModel{
				loading: true,
				opts:    DataTableOptions{Fields: createTestFields()},
			}

			testErr := errors.New("test error")
			newModel, _ := model.Update(dataErrorMsg(testErr))

			if m, ok := newModel.(dataTableModel); ok {
				if m.err == nil {
					t.Error("Error should be set")
				}
				if m.err.Error() != "test error" {
					t.Errorf("Expected 'test error', got %v", m.err)
				}
				if m.loading {
					t.Error("Loading should be set to false on error")
				}
			}
		})

		t.Run("data count message", func(t *testing.T) {
			model := dataTableModel{
				opts: DataTableOptions{Fields: createTestFields()},
			}

			count := 42
			newModel, _ := model.Update(dataCountMsg(count))

			if m, ok := newModel.(dataTableModel); ok {
				if m.totalCount != count {
					t.Errorf("Expected count %d, got %d", count, m.totalCount)
				}
			}
		})
	})

	t.Run("Default Keys", func(t *testing.T) {
		keys := DefaultDataTableKeys()

		if len(keys.Numbers) != 9 {
			t.Errorf("Expected 9 number bindings, got %d", len(keys.Numbers))
		}

		if keys.Actions == nil {
			t.Error("Actions map should be initialized")
		}
	})

	t.Run("Actions", func(t *testing.T) {
		t.Run("action key handling", func(t *testing.T) {
			actionCalled := false
			action := Action{
				Key:         "d",
				Description: "delete",
				Handler: func(record DataRecord) tea.Cmd {
					actionCalled = true
					return nil
				},
			}

			keys := DefaultDataTableKeys()
			keys.Actions["d"] = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete"))

			model := dataTableModel{
				source:  &MockDataSource{records: createMockRecords()},
				records: createMockRecords(),
				keys:    keys,
				opts: DataTableOptions{
					Fields:  createTestFields(),
					Actions: []Action{action},
				},
			}

			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
			if cmd != nil {
				cmd()
			}

			if !actionCalled {
				t.Error("Action handler should be called")
			}
		})
	})

	t.Run("Field", func(t *testing.T) {
		t.Run("field without formatter", func(t *testing.T) {
			field := Field{Name: "test"}

			record := NewMockRecord(1, map[string]any{
				"test": "value",
			})

			value := record.GetField(field.Name)
			displayValue := fmt.Sprintf("%v", value)

			if displayValue != "value" {
				t.Errorf("Expected 'value', got '%s'", displayValue)
			}
		})

		t.Run("field with formatter", func(t *testing.T) {
			field := Field{
				Name: "test",
				Formatter: func(v any) string {
					return strings.ToUpper(fmt.Sprintf("%v", v))
				},
			}

			record := NewMockRecord(1, map[string]any{
				"test": "value",
			})

			value := record.GetField(field.Name)
			displayValue := field.Formatter(value)

			if displayValue != "VALUE" {
				t.Errorf("Expected 'VALUE', got '%s'", displayValue)
			}
		})
	})
}
