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
// Background colors (dark mode, Iceberg-inspired)
	ColorBGBase      = Pepper.Hex()   // #201F26 - Darkest base
	ColorBGSecondary = BBQ.Hex()      // #2d2c35 - Secondary background
	ColorBGTertiary  = Charcoal.Hex() // #3A3943 - Tertiary/elevated
	ColorBGInput     = Iron.Hex()     // #4D4C57 - Input fields/focus

	// Text colors (light to dark hierarchy)
	ColorTextPrimary   = Salt.Hex()   // #F1EFEF - Primary text (brightest)
	ColorTextSecondary = Smoke.Hex()  // #BFBCC8 - Secondary text
	ColorTextMuted     = Squid.Hex()  // #858392 - Muted/comments
	ColorTextDimmed    = Oyster.Hex() // #605F6B - Dimmed text

	// Semantic colors (Iceberg-inspired: cool blues/purples with warm accents)
	ColorPrimary = Malibu.Hex()   // #00A4FF - Blue (primary accent)
	ColorSuccess = Julep.Hex()    // #00FFB2 - Green (success/positive)
	ColorError   = Sriracha.Hex() // #EB4268 - Red (errors)
	ColorWarning = Tang.Hex()     // #FF985A - Orange (warnings)
	ColorInfo    = Violet.Hex()   // #C259FF - Purple (info)
	ColorAccent  = Lichen.Hex()   // #5CDFEA - Teal (secondary accent)

	// Base styles
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

	// Layout styles
	BoxStyle      = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorPrimary))
	ErrorBoxStyle = newPStyle(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(ColorError))
	HeaderStyle   = newPBoldStyle(0, 1).Foreground(lipgloss.Color(ColorPrimary))
	CellStyle     = newPStyle(0, 1).Foreground(lipgloss.Color(ColorTextPrimary))

	// List styles
	ListItemStyle     = newStyle().Foreground(lipgloss.Color(ColorTextPrimary)).PaddingLeft(2)
	SelectedItemStyle = newBoldStyle().Foreground(lipgloss.Color(ColorPrimary)).PaddingLeft(2)

	// Table/data view styles (replacing ANSI code references)
	TableStyle         = newStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(ColorTextMuted))
	TableHeaderStyle   = newBoldStyle().Foreground(lipgloss.Color(ColorAccent))
	TableTitleStyle    = newBoldStyle().Foreground(lipgloss.Color(ColorPrimary))
	TableSelectedStyle = newBoldStyle().Foreground(lipgloss.Color(ColorTextPrimary)).Background(lipgloss.Color(ColorBGInput))

	// Task-specific styles
	TaskTitleStyle = newBoldStyle().Foreground(lipgloss.Color(ColorTextPrimary))
	TaskIDStyle    = newStyle().Foreground(lipgloss.Color(ColorTextMuted)).Width(8)

	// Status styles (Iceberg-inspired: muted → blue → red → green)
	StatusTodo       = newStyle().Foreground(lipgloss.Color(ColorTextMuted))     // Gray (muted)
	StatusInProgress = newStyle().Foreground(lipgloss.Color(ColorPrimary))       // Blue (active)
	StatusBlocked    = newStyle().Foreground(lipgloss.Color(ColorError))         // Red (blocked)
	StatusDone       = newStyle().Foreground(lipgloss.Color(ColorSuccess))       // Green (success)
	StatusPending    = newStyle().Foreground(lipgloss.Color(ColorWarning))       // Orange (pending)
	StatusCompleted  = newStyle().Foreground(lipgloss.Color(ColorSuccess))       // Green (completed)
	StatusAbandoned  = newStyle().Foreground(lipgloss.Color(ColorTextDimmed))    // Dimmed gray (abandoned)
	StatusDeleted    = newStyle().Foreground(lipgloss.Color(Cherry.Hex()))       // Dark red (deleted)

	// Priority styles (Iceberg-inspired: red → orange → gray)
	PriorityHigh    = newBoldStyle().Foreground(lipgloss.Color(Cherry.Hex()))   // #FF388B - Bright red
	PriorityMedium  = newStyle().Foreground(lipgloss.Color(Tang.Hex()))         // #FF985A - Orange
	PriorityLow     = newStyle().Foreground(lipgloss.Color(ColorAccent))        // Teal (low)
	PriorityNone    = newStyle().Foreground(lipgloss.Color(ColorTextMuted))     // Gray (no priority)
	PriorityLegacy  = newStyle().Foreground(lipgloss.Color(Urchin.Hex()))       // #C337E0 - Magenta (legacy)

	// Content type styles (distinctive colors for different media)
	MovieStyle = newBoldStyle().Foreground(lipgloss.Color(Coral.Hex()))   // #FF577D - Pink/coral
	TVStyle    = newBoldStyle().Foreground(lipgloss.Color(Violet.Hex()))  // #C259FF - Purple
	BookStyle  = newBoldStyle().Foreground(lipgloss.Color(Guac.Hex()))    // #12C78F - Green
	MusicStyle = newBoldStyle().Foreground(lipgloss.Color(Lichen.Hex()))  // #5CDFEA - Teal

	// Diff styles
	AdditionStyle = newStyle().Foreground(lipgloss.Color(Pickle.Hex()))  // #00A475 - Green
	DeletionStyle = newStyle().Foreground(lipgloss.Color(Pom.Hex()))     // #AB2454 - Red
)
