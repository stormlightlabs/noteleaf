package ui

import (
	"image/color"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	lipglossv2 "github.com/charmbracelet/lipgloss/v2"
)

var PrimaryColors []Key = []Key{
	Guac,
	Julep,
	Bok,
	Pickle,
	NeueGuac,
}

var SecondaryColors []Key = []Key{
	Malibu,
	Sardine,
	Lichen,
}

var TertiaryColors []Key = []Key{
	Violet,
	Mauve,
	Plum,
	Orchid,
	Charple,
	Hazy,
}

var ProvisionalColors []Key = []Key{NeueGuac, NeueZinc}

var AdditionColors []Key = []Key{Pickle, Gator, Spinach}

var DeletionColors []Key = []Key{Pom, Steak, Toast}

var NoteleafColorScheme fang.ColorSchemeFunc = noteleafColorScheme

func noteleafColorScheme(c lipglossv2.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           c(Salt, Pepper),                  // Light/Dark base text
		Title:          c(Guac, Julep),                   // Green primary for titles
		Description:    c(Squid, Smoke),                  // Muted gray for descriptions
		Codeblock:      c(Butter, BBQ),                   // Light/Dark background for code
		Program:        c(Malibu, Sardine),               // Blue for program names
		DimmedArgument: c(Oyster, Ash),                   // Subtle gray for dimmed text
		Comment:        c(Pickle, NeueGuac),              // Green for comments
		Flag:           c(Violet, Mauve),                 // Purple for flags
		FlagDefault:    c(Lichen, Turtle),                // Teal for flag defaults
		Command:        c(Julep, Guac),                   // Bright green for commands
		QuotedString:   c(Citron, Mustard),               // Yellow for quoted strings
		Argument:       c(Sapphire, Guppy),               // Blue for arguments
		Help:           c(Smoke, Iron),                   // Gray for help text
		Dash:           c(Iron, Oyster),                  // Medium gray for dashes
		ErrorHeader:    [2]color.Color{Cherry, Sriracha}, // Red for error headers (fg, bg)
		ErrorDetails:   c(Coral, Salmon),                 // Red/pink for error details
	}
}

// Palette provides semantic color access
type Palette struct {
	scheme fang.ColorScheme
}

// NewPalette creates a new palette instance
func NewPalette(isDark bool) *Palette {
	return &Palette{
		scheme: noteleafColorScheme(lipglossv2.LightDark(isDark)),
	}
}

var (
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Julep.Hex())).
		Bold(true)

	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Cherry.Hex())).
		Bold(true)

	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Malibu.Hex()))

	Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Citron.Hex())).
		Bold(true)

	Path = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Mustard.Hex())).
		Italic(true)
)

var (
	TaskTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Salt.Hex())).
			Bold(true)

	TaskID = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Squid.Hex())).
		Width(8)
)

var (
	StatusPending = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Citron.Hex()))

	StatusCompleted = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Julep.Hex()))
)

var (
	PriorityHigh = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Cherry.Hex())).
			Bold(true)

	PriorityMedium = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Citron.Hex()))

	PriorityLow = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Squid.Hex()))
)

var (
	MovieStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Coral.Hex())).
			Bold(true)

	TVStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(Violet.Hex())).
		Bold(true)

	BookStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Guac.Hex())).
			Bold(true)

	MusicStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Lichen.Hex())).
			Bold(true)
)

// Table and UI styles
var (
	TableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(Smoke.Hex()))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Salt.Hex())).
			Background(lipgloss.Color(Squid.Hex())).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(Smoke.Hex())).
			BorderBottom(true).
			Bold(false)
)
