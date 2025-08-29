package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestGetStatusSymbol(t *testing.T) {
	testCases := []struct {
		status   string
		expected string
	}{
		{models.StatusTodo, StatusTodoSymbol},
		{models.StatusInProgress, StatusInProgressSymbol},
		{models.StatusBlocked, StatusBlockedSymbol},
		{models.StatusDone, StatusDoneSymbol},
		{models.StatusAbandoned, StatusAbandonedSymbol},
		{models.StatusPending, StatusPendingSymbol},
		{models.StatusCompleted, StatusCompletedSymbol},
		{models.StatusDeleted, StatusDeletedSymbol},
		{"unknown", StatusPendingSymbol},
	}

	for _, tc := range testCases {
		t.Run("status_"+tc.status, func(t *testing.T) {
			result := GetStatusSymbol(tc.status)
			if result != tc.expected {
				t.Errorf("Expected symbol %s for status %s, got %s", tc.expected, tc.status, result)
			}
		})
	}
}

func TestGetStatusStyle(t *testing.T) {
	testCases := []struct {
		status string
		style  lipgloss.Style
	}{
		{models.StatusTodo, TodoStyle},
		{models.StatusInProgress, InProgressStyle},
		{models.StatusBlocked, BlockedStyle},
		{models.StatusDone, DoneStyle},
		{models.StatusAbandoned, AbandonedStyle},
		{models.StatusPending, PendingStyle},
		{models.StatusCompleted, CompletedStyle},
		{models.StatusDeleted, DeletedStyle},
		{"unknown", PendingStyle},
	}

	for _, tc := range testCases {
		t.Run("style_"+tc.status, func(t *testing.T) {
			result := GetStatusStyle(tc.status)
			expectedColor := tc.style.GetForeground()
			resultColor := result.GetForeground()
			if expectedColor != resultColor {
				t.Errorf("Expected color %s for status %s, got %s", expectedColor, tc.status, resultColor)
			}
		})
	}
}

func TestFormatStatusIndicator(t *testing.T) {
	testCases := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
		models.StatusAbandoned,
	}

	for _, status := range testCases {
		t.Run("format_indicator_"+status, func(t *testing.T) {
			result := FormatStatusIndicator(status)
			expectedSymbol := GetStatusSymbol(status)

			if !strings.Contains(result, expectedSymbol) {
				t.Errorf("Expected formatted indicator for %s to contain symbol %s", status, expectedSymbol)
			}

			if result == "" {
				t.Errorf("Expected non-empty formatted indicator for status %s", status)
			}
		})
	}
}

func TestFormatStatusWithText(t *testing.T) {
	testCases := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
		models.StatusAbandoned,
	}

	for _, status := range testCases {
		t.Run("format_with_text_"+status, func(t *testing.T) {
			result := FormatStatusWithText(status)
			expectedSymbol := GetStatusSymbol(status)

			if !strings.Contains(result, expectedSymbol) {
				t.Errorf("Expected formatted status for %s to contain symbol %s", status, expectedSymbol)
			}
			if !strings.Contains(result, status) {
				t.Errorf("Expected formatted status for %s to contain status text", status)
			}

			if result == "" {
				t.Errorf("Expected non-empty formatted status for %s", status)
			}
		})
	}
}

func TestGetStatusDescription(t *testing.T) {
	testCases := []struct {
		status      string
		description string
	}{
		{models.StatusTodo, "Ready to start"},
		{models.StatusInProgress, "Currently working"},
		{models.StatusBlocked, "Waiting on dependency"},
		{models.StatusDone, "Completed successfully"},
		{models.StatusAbandoned, "No longer relevant"},
		{models.StatusPending, "Pending (legacy)"},
		{models.StatusCompleted, "Completed (legacy)"},
		{models.StatusDeleted, "Deleted (legacy)"},
		{"unknown", "Unknown status"},
	}

	for _, tc := range testCases {
		t.Run("description_"+tc.status, func(t *testing.T) {
			result := GetStatusDescription(tc.status)
			if result != tc.description {
				t.Errorf("Expected description %s for status %s, got %s", tc.description, tc.status, result)
			}
		})
	}
}

func TestFormatTaskStatus(t *testing.T) {
	t.Run("nil_task", func(t *testing.T) {
		result := FormatTaskStatus(nil)
		if result != "" {
			t.Errorf("Expected empty string for nil task, got %s", result)
		}
	})

	testCases := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
		models.StatusAbandoned,
	}

	for _, status := range testCases {
		t.Run("format_task_status_"+status, func(t *testing.T) {
			task := &models.Task{
				ID:     1,
				Status: status,
			}

			result := FormatTaskStatus(task)
			expectedSymbol := GetStatusSymbol(status)
			expectedDescription := GetStatusDescription(status)

			if !strings.Contains(result, expectedSymbol) {
				t.Errorf("Expected task status format to contain symbol %s", expectedSymbol)
			}
			if !strings.Contains(result, status) {
				t.Errorf("Expected task status format to contain status %s", status)
			}
			if !strings.Contains(result, expectedDescription) {
				t.Errorf("Expected task status format to contain description %s", expectedDescription)
			}

			if result == "" {
				t.Errorf("Expected non-empty formatted task status for %s", status)
			}
		})
	}
}

func TestStatusLegend(t *testing.T) {
	result := StatusLegend()

	if result == "" {
		t.Error("Expected non-empty status legend")
	}

	expectedStatuses := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
		models.StatusAbandoned,
	}

	for _, status := range expectedStatuses {
		expectedSymbol := GetStatusSymbol(status)
		if !strings.Contains(result, expectedSymbol) {
			t.Errorf("Expected legend to contain symbol %s for status %s", expectedSymbol, status)
		}
		if !strings.Contains(result, status) {
			t.Errorf("Expected legend to contain status text %s", status)
		}
	}
}

func TestGetAllStatusSymbols(t *testing.T) {
	symbols := GetAllStatusSymbols()

	expectedSymbols := map[string]string{
		models.StatusTodo:       StatusTodoSymbol,
		models.StatusInProgress: StatusInProgressSymbol,
		models.StatusBlocked:    StatusBlockedSymbol,
		models.StatusDone:       StatusDoneSymbol,
		models.StatusAbandoned:  StatusAbandonedSymbol,
		models.StatusPending:    StatusPendingSymbol,
		models.StatusCompleted:  StatusCompletedSymbol,
		models.StatusDeleted:    StatusDeletedSymbol,
	}

	if len(symbols) != len(expectedSymbols) {
		t.Errorf("Expected %d status symbols, got %d", len(expectedSymbols), len(symbols))
	}

	for status, expectedSymbol := range expectedSymbols {
		if symbol, exists := symbols[status]; !exists {
			t.Errorf("Expected status %s to exist in symbols map", status)
		} else if symbol != expectedSymbol {
			t.Errorf("Expected symbol %s for status %s, got %s", expectedSymbol, status, symbol)
		}
	}
}

func TestIsValidStatusTransition(t *testing.T) {
	testCases := []struct {
		from     string
		to       string
		expected bool
	}{
		{models.StatusTodo, models.StatusInProgress, true},
		{models.StatusTodo, models.StatusBlocked, true},
		{models.StatusTodo, models.StatusDone, true},
		{models.StatusTodo, models.StatusAbandoned, true},
		{models.StatusTodo, models.StatusTodo, false},

		{models.StatusInProgress, models.StatusTodo, true},
		{models.StatusInProgress, models.StatusBlocked, true},
		{models.StatusInProgress, models.StatusDone, true},
		{models.StatusInProgress, models.StatusAbandoned, true},
		{models.StatusInProgress, models.StatusInProgress, false},

		{models.StatusBlocked, models.StatusTodo, true},
		{models.StatusBlocked, models.StatusInProgress, true},
		{models.StatusBlocked, models.StatusDone, true},
		{models.StatusBlocked, models.StatusAbandoned, true},
		{models.StatusBlocked, models.StatusBlocked, false},

		{models.StatusDone, models.StatusTodo, true},
		{models.StatusDone, models.StatusInProgress, true},
		{models.StatusDone, models.StatusBlocked, false},
		{models.StatusDone, models.StatusAbandoned, false},
		{models.StatusDone, models.StatusDone, false},

		{models.StatusAbandoned, models.StatusTodo, true},
		{models.StatusAbandoned, models.StatusInProgress, true},
		{models.StatusAbandoned, models.StatusBlocked, false},
		{models.StatusAbandoned, models.StatusDone, false},
		{models.StatusAbandoned, models.StatusAbandoned, false},

		{models.StatusPending, models.StatusTodo, true},
		{models.StatusPending, models.StatusInProgress, true},
		{models.StatusPending, models.StatusBlocked, false},
		{models.StatusCompleted, models.StatusDone, true},
		{models.StatusCompleted, models.StatusTodo, false},

		{"unknown", models.StatusTodo, false},
		{models.StatusTodo, "unknown", false},
	}

	for _, tc := range testCases {
		t.Run("transition_"+tc.from+"_to_"+tc.to, func(t *testing.T) {
			result := IsValidStatusTransition(tc.from, tc.to)
			if result != tc.expected {
				t.Errorf("Expected transition from %s to %s to be %v, got %v",
					tc.from, tc.to, tc.expected, result)
			}
		})
	}
}

func TestUnicodeSymbolConstants(t *testing.T) {
	symbols := []struct {
		name   string
		symbol string
		code   string
	}{
		{"TodoSymbol", StatusTodoSymbol, "●"},
		{"InProgressSymbol", StatusInProgressSymbol, "◐"},
		{"BlockedSymbol", StatusBlockedSymbol, "■"},
		{"DoneSymbol", StatusDoneSymbol, "✓"},
		{"AbandonedSymbol", StatusAbandonedSymbol, "⚫"},
		{"PendingSymbol", StatusPendingSymbol, "○"},
		{"CompletedSymbol", StatusCompletedSymbol, "✓"},
		{"DeletedSymbol", StatusDeletedSymbol, "✗"},
	}

	for _, s := range symbols {
		t.Run("symbol_"+s.name, func(t *testing.T) {
			if s.symbol != s.code {
				t.Errorf("Expected %s to be %s, got %s", s.name, s.code, s.symbol)
			}
		})
	}

	symbolMap := make(map[string][]string)
	for _, s := range symbols {
		symbolMap[s.symbol] = append(symbolMap[s.symbol], s.name)
	}

	for symbol, names := range symbolMap {
		if len(names) > 1 && symbol != "✓" {
			t.Errorf("Symbol %s is used by multiple constants: %v", symbol, names)
		}
	}
}

func TestStyleConstants(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"TodoStyle", TodoStyle},
		{"InProgressStyle", InProgressStyle},
		{"BlockedStyle", BlockedStyle},
		{"DoneStyle", DoneStyle},
		{"AbandonedStyle", AbandonedStyle},
		{"PendingStyle", PendingStyle},
		{"CompletedStyle", CompletedStyle},
		{"DeletedStyle", DeletedStyle},
	}

	for _, s := range styles {
		t.Run("style_"+s.name, func(t *testing.T) {
			testText := "test"
			rendered := s.style.Render(testText)
			// In test environment without terminal, just check that rendering produces output
			if rendered == "" {
				t.Errorf("Expected %s to produce output when rendering text", s.name)
			}
		})
	}
}

// Priority UI Tests

func TestGetPrioritySymbol(t *testing.T) {
	testCases := []struct {
		priority string
		expected string
	}{
		{models.PriorityHigh, PriorityHighSymbol},
		{models.PriorityMedium, PriorityMediumSymbol},
		{models.PriorityLow, PriorityLowSymbol},
		{"", PriorityNoneSymbol},
		{"A", PriorityHighSymbol},
		{"Z", PriorityHighSymbol},
		{"1", PriorityLowSymbol},
		{"2", PriorityMediumSymbol},
		{"3", PriorityMediumSymbol},
		{"4", PriorityHighSymbol},
		{"5", PriorityHighSymbol},
		{"unknown", PriorityNoneSymbol},
	}

	for _, tc := range testCases {
		t.Run("priority_symbol_"+tc.priority, func(t *testing.T) {
			result := GetPrioritySymbol(tc.priority)
			if result != tc.expected {
				t.Errorf("Expected symbol %s for priority %s, got %s", tc.expected, tc.priority, result)
			}
		})
	}
}

func TestGetPriorityPattern(t *testing.T) {
	testCases := []struct {
		priority string
		expected string
	}{
		{models.PriorityHigh, PriorityHighPattern},
		{models.PriorityMedium, PriorityMediumPattern},
		{models.PriorityLow, PriorityLowPattern},
		{"", PriorityNonePattern},
		{"A", PriorityHighPattern},
		{"C", PriorityHighPattern},
		{"M", PriorityMediumPattern},
		{"Z", PriorityLowPattern},
		{"1", PriorityLowPattern},
		{"2", PriorityMediumPattern},
		{"3", PriorityMediumPattern},
		{"4", PriorityHighPattern},
		{"5", PriorityHighPattern},
		{"unknown", PriorityNonePattern},
	}

	for _, tc := range testCases {
		t.Run("priority_pattern_"+tc.priority, func(t *testing.T) {
			result := GetPriorityPattern(tc.priority)
			if result != tc.expected {
				t.Errorf("Expected pattern %s for priority %s, got %s", tc.expected, tc.priority, result)
			}
		})
	}
}

func TestGetPriorityStyle(t *testing.T) {
	testCases := []struct {
		priority string
		style    lipgloss.Style
	}{
		{models.PriorityHigh, PriorityHighStyle},
		{models.PriorityMedium, PriorityMediumStyle},
		{models.PriorityLow, PriorityLowStyle},
		{"", PriorityNoneStyle},
		{"A", PriorityLegacyStyle},
		{"1", PriorityLegacyStyle},
		{"unknown", PriorityNoneStyle},
	}

	for _, tc := range testCases {
		t.Run("priority_style_"+tc.priority, func(t *testing.T) {
			result := GetPriorityStyle(tc.priority)
			expectedColor := tc.style.GetForeground()
			resultColor := result.GetForeground()
			if expectedColor != resultColor {
				t.Errorf("Expected color %s for priority %s, got %s", expectedColor, tc.priority, resultColor)
			}
		})
	}
}

func TestFormatPriorityIndicator(t *testing.T) {
	testCases := []string{models.PriorityHigh, models.PriorityMedium, models.PriorityLow, "", "A", "1"}

	for _, priority := range testCases {
		t.Run("format_priority_indicator_"+priority, func(t *testing.T) {
			got := FormatPriorityIndicator(priority)
			want := GetPriorityPattern(priority)

			if !strings.Contains(got, want) {
				t.Errorf("Expected formatted priority indicator for %s to contain pattern %s", priority, want)
			}

			if got == "" {
				t.Errorf("Expected non-empty formatted priority indicator for priority %s", priority)
			}
		})
	}
}

func TestFormatPriorityWithText(t *testing.T) {
	testCases := []struct {
		priority      string
		shouldContain []string
	}{
		{models.PriorityHigh, []string{PriorityHighPattern, models.PriorityHigh}},
		{models.PriorityMedium, []string{PriorityMediumPattern, models.PriorityMedium}},
		{models.PriorityLow, []string{PriorityLowPattern, models.PriorityLow}},
		{"", []string{PriorityNonePattern, "None"}},
		{"A", []string{PriorityHighPattern, "A"}},
		{"1", []string{PriorityLowPattern, "1"}},
	}

	for _, tc := range testCases {
		t.Run("format_priority_with_text_"+tc.priority, func(t *testing.T) {
			got := FormatPriorityWithText(tc.priority)

			for _, want := range tc.shouldContain {
				if !strings.Contains(got, want) {
					t.Errorf("Expected formatted priority for %s to contain %s, got %s", tc.priority, want, got)
				}
			}

			if got == "" {
				t.Errorf("Expected non-empty formatted priority for %s", tc.priority)
			}
		})
	}
}

func TestGetPriorityDescription(t *testing.T) {
	testCases := []struct {
		priority    string
		description string
	}{
		{models.PriorityHigh, "Urgent - do first"},
		{models.PriorityMedium, "Important - schedule soon"},
		{models.PriorityLow, "Nice to have - when time permits"},
		{"", "No priority set"},
		{"A", "Priority A (legacy)"},
		{"Z", "Priority Z (legacy)"},
		{"1", "Priority 1 (lowest)"},
		{"2", "Priority 2 (low)"},
		{"3", "Priority 3 (medium)"},
		{"4", "Priority 4 (high)"},
		{"5", "Priority 5 (highest)"},
		{"unknown", "Unknown priority"},
	}

	for _, tc := range testCases {
		t.Run("priority_description_"+tc.priority, func(t *testing.T) {
			got := GetPriorityDescription(tc.priority)
			want := tc.description
			if got != want {
				t.Errorf("Expected description %s for priority %s, got %s", want, tc.priority, got)
			}
		})
	}
}

func TestFormatTaskPriority(t *testing.T) {
	t.Run("nil_task", func(t *testing.T) {
		got := FormatTaskPriority(nil)
		if got != "" {
			t.Errorf("Expected empty string for nil task, got %s", got)
		}
	})

	testCases := []struct {
		priority      string
		shouldContain []string
	}{
		{models.PriorityHigh, []string{PriorityHighPattern, models.PriorityHigh, "Urgent - do first"}},
		{models.PriorityMedium, []string{PriorityMediumPattern, models.PriorityMedium, "Important - schedule soon"}},
		{models.PriorityLow, []string{PriorityLowPattern, models.PriorityLow, "Nice to have - when time permits"}},
		{"", []string{PriorityNonePattern, "No priority set"}},
		{"A", []string{PriorityHighPattern, "A", "Priority A (legacy)"}},
		{"1", []string{PriorityLowPattern, "1", "Priority 1 (lowest)"}},
	}

	for _, tc := range testCases {
		t.Run("format_task_priority_"+tc.priority, func(t *testing.T) {
			task := &models.Task{ID: 1, Priority: tc.priority}
			got := FormatTaskPriority(task)

			for _, want := range tc.shouldContain {
				if !strings.Contains(got, want) {
					t.Errorf("Expected task priority format to contain %s, got %s", want, got)
				}
			}

			if got == "" {
				t.Errorf("Expected non-empty formatted task priority for %s", tc.priority)
			}
		})
	}
}

func TestPriorityLegend(t *testing.T) {
	got := PriorityLegend()

	if got == "" {
		t.Error("Expected non-empty priority legend")
	}

	expectedPriorities := []string{models.PriorityHigh, models.PriorityMedium, models.PriorityLow, "None"}
	expectedPatterns := []string{PriorityHighPattern, PriorityMediumPattern, PriorityLowPattern, PriorityNonePattern}

	for _, want := range expectedPriorities {
		if !strings.Contains(got, want) {
			t.Errorf("Expected legend to contain priority text %s", want)
		}
	}

	for _, want := range expectedPatterns {
		if !strings.Contains(got, want) {
			t.Errorf("Expected legend to contain pattern %s", want)
		}
	}
}

func TestGetAllPrioritySymbols(t *testing.T) {
	symbols := GetAllPrioritySymbols()

	expectedSymbols := map[string]string{
		models.PriorityHigh:   PriorityHighSymbol,
		models.PriorityMedium: PriorityMediumSymbol,
		models.PriorityLow:    PriorityLowSymbol,
		"":                    PriorityNoneSymbol,
	}

	if len(symbols) != len(expectedSymbols) {
		t.Errorf("Expected %d priority symbols, got %d", len(expectedSymbols), len(symbols))
	}

	for priority, expectedSymbol := range expectedSymbols {
		if symbol, exists := symbols[priority]; !exists {
			t.Errorf("Expected priority %s to exist in symbols map", priority)
		} else if symbol != expectedSymbol {
			t.Errorf("Expected symbol %s for priority %s, got %s", expectedSymbol, priority, symbol)
		}
	}
}

func TestGetAllPriorityPatterns(t *testing.T) {
	patterns := GetAllPriorityPatterns()

	expectedPatterns := map[string]string{
		models.PriorityHigh:   PriorityHighPattern,
		models.PriorityMedium: PriorityMediumPattern,
		models.PriorityLow:    PriorityLowPattern,
		"":                    PriorityNonePattern,
	}

	if len(patterns) != len(expectedPatterns) {
		t.Errorf("Expected %d priority patterns, got %d", len(expectedPatterns), len(patterns))
	}

	for priority, want := range expectedPatterns {
		if got, exists := patterns[priority]; !exists {
			t.Errorf("Expected priority %s to exist in patterns map", priority)
		} else if got != want {
			t.Errorf("Expected pattern %s for priority %s, got %s", want, priority, got)
		}
	}
}

func TestGetPriorityDisplayType(t *testing.T) {
	testCases := []struct {
		priority string
		expected string
	}{
		{models.PriorityHigh, "text"},
		{models.PriorityMedium, "text"},
		{models.PriorityLow, "text"},
		{"1", "numeric"},
		{"2", "numeric"},
		{"3", "numeric"},
		{"4", "numeric"},
		{"5", "numeric"},
		{"A", "legacy"},
		{"Z", "legacy"},
		{"", "none"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		t.Run("priority_display_type_"+tc.priority, func(t *testing.T) {
			got := GetPriorityDisplayType(tc.priority)
			if got != tc.expected {
				t.Errorf("Expected display type %s for priority %s, got %s", tc.expected, tc.priority, got)
			}
		})
	}
}

func TestPriorityUnicodeSymbolConstants(t *testing.T) {
	symbols := []struct {
		name   string
		symbol string
		code   string
	}{
		{"PriorityHighSymbol", PriorityHighSymbol, "★"},
		{"PriorityMediumSymbol", PriorityMediumSymbol, "☆"},
		{"PriorityLowSymbol", PriorityLowSymbol, "◦"},
		{"PriorityNoneSymbol", PriorityNoneSymbol, "○"},
	}

	for _, s := range symbols {
		t.Run("priority_symbol_"+s.name, func(t *testing.T) {
			if s.symbol != s.code {
				t.Errorf("Expected %s to be %s, got %s", s.name, s.code, s.symbol)
			}
		})
	}
}

func TestPriorityPatternConstants(t *testing.T) {
	patterns := []struct {
		name    string
		pattern string
		code    string
	}{
		{"PriorityHighPattern", PriorityHighPattern, "★★★"},
		{"PriorityMediumPattern", PriorityMediumPattern, "★★☆"},
		{"PriorityLowPattern", PriorityLowPattern, "★☆☆"},
		{"PriorityNonePattern", PriorityNonePattern, "☆☆☆"},
	}

	for _, p := range patterns {
		t.Run("priority_pattern_"+p.name, func(t *testing.T) {
			if p.pattern != p.code {
				t.Errorf("Expected %s to be %s, got %s", p.name, p.code, p.pattern)
			}
		})
	}
}

func TestPriorityStyleConstants(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"PriorityHighStyle", PriorityHighStyle},
		{"PriorityMediumStyle", PriorityMediumStyle},
		{"PriorityLowStyle", PriorityLowStyle},
		{"PriorityNoneStyle", PriorityNoneStyle},
		{"PriorityLegacyStyle", PriorityLegacyStyle},
	}

	for _, s := range styles {
		t.Run("priority_style_"+s.name, func(t *testing.T) {
			testText := "test"
			rendered := s.style.Render(testText)
			if rendered == "" {
				t.Errorf("Expected %s to produce output when rendering text", s.name)
			}
		})
	}
}
