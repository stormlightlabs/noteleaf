package ui

import (
	"image/color"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	lipglossv2 "github.com/charmbracelet/lipgloss/v2"
)

var (
	PrimaryColors     = []Key{Guac, Julep, Bok, Pickle, NeueGuac}
	SecondaryColors   = []Key{Malibu, Sardine, Lichen}
	TertiaryColors    = []Key{Violet, Mauve, Plum, Orchid, Charple, Hazy}
	ProvisionalColors = []Key{NeueGuac, NeueZinc}
	AdditionColors    = []Key{Pickle, Gator, Spinach}
	DeletionColors    = []Key{Pom, Steak, Toast}
)

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
	ColorPrimary = Thunder.Hex() // Blue
	ColorAccent  = Cumin.Hex()   // Yellow/Gold
	ColorError   = Paprika.Hex() // Red/Pink
	ColorText    = Salt.Hex()    // Light text
	ColorBG      = Pepper.Hex()  // Dark background

	PrimaryStyle      = newStyle().Foreground(lipgloss.Color(ColorPrimary))
	AccentStyle       = newStyle().Foreground(lipgloss.Color(ColorAccent))
	ErrorStyle        = newStyle().Foreground(lipgloss.Color(ColorError))
	TextStyle         = newStyle().Foreground(lipgloss.Color(ColorText))
	TitleStyle        = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorAccent))
	SubtitleStyle     = newEmStyle().Foreground(lipgloss.Color(ColorPrimary))
	SuccessStyle      = newBoldStyle().Foreground(lipgloss.Color(ColorPrimary))
	WarningStyle      = newBoldStyle().Foreground(lipgloss.Color(ColorAccent))
	InfoStyle         = newStyle().Foreground(lipgloss.Color(ColorText))
	BoxStyle          = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPrimary))
	ErrorBoxStyle     = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorError))
	ListItemStyle     = newStyle().Foreground(lipgloss.Color(ColorText)).PaddingLeft(2)
	SelectedItemStyle = newBoldStyle().Foreground(lipgloss.Color(ColorAccent)).PaddingLeft(2)
	HeaderStyle       = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorPrimary))
	CellStyle         = newPStyle(0, 1).Foreground(lipgloss.Color(ColorText))

	TaskTitleStyle = newBoldStyle().Foreground(lipgloss.Color(Salt.Hex()))
	TaskIDStyle    = newStyle().Foreground(lipgloss.Color(Squid.Hex())).Width(8)

	StatusPending   = newStyle().Foreground(lipgloss.Color(Citron.Hex()))
	StatusCompleted = newStyle().Foreground(lipgloss.Color(Julep.Hex()))

	PriorityHigh   = newBoldStyle().Foreground(lipgloss.Color(Cherry.Hex()))
	PriorityMedium = newStyle().Foreground(lipgloss.Color(Citron.Hex()))
	PriorityLow    = newStyle().Foreground(lipgloss.Color(Squid.Hex()))

	MovieStyle = newBoldStyle().Foreground(lipgloss.Color(Coral.Hex()))
	TVStyle    = newBoldStyle().Foreground(lipgloss.Color(Violet.Hex()))
	BookStyle  = newBoldStyle().Foreground(lipgloss.Color(Guac.Hex()))
	MusicStyle = newBoldStyle().Foreground(lipgloss.Color(Lichen.Hex()))

	TableStyle         = newStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(Smoke.Hex()))
	SelectedStyle      = newBoldStyle().Foreground(lipgloss.Color(Salt.Hex())).Background(lipgloss.Color(Squid.Hex()))
	TitleColorStyle    = newBoldStyle().Foreground(lipgloss.Color("212"))
	SelectedColorStyle = newBoldStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("212"))
	HeaderColorStyle   = newBoldStyle().Foreground(lipgloss.Color("240"))
)
