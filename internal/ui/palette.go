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

// noteleafColorScheme provides Iceberg-inspired colors for CLI help/documentation
//
//	Philosophy: Cool blues as primary, warm accents for emphasis, hierarchical text colors
//	See: https://github.com/cocopon/iceberg.vim for more information
func noteleafColorScheme(c lipglossv2.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           c(Salt, Pepper),                    // Primary text on dark background
		Title:          c(Malibu, Malibu),                  // Blue primary for titles (Iceberg primary)
		Description:    c(Smoke, Smoke),                    // Secondary text for descriptions
		Codeblock:      c(Butter, BBQ),                     // Light/Dark background for code blocks
		Program:        c(Malibu, Sardine),                 // Blue for program names (primary accent)
		DimmedArgument: c(Oyster, Ash),                     // Dimmed text for optional arguments
		Comment:        c(Squid, Squid),                    // Muted gray for comments (Iceberg comment)
		Flag:           c(Hazy, Jelly),                     // Purple for flags (Iceberg special)
		FlagDefault:    c(Lichen, Turtle),                  // Teal for flag defaults (secondary accent)
		Command:        c(Julep, Julep),                    // Green for commands (success/positive)
		QuotedString:   c(Tang, Tang),                      // Orange for quoted strings (warning/warm)
		Argument:       c(Lichen, Lichen),                  // Teal for arguments (secondary accent)
		Help:           c(Squid, Squid),                    // Muted gray for help text
		Dash:           c(Oyster, Oyster),                  // Dimmed gray for dashes/separators
		ErrorHeader:    [2]color.Color{Sriracha, Sriracha}, // Red for error headers (Iceberg error)
		ErrorDetails:   c(Coral, Salmon),                   // Pink/coral for error details
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
	ColorBGBase      = Pepper.Hex()   // #201F26 - Darkest base
	ColorBGSecondary = BBQ.Hex()      // #2d2c35 - Secondary background
	ColorBGTertiary  = Charcoal.Hex() // #3A3943 - Tertiary/elevated
	ColorBGInput     = Iron.Hex()     // #4D4C57 - Input fields/focus

	ColorTextPrimary   = Salt.Hex()   // #F1EFEF - Primary text (brightest)
	ColorTextSecondary = Smoke.Hex()  // #BFBCC8 - Secondary text
	ColorTextMuted     = Squid.Hex()  // #858392 - Muted/comments
	ColorTextDimmed    = Oyster.Hex() // #605F6B - Dimmed text

	ColorPrimary = Malibu.Hex()   // #00A4FF - Blue (primary accent)
	ColorSuccess = Julep.Hex()    // #00FFB2 - Green (success/positive)
	ColorError   = Sriracha.Hex() // #EB4268 - Red (errors)
	ColorWarning = Tang.Hex()     // #FF985A - Orange (warnings)
	ColorInfo    = Violet.Hex()   // #C259FF - Purple (info)
	ColorAccent  = Lichen.Hex()   // #5CDFEA - Teal (secondary accent)

	PrimaryStyle  = newStyle().Foreground(lipgloss.Color(ColorPrimary))
	SuccessStyle  = newBoldStyle().Foreground(lipgloss.Color(ColorSuccess))
	ErrorStyle    = newBoldStyle().Foreground(lipgloss.Color(ColorError))
	WarningStyle  = newBoldStyle().Foreground(lipgloss.Color(ColorWarning))
	InfoStyle     = newStyle().Foreground(lipgloss.Color(ColorTextSecondary))
	AccentStyle   = newStyle().Foreground(lipgloss.Color(ColorAccent))
	TextStyle     = newStyle().Foreground(lipgloss.Color(ColorTextPrimary))
	MutedStyle    = newStyle().Foreground(lipgloss.Color(ColorTextMuted))
	TitleStyle    = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorPrimary))
	SubtitleStyle = newEmStyle().Foreground(lipgloss.Color(ColorAccent))

	BoxStyle      = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPrimary))
	ErrorBoxStyle = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorError))
	HeaderStyle   = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorPrimary))
	CellStyle     = newPStyle(0, 1).Foreground(lipgloss.Color(ColorTextPrimary))

	ListItemStyle     = newStyle().Foreground(lipgloss.Color(ColorTextPrimary)).PaddingLeft(2)
	SelectedItemStyle = newBoldStyle().Foreground(lipgloss.Color(ColorPrimary)).PaddingLeft(2)

	TableStyle         = newStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(ColorTextMuted))
	TableHeaderStyle   = newBoldStyle().Foreground(lipgloss.Color(ColorAccent))
	TableTitleStyle    = newBoldStyle().Foreground(lipgloss.Color(ColorPrimary))
	TableSelectedStyle = newBoldStyle().Foreground(lipgloss.Color(ColorTextPrimary)).Background(lipgloss.Color(ColorBGInput))

	TaskTitleStyle = newBoldStyle().Foreground(lipgloss.Color(ColorTextPrimary))
	TaskIDStyle    = newStyle().Foreground(lipgloss.Color(ColorTextMuted)).Width(8)

	StatusTodo       = newStyle().Foreground(lipgloss.Color(ColorTextMuted))  // Gray (muted)
	StatusInProgress = newStyle().Foreground(lipgloss.Color(ColorPrimary))    // Blue (active)
	StatusBlocked    = newStyle().Foreground(lipgloss.Color(ColorError))      // Red (blocked)
	StatusDone       = newStyle().Foreground(lipgloss.Color(ColorSuccess))    // Green (success)
	StatusPending    = newStyle().Foreground(lipgloss.Color(ColorWarning))    // Orange (pending)
	StatusCompleted  = newStyle().Foreground(lipgloss.Color(ColorSuccess))    // Green (completed)
	StatusAbandoned  = newStyle().Foreground(lipgloss.Color(ColorTextDimmed)) // Dimmed gray (abandoned)
	StatusDeleted    = newStyle().Foreground(lipgloss.Color(Pom.Hex()))       // Dark red (deleted)

	PriorityHigh   = newBoldStyle().Foreground(lipgloss.Color(Pom.Hex()))  // #FF388B - Bright red
	PriorityMedium = newStyle().Foreground(lipgloss.Color(Tang.Hex()))     // #FF985A - Orange
	PriorityLow    = newStyle().Foreground(lipgloss.Color(ColorAccent))    // #5CDFEA - Teal (low)
	PriorityNone   = newStyle().Foreground(lipgloss.Color(ColorTextMuted)) // #858392 - Gray (no priority)
	PriorityLegacy = newStyle().Foreground(lipgloss.Color(Urchin.Hex()))   // #C337E0 - Magenta (legacy)

	MovieStyle = newBoldStyle().Foreground(lipgloss.Color(Coral.Hex()))  // #FF577D - Pink/coral
	TVStyle    = newBoldStyle().Foreground(lipgloss.Color(Violet.Hex())) // #C259FF - Purple
	BookStyle  = newBoldStyle().Foreground(lipgloss.Color(Guac.Hex()))   // #12C78F - Green
	MusicStyle = newBoldStyle().Foreground(lipgloss.Color(Lichen.Hex())) // #5CDFEA - Teal

	AdditionStyle = newStyle().Foreground(lipgloss.Color(Pickle.Hex())) // #00A475 - Green
	DeletionStyle = newStyle().Foreground(lipgloss.Color(Pom.Hex()))    // #AB2454 - Red
)
