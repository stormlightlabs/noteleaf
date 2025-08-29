package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

const (
	// U+25CF Black Circle
	StatusTodoSymbol = "●"
	// U+25D0 Circle with Left Half Black
	StatusInProgressSymbol = "◐"
	// U+25A0 Black Square
	StatusBlockedSymbol = "■"
	// U+2713 Check Mark
	StatusDoneSymbol = "✓"
	// U+26AB Medium Black Circle
	StatusAbandonedSymbol = "⚫"
	// U+25CB White Circle
	StatusPendingSymbol = "○"
	// U+2713 Check Mark
	StatusCompletedSymbol = "✓"
	// U+2717 Ballot X
	StatusDeletedSymbol = "✗"
	// U+2605 Black Star
	PriorityHighSymbol = "★"
	// U+2606 White Star
	PriorityMediumSymbol = "☆"
	// U+25E6 White Bullet
	PriorityLowSymbol = "◦"
	// U+25CB White Circle
	PriorityNoneSymbol = "○"
	// Three stars
	PriorityHighPattern = "★★★"
	// Two stars, one outline
	PriorityMediumPattern = "★★☆"
	// One star, two outline
	PriorityLowPattern = "★☆☆"
	// Three outline stars
	PriorityNonePattern = "☆☆☆"
)

var (
	// Gray
	TodoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	// Blue
	InProgressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	// Red
	BlockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	// Green
	DoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	// Dark Gray
	AbandonedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	// Light Gray
	PendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	// Green
	CompletedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	// Dark Red
	DeletedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	// Bright Red - highest urgency
	PriorityHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	// Yellow - medium urgency
	PriorityMediumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	// Cyan - low urgency
	PriorityLowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	// Gray - no priority
	PriorityNoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	// For legacy A-Z and numeric priorities
	PriorityLegacyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")) // Magenta
)

// GetStatusSymbol returns the unicode symbol for a given status
//
//	Default to pending
func GetStatusSymbol(status string) string {
	switch status {
	case models.StatusTodo:
		return StatusTodoSymbol
	case models.StatusInProgress:
		return StatusInProgressSymbol
	case models.StatusBlocked:
		return StatusBlockedSymbol
	case models.StatusDone:
		return StatusDoneSymbol
	case models.StatusAbandoned:
		return StatusAbandonedSymbol
	case models.StatusPending:
		return StatusPendingSymbol
	case models.StatusCompleted:
		return StatusCompletedSymbol
	case models.StatusDeleted:
		return StatusDeletedSymbol
	default:
		return StatusPendingSymbol
	}
}

// GetStatusStyle returns the color style for a given status
//
//	Defaults to pending style
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case models.StatusTodo:
		return TodoStyle
	case models.StatusInProgress:
		return InProgressStyle
	case models.StatusBlocked:
		return BlockedStyle
	case models.StatusDone:
		return DoneStyle
	case models.StatusAbandoned:
		return AbandonedStyle
	case models.StatusPending:
		return PendingStyle
	case models.StatusCompleted:
		return CompletedStyle
	case models.StatusDeleted:
		return DeletedStyle
	default:
		return PendingStyle
	}
}

// FormatStatusIndicator returns a styled status symbol and text
func FormatStatusIndicator(status string) string {
	symbol := GetStatusSymbol(status)
	style := GetStatusStyle(status)
	return style.Render(symbol)
}

// FormatStatusWithText returns a styled status symbol with status text
func FormatStatusWithText(status string) string {
	symbol := GetStatusSymbol(status)
	style := GetStatusStyle(status)
	return style.Render(fmt.Sprintf("%s %s", symbol, status))
}

// GetStatusDescription returns a human-friendly description of the status
func GetStatusDescription(status string) string {
	switch status {
	case models.StatusTodo:
		return "Ready to start"
	case models.StatusInProgress:
		return "Currently working"
	case models.StatusBlocked:
		return "Waiting on dependency"
	case models.StatusDone:
		return "Completed successfully"
	case models.StatusAbandoned:
		return "No longer relevant"
	case models.StatusPending:
		return "Pending (legacy)"
	case models.StatusCompleted:
		return "Completed (legacy)"
	case models.StatusDeleted:
		return "Deleted (legacy)"
	default:
		return "Unknown status"
	}
}

// FormatTaskStatus returns a complete status display with symbol, status, and description
func FormatTaskStatus(task *models.Task) string {
	if task == nil {
		return ""
	}

	symbol := GetStatusSymbol(task.Status)
	style := GetStatusStyle(task.Status)
	description := GetStatusDescription(task.Status)

	return fmt.Sprintf("%s %s - %s", style.Render(symbol), style.Render(task.Status), description)
}

// StatusLegend returns a formatted legend showing all status symbols
func StatusLegend() string {
	var parts []string

	statuses := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
		models.StatusAbandoned,
	}

	for _, status := range statuses {
		parts = append(parts, FormatStatusWithText(status))
	}

	return strings.Join(parts, "  ")
}

// GetAllStatusSymbols returns a map of all status symbols for reference
func GetAllStatusSymbols() map[string]string {
	return map[string]string{
		models.StatusTodo:       StatusTodoSymbol,
		models.StatusInProgress: StatusInProgressSymbol,
		models.StatusBlocked:    StatusBlockedSymbol,
		models.StatusDone:       StatusDoneSymbol,
		models.StatusAbandoned:  StatusAbandonedSymbol,
		models.StatusPending:    StatusPendingSymbol,
		models.StatusCompleted:  StatusCompletedSymbol,
		models.StatusDeleted:    StatusDeletedSymbol,
	}
}

// IsValidStatusTransition checks if a status transition is logically valid
//
//	From todo, can go to in-progress, blocked, done, or abandoned
//	From in-progress, can go to blocked, done, abandoned, or back to todo
//	From blocked, can go to todo, in-progress, done, or abandoned
//	From done, can only be reopened to todo or in-progress
//	From abandoned, can be reopened to todo or in-progress
func IsValidStatusTransition(from, to string) bool {
	if from == models.StatusTodo {
		validTo := []string{models.StatusInProgress, models.StatusBlocked, models.StatusDone, models.StatusAbandoned}
		return slices.Contains(validTo, to)
	}

	if from == models.StatusInProgress {
		validTo := []string{models.StatusTodo, models.StatusBlocked, models.StatusDone, models.StatusAbandoned}
		return slices.Contains(validTo, to)
	}

	if from == models.StatusBlocked {
		validTo := []string{models.StatusTodo, models.StatusInProgress, models.StatusDone, models.StatusAbandoned}
		return slices.Contains(validTo, to)
	}

	if from == models.StatusDone {
		validTo := []string{models.StatusTodo, models.StatusInProgress}
		return slices.Contains(validTo, to)
	}

	if from == models.StatusAbandoned {
		validTo := []string{models.StatusTodo, models.StatusInProgress}
		return slices.Contains(validTo, to)
	}

	if from == models.StatusPending {
		return to == models.StatusTodo || to == models.StatusInProgress
	}

	if from == models.StatusCompleted {
		return to == models.StatusDone
	}

	return false
}

// GetPrioritySymbol returns the unicode symbol for a given priority
func GetPrioritySymbol(priority string) string {
	switch priority {
	case models.PriorityHigh:
		return PriorityHighSymbol
	case models.PriorityMedium:
		return PriorityMediumSymbol
	case models.PriorityLow:
		return PriorityLowSymbol
	case "":
		return PriorityNoneSymbol
	default:
		if len(priority) == 1 && priority >= "A" && priority <= "Z" {
			return PriorityHighSymbol
		}
		switch priority {
		case "1":
			return PriorityLowSymbol
		case "2", "3":
			return PriorityMediumSymbol
		case "4", "5":
			return PriorityHighSymbol
		default:
			return PriorityNoneSymbol
		}
	}
}

// GetPriorityPattern returns the star pattern for a given priority
func GetPriorityPattern(priority string) string {
	switch priority {
	case models.PriorityHigh:
		return PriorityHighPattern
	case models.PriorityMedium:
		return PriorityMediumPattern
	case models.PriorityLow:
		return PriorityLowPattern
	case "":
		return PriorityNonePattern
	default:
		if len(priority) == 1 && priority >= "A" && priority <= "Z" {
			if priority <= "C" {
				return PriorityHighPattern
			} else if priority <= "M" {
				return PriorityMediumPattern
			} else {
				return PriorityLowPattern
			}
		}
		switch priority {
		case "1":
			return PriorityLowPattern
		case "2", "3":
			return PriorityMediumPattern
		case "4", "5":
			return PriorityHighPattern
		default:
			return PriorityNonePattern
		}
	}
}

// GetPriorityStyle returns the color style for a given priority
func GetPriorityStyle(priority string) lipgloss.Style {
	switch priority {
	case models.PriorityHigh:
		return PriorityHighStyle
	case models.PriorityMedium:
		return PriorityMediumStyle
	case models.PriorityLow:
		return PriorityLowStyle
	case "":
		return PriorityNoneStyle
	default:
		if len(priority) == 1 && priority >= "A" && priority <= "Z" {
			return PriorityLegacyStyle
		}
		switch priority {
		case "1", "2", "3", "4", "5":
			return PriorityLegacyStyle
		default:
			return PriorityNoneStyle
		}
	}
}

// FormatPriorityIndicator returns a styled priority pattern
func FormatPriorityIndicator(priority string) string {
	pattern := GetPriorityPattern(priority)
	style := GetPriorityStyle(priority)
	return style.Render(pattern)
}

// FormatPriorityWithText returns a styled priority with text description
func FormatPriorityWithText(priority string) string {
	pattern := GetPriorityPattern(priority)
	style := GetPriorityStyle(priority)

	if priority == "" {
		return style.Render(fmt.Sprintf("%s None", pattern))
	}

	return style.Render(fmt.Sprintf("%s %s", pattern, priority))
}

// GetPriorityDescription returns a human-friendly description of the priority
func GetPriorityDescription(priority string) string {
	switch priority {
	case models.PriorityHigh:
		return "Urgent - do first"
	case models.PriorityMedium:
		return "Important - schedule soon"
	case models.PriorityLow:
		return "Nice to have - when time permits"
	case "":
		return "No priority set"
	default:
		if len(priority) == 1 && priority >= "A" && priority <= "Z" {
			return fmt.Sprintf("Priority %s (legacy)", priority)
		}
		switch priority {
		case "1":
			return "Priority 1 (lowest)"
		case "2":
			return "Priority 2 (low)"
		case "3":
			return "Priority 3 (medium)"
		case "4":
			return "Priority 4 (high)"
		case "5":
			return "Priority 5 (highest)"
		default:
			return "Unknown priority"
		}
	}
}

// FormatTaskPriority returns a complete priority display with pattern, priority, and description
func FormatTaskPriority(task *models.Task) string {
	if task == nil {
		return ""
	}

	pattern := GetPriorityPattern(task.Priority)
	style := GetPriorityStyle(task.Priority)
	description := GetPriorityDescription(task.Priority)

	if task.Priority == "" {
		return fmt.Sprintf("%s %s", style.Render(pattern), description)
	}

	return fmt.Sprintf("%s %s - %s", style.Render(pattern), style.Render(task.Priority), description)
}

// PriorityLegend returns a formatted legend showing all priority patterns
func PriorityLegend() string {
	var parts []string

	priorities := []string{
		models.PriorityHigh, models.PriorityMedium, models.PriorityLow, "",
	}

	for _, priority := range priorities {
		parts = append(parts, FormatPriorityWithText(priority))
	}

	return strings.Join(parts, "  ")
}

// GetAllPrioritySymbols returns a map of all priority symbols for reference
func GetAllPrioritySymbols() map[string]string {
	return map[string]string{
		models.PriorityHigh:   PriorityHighSymbol,
		models.PriorityMedium: PriorityMediumSymbol,
		models.PriorityLow:    PriorityLowSymbol,
		"":                    PriorityNoneSymbol,
	}
}

// GetAllPriorityPatterns returns a map of all priority patterns for reference
func GetAllPriorityPatterns() map[string]string {
	return map[string]string{
		models.PriorityHigh:   PriorityHighPattern,
		models.PriorityMedium: PriorityMediumPattern,
		models.PriorityLow:    PriorityLowPattern,
		"":                    PriorityNonePattern,
	}
}

// GetPriorityDisplayType returns the display type for a priority (text, numeric, or legacy)
func GetPriorityDisplayType(priority string) string {
	switch priority {
	case models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
		return "text"
	case "1", "2", "3", "4", "5":
		return "numeric"
	case "":
		return "none"
	default:
		if len(priority) == 1 && priority >= "A" && priority <= "Z" {
			return "legacy"
		}
		return "unknown"
	}
}
