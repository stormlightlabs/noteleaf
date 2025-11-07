package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Style constructor helpers
func newStyle() lipgloss.Style              { return lipgloss.NewStyle() }
func newPStyle(v, h int) lipgloss.Style     { return lipgloss.NewStyle().Padding(v, h) }
func newBoldStyle() lipgloss.Style          { return newStyle().Bold(true) }
func newPBoldStyle(v, h int) lipgloss.Style { return newPStyle(v, h).Bold(true) }
func newEmStyle() lipgloss.Style            { return newStyle().Italic(true) }

// Rendering helpers (private, used by public API)
func success(msg string) string      { return SuccessStyle.Render("✓ " + msg) }
func errorMsg(msg string) string     { return ErrorStyle.Render("✗ " + msg) }
func warning(msg string) string      { return WarningStyle.Render("⚠ " + msg) }
func info(msg string) string         { return InfoStyle.Render("ℹ " + msg) }
func infop(msg string) string        { return InfoStyle.Render(msg) }
func title(msg string) string        { return TitleStyle.Render(msg) }
func subtitle(msg string) string     { return SubtitleStyle.Render(msg) }
func box(content string) string      { return BoxStyle.Render(content) }
func errorBox(content string) string { return ErrorBoxStyle.Render(content) }
func text(content string) string     { return TextStyle.Render(content) }
func muted(content string) string    { return MutedStyle.Render(content) }
func accent(content string) string   { return AccentStyle.Render(content) }
func header(content string) string   { return HeaderStyle.Render(content) }
func primary(content string) string  { return PrimaryStyle.Render(content) }

// Success prints a formatted success message
func Success(format string, a ...any) {
	fmt.Print(success(fmt.Sprintf(format, a...)))
}

// Successln prints a formatted success message with a newline
func Successln(format string, a ...any) {
	fmt.Println(success(fmt.Sprintf(format, a...)))
}

// Error prints a formatted error message
func Error(format string, a ...any) {
	fmt.Print(errorMsg(fmt.Sprintf(format, a...)))
}

// Errorln prints a formatted error message with a newline
func Errorln(format string, a ...any) {
	fmt.Println(errorMsg(fmt.Sprintf(format, a...)))
}

// Warning prints a formatted warning message
func Warning(format string, a ...any) {
	fmt.Print(warning(fmt.Sprintf(format, a...)))
}

// Warningln prints a formatted warning message with a newline
func Warningln(format string, a ...any) {
	fmt.Println(warning(fmt.Sprintf(format, a...)))
}

// Info prints a formatted info message
func Info(format string, a ...any) {
	fmt.Print(info(fmt.Sprintf(format, a...)))
}

// Infoln prints a formatted info message with a newline
func Infoln(format string, a ...any) {
	fmt.Println(infop(fmt.Sprintf(format, a...)))
}

// Infop prints a formatted info message, sans icon
func Infop(format string, a ...any) {
	fmt.Print(infop(fmt.Sprintf(format, a...)))
}

// Infopln prints a formatted info message with a newline, sans icon
func Infopln(format string, a ...any) {
	fmt.Println(info(fmt.Sprintf(format, a...)))
}

// Title prints a formatted title
func Title(format string, a ...any) {
	fmt.Print(title(fmt.Sprintf(format, a...)))
}

// Titleln prints a formatted title with a newline
func Titleln(format string, a ...any) {
	fmt.Println(title(fmt.Sprintf(format, a...)))
}

// Subtitle prints a formatted subtitle
func Subtitle(format string, a ...any) {
	fmt.Print(subtitle(fmt.Sprintf(format, a...)))
}

// Subtitleln prints a formatted subtitle with a newline
func Subtitleln(format string, a ...any) {
	fmt.Println(subtitle(fmt.Sprintf(format, a...)))
}

// Box prints content in a styled box
func Box(format string, a ...any) {
	fmt.Print(box(fmt.Sprintf(format, a...)))
}

// Boxln prints content in a styled box with a newline
func Boxln(format string, a ...any) {
	fmt.Println(box(fmt.Sprintf(format, a...)))
}

// ErrorBox prints error content in a styled error box
func ErrorBox(format string, a ...any) {
	fmt.Print(errorBox(fmt.Sprintf(format, a...)))
}

// ErrorBoxln prints error content in a styled error box with a newline
func ErrorBoxln(format string, a ...any) {
	fmt.Println(errorBox(fmt.Sprintf(format, a...)))
}

func Newline()                         { fmt.Println() }
func Plain(format string, a ...any)    { fmt.Print(text(fmt.Sprintf(format, a...))) }
func Plainln(format string, a ...any)  { fmt.Println(text(fmt.Sprintf(format, a...))) }
func Header(format string, a ...any)   { fmt.Print(header(fmt.Sprintf(format, a...))) }
func Headerln(format string, a ...any) { fmt.Println(header(fmt.Sprintf(format, a...))) }

// Muted prints muted/secondary text
func Muted(format string, a ...any) {
	fmt.Print(muted(fmt.Sprintf(format, a...)))
}

// Mutedln prints muted/secondary text with a newline
func Mutedln(format string, a ...any) {
	fmt.Println(muted(fmt.Sprintf(format, a...)))
}

// Accent prints accent-colored text
func Accent(format string, a ...any) {
	fmt.Print(accent(fmt.Sprintf(format, a...)))
}

// Accentln prints accent-colored text with a newline
func Accentln(format string, a ...any) {
	fmt.Println(accent(fmt.Sprintf(format, a...)))
}

// Primary prints primary-colored text
func Primary(format string, a ...any) {
	fmt.Print(primary(fmt.Sprintf(format, a...)))
}

// Primaryln prints primary-colored text with a newline
func Primaryln(format string, a ...any) {
	fmt.Println(primary(fmt.Sprintf(format, a...)))
}
