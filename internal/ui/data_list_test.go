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
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type MockListItem struct {
	title       string
	description string
	filterValue string
}

func (m MockListItem) GetID() int64 {
	return 1 // Mock ID
}

func (m MockListItem) SetID(id int64) {
	// Mock - no-op
}

func (m MockListItem) GetTableName() string {
	return "mock_items"
}

func (m MockListItem) GetCreatedAt() time.Time {
	return time.Time{} // Mock - zero time
}

func (m MockListItem) SetCreatedAt(t time.Time) {
	// Mock - no-op
}

func (m MockListItem) GetUpdatedAt() time.Time {
	return time.Time{} // Mock - zero time
}

func (m MockListItem) SetUpdatedAt(t time.Time) {
	// Mock - no-op
}

func (m MockListItem) GetTitle() string {
	return m.title
}

func (m MockListItem) GetDescription() string {
	return m.description
}

func (m MockListItem) GetFilterValue() string {
	return m.filterValue
}

func NewMockItem(id int64, title, description, filterValue string) MockListItem {
	return MockListItem{
		title:       title,
		description: description,
		filterValue: filterValue,
	}
}

type MockListSource struct {
	items       []ListItem
	loadError   error
	countError  error
	searchError error
}

func (m *MockListSource) Load(ctx context.Context, opts ListOptions) ([]ListItem, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}

	filtered := make([]ListItem, 0)
	for _, item := range m.items {
		include := true
		for filterField, filterValue := range opts.Filters {
			if filterField == "title" && item.GetTitle() != filterValue {
				include = false
				break
			}
		}
		if include {
			filtered = append(filtered, item)
		}
	}

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, nil
}

func (m *MockListSource) Count(ctx context.Context, opts ListOptions) (int, error) {
	if m.countError != nil {
		return 0, m.countError
	}

	count := 0
	for _, item := range m.items {
		include := true
		for filterField, filterValue := range opts.Filters {
			if filterField == "title" && item.GetTitle() != filterValue {
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

func (m *MockListSource) Search(ctx context.Context, query string, opts ListOptions) ([]ListItem, error) {
	if m.searchError != nil {
		return nil, m.searchError
	}

	results := make([]ListItem, 0)
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.GetTitle()), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(item.GetDescription()), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(item.GetFilterValue()), strings.ToLower(query)) {
			results = append(results, item)
		}
	}

	return results, nil
}

func createMockItems() []ListItem {
	return []ListItem{
		NewMockItem(1, "First Item", "Description of first item", "item1 tag1"),
		NewMockItem(2, "Second Item", "Description of second item", "item2 tag2"),
		NewMockItem(3, "Third Item", "Description of third item", "item3 tag1"),
	}
}

func TestDataList(t *testing.T) {
	t.Run("Options", func(t *testing.T) {
		t.Run("default options", func(t *testing.T) {
			source := &MockListSource{items: createMockItems()}
			opts := DataListOptions{}

			list := NewDataList(source, opts)
			if list.opts.Output == nil {
				t.Error("Output should default to os.Stdout")
			}
			if list.opts.Input == nil {
				t.Error("Input should default to os.Stdin")
			}
			if list.opts.Title != "Items" {
				t.Error("Title should default to 'Items'")
			}
			if list.opts.ItemRenderer == nil {
				t.Error("ItemRenderer should have a default")
			}
		})

		t.Run("custom options", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockListSource{items: createMockItems()}
			opts := DataListOptions{
				Output:     &buf,
				Static:     true,
				Title:      "Test List",
				ShowSearch: true,
				Searchable: true,
				ViewHandler: func(item ListItem) string {
					return fmt.Sprintf("Viewing: %s", item.GetTitle())
				},
				ItemRenderer: func(item ListItem, selected bool) string {
					prefix := "  "
					if selected {
						prefix = "> "
					}
					return fmt.Sprintf("%s%s", prefix, item.GetTitle())
				},
			}

			list := NewDataList(source, opts)
			if list.opts.Output != &buf {
				t.Error("Custom output not set")
			}
			if !list.opts.Static {
				t.Error("Static mode not set")
			}
			if list.opts.Title != "Test List" {
				t.Error("Custom title not set")
			}
			if !list.opts.ShowSearch {
				t.Error("ShowSearch not set")
			}
			if !list.opts.Searchable {
				t.Error("Searchable not set")
			}
		})
	})

	t.Run("Static Mode", func(t *testing.T) {
		t.Run("successful static display", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockListSource{items: createMockItems()}

			list := NewDataList(source, DataListOptions{
				Output: &buf,
				Static: true,
				Title:  "Test List",
			})

			err := list.Browse(context.Background())
			if err != nil {
				t.Fatalf("Browse failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Test List") {
				t.Error("Title not displayed")
			}
			if !strings.Contains(output, "First Item") {
				t.Error("First item not displayed")
			}
			if !strings.Contains(output, "Second Item") {
				t.Error("Second item not displayed")
			}
		})

		t.Run("static display with no items", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockListSource{items: []ListItem{}}

			list := NewDataList(source, DataListOptions{
				Output: &buf,
				Static: true,
			})

			err := list.Browse(context.Background())
			if err != nil {
				t.Fatalf("Browse failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "No items found") {
				t.Error("No items message not displayed")
			}
		})

		t.Run("static display with load error", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockListSource{
				loadError: errors.New("connection failed"),
			}

			list := NewDataList(source, DataListOptions{
				Output: &buf,
				Static: true,
			})

			err := list.Browse(context.Background())
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			output := buf.String()
			if !strings.Contains(output, "Error: connection failed") {
				t.Error("Error message not displayed")
			}
		})

		t.Run("static display with filters", func(t *testing.T) {
			var buf bytes.Buffer
			source := &MockListSource{items: createMockItems()}

			list := NewDataList(source, DataListOptions{
				Output: &buf,
				Static: true,
			})

			opts := ListOptions{
				Filters: map[string]any{
					"title": "First Item",
				},
			}

			err := list.BrowseWithOptions(context.Background(), opts)
			if err != nil {
				t.Fatalf("Browse failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "First Item") {
				t.Error("Filtered item not displayed")
			}
			if strings.Contains(output, "Second Item") {
				t.Error("Non-matching item should be filtered out")
			}
		})
	})

	t.Run("Model", func(t *testing.T) {
		t.Run("initial model state", func(t *testing.T) {
			model := dataListModel{
				opts: DataListOptions{
					Title: "Test",
				},
				loading: true,
			}

			if model.selected != 0 {
				t.Error("Initial selected should be 0")
			}
			if model.viewing {
				t.Error("Initial viewing should be false")
			}
			if model.searching {
				t.Error("Initial searching should be false")
			}
			if !model.loading {
				t.Error("Initial loading should be true")
			}
		})

		t.Run("load items command", func(t *testing.T) {
			source := &MockListSource{items: createMockItems()}

			model := dataListModel{
				source:   source,
				keys:     DefaultDataListKeys(),
				listOpts: ListOptions{},
			}

			cmd := model.loadItems()
			if cmd == nil {
				t.Fatal("loadItems should return a command")
			}

			msg := cmd()
			switch msg := msg.(type) {
			case listLoadedMsg:
				items := []ListItem(msg)
				if len(items) != 3 {
					t.Errorf("Expected 3 items, got %d", len(items))
				}
			case listErrorMsg:
				t.Fatalf("Unexpected error: %v", error(msg))
			default:
				t.Fatalf("Unexpected message type: %T", msg)
			}
		})

		t.Run("load items with error", func(t *testing.T) {
			source := &MockListSource{
				loadError: errors.New("load failed"),
			}

			model := dataListModel{
				source:   source,
				listOpts: ListOptions{},
			}

			cmd := model.loadItems()
			msg := cmd()

			switch msg := msg.(type) {
			case listErrorMsg:
				err := error(msg)
				if !strings.Contains(err.Error(), "load failed") {
					t.Errorf("Expected load error, got: %v", err)
				}
			default:
				t.Fatalf("Expected listErrorMsg, got: %T", msg)
			}
		})

		t.Run("search items command", func(t *testing.T) {
			source := &MockListSource{items: createMockItems()}

			model := dataListModel{
				source:   source,
				listOpts: ListOptions{},
			}

			cmd := model.searchItems("First")
			if cmd == nil {
				t.Fatal("searchItems should return a command")
			}

			msg := cmd()
			switch msg := msg.(type) {
			case listLoadedMsg:
				items := []ListItem(msg)
				if len(items) != 1 {
					t.Errorf("Expected 1 search result, got %d", len(items))
				}
				if items[0].GetTitle() != "First Item" {
					t.Error("Search should return matching item")
				}
			case listErrorMsg:
				t.Fatalf("Unexpected error: %v", error(msg))
			default:
				t.Fatalf("Unexpected message type: %T", msg)
			}
		})

		t.Run("search items with error", func(t *testing.T) {
			source := &MockListSource{
				items:       createMockItems(),
				searchError: errors.New("search failed"),
			}

			model := dataListModel{
				source:   source,
				listOpts: ListOptions{},
			}

			cmd := model.searchItems("test")
			msg := cmd()

			switch msg := msg.(type) {
			case listErrorMsg:
				err := error(msg)
				if !strings.Contains(err.Error(), "search failed") {
					t.Errorf("Expected search error, got: %v", err)
				}
			default:
				t.Fatalf("Expected listErrorMsg, got: %T", msg)
			}
		})

		t.Run("view item command", func(t *testing.T) {
			viewHandler := func(item ListItem) string {
				return fmt.Sprintf("Viewing: %s", item.GetTitle())
			}

			model := dataListModel{
				opts: DataListOptions{
					ViewHandler: viewHandler,
				},
			}

			item := createMockItems()[0]
			cmd := model.viewItem(item)
			msg := cmd()

			switch msg := msg.(type) {
			case listViewMsg:
				content := string(msg)
				if !strings.Contains(content, "Viewing: First Item") {
					t.Error("View content not formatted correctly")
				}
			default:
				t.Fatalf("Expected listViewMsg, got: %T", msg)
			}
		})
	})

	t.Run("Key Handling", func(t *testing.T) {
		source := &MockListSource{items: createMockItems()}

		t.Run("navigation keys", func(t *testing.T) {
			model := dataListModel{
				source:   source,
				items:    createMockItems(),
				selected: 1,
				keys:     DefaultDataListKeys(),
				opts:     DataListOptions{},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			if m, ok := newModel.(dataListModel); ok {
				if m.selected != 0 {
					t.Errorf("Up key should move selection to 0, got %d", m.selected)
				}
			}

			model.selected = 1
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(dataListModel); ok {
				if m.selected != 2 {
					t.Errorf("Down key should move selection to 2, got %d", m.selected)
				}
			}
		})

		t.Run("boundary conditions", func(t *testing.T) {
			model := dataListModel{
				source:   source,
				items:    createMockItems(),
				selected: 0,
				keys:     DefaultDataListKeys(),
				opts:     DataListOptions{},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			if m, ok := newModel.(dataListModel); ok {
				if m.selected != 0 {
					t.Error("Up key at top should not change selection")
				}
			}

			model.selected = 2
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(dataListModel); ok {
				if m.selected != 2 {
					t.Error("Down key at bottom should not change selection")
				}
			}
		})

		t.Run("search key", func(t *testing.T) {
			model := dataListModel{
				source: source,
				keys:   DefaultDataListKeys(),
				opts: DataListOptions{
					ShowSearch: true,
				},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
			if m, ok := newModel.(dataListModel); ok {
				if !m.searching {
					t.Error("Search key should enable search mode")
				}
				if m.searchQuery != "" {
					t.Error("Search query should be empty initially")
				}
			}
		})

		t.Run("search mode input", func(t *testing.T) {
			model := dataListModel{
				source:    source,
				keys:      DefaultDataListKeys(),
				searching: true,
				opts: DataListOptions{
					Searchable: true,
				},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
			if m, ok := newModel.(dataListModel); ok {
				if m.searchQuery != "a" {
					t.Errorf("Expected search query 'a', got '%s'", m.searchQuery)
				}
			}

			model.searchQuery = "ab"
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("backspace")})
			if m, ok := newModel.(dataListModel); ok {
				if m.searchQuery != "a" {
					t.Errorf("Backspace should remove last character, got '%s'", m.searchQuery)
				}
			}

			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("esc")})
			if m, ok := newModel.(dataListModel); ok {
				if m.searching {
					t.Error("Escape should exit search mode")
				}
			}
		})

		t.Run("view key with handler", func(t *testing.T) {
			viewHandler := func(item ListItem) string {
				return "test view"
			}

			model := dataListModel{
				source: source,
				items:  createMockItems(),
				keys:   DefaultDataListKeys(),
				opts: DataListOptions{
					ViewHandler: viewHandler,
				},
			}

			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
			if cmd == nil {
				t.Error("View key should return command when handler is set")
			}
		})

		t.Run("refresh key", func(t *testing.T) {
			model := dataListModel{
				source: source,
				keys:   DefaultDataListKeys(),
				opts:   DataListOptions{},
			}

			newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
			if cmd == nil {
				t.Error("Refresh key should return command")
			}
			if m, ok := newModel.(dataListModel); ok {
				if !m.loading {
					t.Error("Refresh should set loading to true")
				}
			}
		})

		t.Run("help mode", func(t *testing.T) {
			model := dataListModel{
				keys:        DefaultDataListKeys(),
				showingHelp: true,
				opts:        DataListOptions{},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(dataListModel); ok {
				if m.selected != 0 {
					t.Error("Navigation should be ignored in help mode")
				}
			}

			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			if m, ok := newModel.(dataListModel); ok {
				if m.showingHelp {
					t.Error("Help key should exit help mode")
				}
			}
		})

		t.Run("viewing mode", func(t *testing.T) {
			model := dataListModel{
				keys:        DefaultDataListKeys(),
				viewing:     true,
				viewContent: "test content",
				opts:        DataListOptions{},
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
			if m, ok := newModel.(dataListModel); ok {
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
		source := &MockListSource{items: createMockItems()}

		t.Run("normal view", func(t *testing.T) {
			model := dataListModel{
				source: source,
				items:  createMockItems(),
				keys:   DefaultDataListKeys(),
				help:   help.New(),
				opts: DataListOptions{
					Title:        "Test List",
					ItemRenderer: defaultItemRenderer,
				},
			}

			view := model.View()
			if !strings.Contains(view, "Test List") {
				t.Error("Title not displayed")
			}
			if !strings.Contains(view, "First Item") {
				t.Error("Item data not displayed")
			}
			if !strings.Contains(view, "> ") {
				t.Error("Selection indicator not displayed")
			}
		})

		t.Run("loading view", func(t *testing.T) {
			model := dataListModel{
				loading: true,
				opts:    DataListOptions{Title: "Test"},
			}

			view := model.View()
			if !strings.Contains(view, "Loading...") {
				t.Error("Loading message not displayed")
			}
		})

		t.Run("error view", func(t *testing.T) {
			model := dataListModel{
				err:  errors.New("test error"),
				opts: DataListOptions{Title: "Test"},
			}

			view := model.View()
			if !strings.Contains(view, "Error: test error") {
				t.Error("Error message not displayed")
			}
		})

		t.Run("empty items view", func(t *testing.T) {
			model := dataListModel{
				items: []ListItem{},
				opts:  DataListOptions{Title: "Test"},
			}

			view := model.View()
			if !strings.Contains(view, "No items found") {
				t.Error("Empty message not displayed")
			}
		})

		t.Run("search mode view", func(t *testing.T) {
			model := dataListModel{
				searching:   true,
				searchQuery: "test",
				opts:        DataListOptions{Title: "Test"},
			}

			view := model.View()
			if !strings.Contains(view, "Search: test") {
				t.Error("Search query not displayed")
			}
			if !strings.Contains(view, "Press Enter to search") {
				t.Error("Search instructions not displayed")
			}
		})

		t.Run("viewing mode", func(t *testing.T) {
			vp := viewport.New(80, 20)
			vp.SetContent("# Test Content\nDetails here")

			model := dataListModel{
				viewing:      true,
				viewContent:  "# Test Content\nDetails here",
				viewViewport: vp,
				opts:         DataListOptions{},
			}

			view := model.View()
			if !strings.Contains(view, "# Test Content") {
				t.Error("View content not displayed")
			}
			if !strings.Contains(view, "q/esc: back") {
				t.Error("Return instructions not displayed")
			}
		})

		t.Run("search in title", func(t *testing.T) {
			model := dataListModel{
				items:       createMockItems(),
				searchQuery: "First",
				keys:        DefaultDataListKeys(),
				help:        help.New(),
				opts: DataListOptions{
					Title:        "Test",
					ItemRenderer: defaultItemRenderer,
				},
			}

			view := model.View()
			if !strings.Contains(view, "Search: First") {
				t.Error("Search query should be displayed in title")
			}
		})

		t.Run("custom item renderer", func(t *testing.T) {
			customRenderer := func(item ListItem, selected bool) string {
				if selected {
					return fmt.Sprintf("*** %s ***", item.GetTitle())
				}
				return fmt.Sprintf("    %s", item.GetTitle())
			}

			model := dataListModel{
				items: createMockItems(),
				keys:  DefaultDataListKeys(),
				help:  help.New(),
				opts: DataListOptions{
					ItemRenderer: customRenderer,
				},
			}

			view := model.View()
			if !strings.Contains(view, "*** First Item ***") {
				t.Error("Custom renderer not applied for selected item")
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		source := &MockListSource{items: createMockItems()}

		t.Run("list loaded message", func(t *testing.T) {
			model := dataListModel{
				source:  source,
				loading: true,
				opts:    DataListOptions{},
			}

			items := createMockItems()[:2]
			newModel, _ := model.Update(listLoadedMsg(items))

			if m, ok := newModel.(dataListModel); ok {
				if len(m.items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(m.items))
				}
				if m.loading {
					t.Error("Loading should be set to false")
				}
			}
		})

		t.Run("selected index adjustment", func(t *testing.T) {
			model := dataListModel{
				selected: 5,
				opts:     DataListOptions{},
			}

			items := createMockItems()[:2]
			newModel, _ := model.Update(listLoadedMsg(items))

			if m, ok := newModel.(dataListModel); ok {
				if m.selected != 1 {
					t.Errorf("Selected should be adjusted to 1, got %d", m.selected)
				}
			}
		})

		t.Run("list view message", func(t *testing.T) {
			model := dataListModel{
				opts: DataListOptions{},
			}

			content := "Test view content"
			newModel, _ := model.Update(listViewMsg(content))

			if m, ok := newModel.(dataListModel); ok {
				if !m.viewing {
					t.Error("Viewing mode should be activated")
				}
				if m.viewContent != content {
					t.Error("View content not set correctly")
				}
			}
		})

		t.Run("list error message", func(t *testing.T) {
			model := dataListModel{
				loading: true,
				opts:    DataListOptions{},
			}

			testErr := errors.New("test error")
			newModel, _ := model.Update(listErrorMsg(testErr))

			if m, ok := newModel.(dataListModel); ok {
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

		t.Run("list count message", func(t *testing.T) {
			model := dataListModel{
				opts: DataListOptions{},
			}

			count := 42
			newModel, _ := model.Update(listCountMsg(count))

			if m, ok := newModel.(dataListModel); ok {
				if m.totalCount != count {
					t.Errorf("Expected count %d, got %d", count, m.totalCount)
				}
			}
		})
	})

	t.Run("Default Keys", func(t *testing.T) {
		keys := DefaultDataListKeys()

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
			action := ListAction{
				Key:         "d",
				Description: "delete",
				Handler: func(item ListItem) tea.Cmd {
					actionCalled = true
					return nil
				},
			}

			keys := DefaultDataListKeys()
			keys.Actions["d"] = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete"))

			model := dataListModel{
				source: &MockListSource{items: createMockItems()},
				items:  createMockItems(),
				keys:   keys,
				opts: DataListOptions{
					Actions: []ListAction{action},
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

	t.Run("Default Item Renderer", func(t *testing.T) {
		item := createMockItems()[0]

		t.Run("unselected item", func(t *testing.T) {
			result := defaultItemRenderer(item, false)
			if !strings.HasPrefix(result, "  ") {
				t.Error("Unselected item should have '  ' prefix")
			}
			if !strings.Contains(result, "First Item") {
				t.Error("Item title should be displayed")
			}
			if !strings.Contains(result, "Description of first item") {
				t.Error("Item description should be displayed")
			}
		})

		t.Run("selected item", func(t *testing.T) {
			result := defaultItemRenderer(item, true)
			if !strings.HasPrefix(result, "> ") {
				t.Error("Selected item should have '> ' prefix")
			}
		})

		t.Run("item without description", func(t *testing.T) {
			itemWithoutDesc := NewMockItem(1, "Test", "", "filter")
			result := defaultItemRenderer(itemWithoutDesc, false)
			if strings.Contains(result, " - ") {
				t.Error("Item without description should not have separator")
			}
		})
	})
}
