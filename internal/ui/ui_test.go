package ui

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func stripAnsi(str string) string {
	return regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`).ReplaceAllString(str, "")
}

func withColor(t *testing.T, fn func(r *lipgloss.Renderer)) {
	t.Helper()

	var buf bytes.Buffer
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.TrueColor)
	lipgloss.SetDefaultRenderer(r)

	fn(r)

	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout))
}

func TestUI(t *testing.T) {
	t.Run("Hex", func(t *testing.T) {
		tests := []struct {
			key      Key
			expected string
		}{
			{Cumin, "#BF976F"},
			{Tang, "#FF985A"},
			{Yam, "#FFB587"},
			{Paprika, "#D36C64"},
			{Bengal, "#FF6E63"},
			{Uni, "#FF937D"},
			{Sriracha, "#EB4268"},
			{Coral, "#FF577D"},
			{Salmon, "#FF7F90"},
			{Chili, "#E23080"},
			{Cherry, "#FF388B"},
			{Tuna, "#FF6DAA"},
			{Macaron, "#E940B0"},
			{Pony, "#FF4FBF"},
			{Cheeky, "#FF79D0"},
			{Flamingo, "#F947E3"},
			{Dolly, "#FF60FF"},
			{Blush, "#FF84FF"},
			{Urchin, "#C337E0"},
			{Mochi, "#EB5DFF"},
			{Lilac, "#F379FF"},
			{Prince, "#9C35E1"},
			{Violet, "#C259FF"},
			{Mauve, "#D46EFF"},
			{Grape, "#7134DD"},
			{Plum, "#9953FF"},
			{Orchid, "#AD6EFF"},
			{Jelly, "#4A30D9"},
			{Charple, "#6B50FF"},
			{Hazy, "#8B75FF"},
			{Ox, "#3331B2"},
			{Sapphire, "#4949FF"},
			{Guppy, "#7272FF"},
			{Oceania, "#2B55B3"},
			{Thunder, "#4776FF"},
			{Anchovy, "#719AFC"},
			{Damson, "#007AB8"},
			{Malibu, "#00A4FF"},
			{Sardine, "#4FBEFE"},
			{Zinc, "#10B1AE"},
			{Turtle, "#0ADCD9"},
			{Lichen, "#5CDFEA"},
			{Guac, "#12C78F"},
			{Julep, "#00FFB2"},
			{Bok, "#68FFD6"},
			{Mustard, "#F5EF34"},
			{Citron, "#E8FF27"},
			{Zest, "#E8FE96"},
			{Pepper, "#201F26"},
			{BBQ, "#2d2c35"},
			{Charcoal, "#3A3943"},
			{Iron, "#4D4C57"},
			{Oyster, "#605F6B"},
			{Squid, "#858392"},
			{Smoke, "#BFBCC8"},
			{Ash, "#DFDBDD"},
			{Salt, "#F1EFEF"},
			{Butter, "#FFFAF1"},
			// Diffs: additions
			{Pickle, "#00A475"},
			{Gator, "#18463D"},
			{Spinach, "#1C3634"},
			{Pom, "#AB2454"},
			{Steak, "#582238"},
			{Toast, "#412130"},
			{NeueGuac, "#00b875"},
			{NeueZinc, "#0e9996"},
			{Key(-1), ""},
		}

		for _, tt := range tests {
			t.Run(tt.key.String(), func(t *testing.T) {
				if hex := tt.key.Hex(); hex != tt.expected {
					t.Errorf("expected hex %q, got %q", tt.expected, hex)
				}
			})
		}
	})

	t.Run("String", func(t *testing.T) {
		tests := []struct {
			key      Key
			expected string
		}{
			{Cumin, "Cumin"},
			{Tang, "Tang"},
			{Yam, "Yam"},
			{Paprika, "Paprika"},
			{Bengal, "Bengal"},
			{Uni, "Uni"},
			{Sriracha, "Sriracha"},
			{Coral, "Coral"},
			{Salmon, "Salmon"},
			{Chili, "Chili"},
			{Cherry, "Cherry"},
			{Tuna, "Tuna"},
			{Macaron, "Macaron"},
			{Pony, "Pony"},
			{Cheeky, "Cheeky"},
			{Flamingo, "Flamingo"},
			{Dolly, "Dolly"},
			{Blush, "Blush"},
			{Urchin, "Urchin"},
			{Mochi, "Mochi"},
			{Lilac, "Lilac"},
			{Prince, "Prince"},
			{Violet, "Violet"},
			{Mauve, "Mauve"},
			{Grape, "Grape"},
			{Plum, "Plum"},
			{Orchid, "Orchid"},
			{Jelly, "Jelly"},
			{Charple, "Charple"},
			{Hazy, "Hazy"},
			{Ox, "Ox"},
			{Sapphire, "Sapphire"},
			{Guppy, "Guppy"},
			{Oceania, "Oceania"},
			{Thunder, "Thunder"},
			{Anchovy, "Anchovy"},
			{Damson, "Damson"},
			{Malibu, "Malibu"},
			{Sardine, "Sardine"},
			{Zinc, "Zinc"},
			{Turtle, "Turtle"},
			{Lichen, "Lichen"},
			{Guac, "Guac"},
			{Julep, "Julep"},
			{Bok, "Bok"},
			{Mustard, "Mustard"},
			{Citron, "Citron"},
			{Zest, "Zest"},
			{Pepper, "Pepper"},
			{BBQ, "BBQ"},
			{Charcoal, "Charcoal"},
			{Iron, "Iron"},
			{Oyster, "Oyster"},
			{Squid, "Squid"},
			{Smoke, "Smoke"},
			{Ash, "Ash"},
			{Salt, "Salt"},
			{Butter, "Butter"},
			{Pickle, "Pickle"},
			{Gator, "Gator"},
			{Spinach, "Spinach"},
			{Pom, "Pom"},
			{Steak, "Steak"},
			{Toast, "Toast"},
			{NeueGuac, "NeueGuac"},
			{NeueZinc, "NeueZinc"},
			{Key(-1), "Key(-1)"},
		}

		for _, tt := range tests {
			t.Run(tt.expected, func(t *testing.T) {
				if str := tt.key.String(); str != tt.expected {
					t.Errorf("expected string %q, got %q", tt.expected, str)
				}
			})
		}
	})

	t.Run("RGBA", func(t *testing.T) {
		t.Run("valid key", func(t *testing.T) {
			r, g, b, a := Cumin.RGBA()
			if a == 0 {
				t.Error("Alpha should not be zero for a valid color")
			}
			if !(r > 0 && g > 0 && b > 0) {
				t.Error("RGB components should be greater than zero for Cumin")
			}
		})

		t.Run("invalid key", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			Key(-1).RGBA()
		})
	})

	t.Run("Color Categories", func(t *testing.T) {
		if !Guac.IsPrimary() {
			t.Error("Guac should be a primary color")
		}
		if Malibu.IsPrimary() {
			t.Error("Malibu should not be a primary color")
		}
		if !Malibu.IsSecondary() {
			t.Error("Malibu should be a secondary color")
		}
		if Guac.IsSecondary() {
			t.Error("Guac should not be a secondary color")
		}
		if !Violet.IsTertiary() {
			t.Error("Violet should be a tertiary color")
		}
		if Guac.IsTertiary() {
			t.Error("Guac should not be a tertiary color")
		}
	})

	t.Run("New Palette", func(t *testing.T) {
		t.Run("dark mode", func(t *testing.T) {
			p := NewPalette(true)
			if p == nil {
				t.Fatal("NewPalette(true) returned nil")
			}
			r, g, b, a := p.scheme.Base.RGBA()
			pr, pg, pb, pa := Pepper.RGBA()
			if r != pr || g != pg || b != pb || a != pa {
				t.Errorf("dark mode base color mismatch: expected Pepper, got %v", p.scheme.Base)
			}
		})

		t.Run("light mode", func(t *testing.T) {
			p := NewPalette(false)
			if p == nil {
				t.Fatal("NewPalette(false) returned nil")
			}
			r, g, b, a := p.scheme.Base.RGBA()
			sr, sg, sb, sa := Salt.RGBA()
			if r != sr || g != sg || b != sb || a != sa {
				t.Errorf("light mode base color mismatch: expected Salt, got %v", p.scheme.Base)
			}
		})
	})

	t.Run("Logo", func(t *testing.T) {
		logos := []struct {
			logo     Logo
			name     string
			contains string
		}{
			{Colossal, "Collosal", "888b    888"},
			{Georgia, "Georgia", "`7MN.   `7MF'"},
			{Alligator, "Alligator", "::::    :::  ::::::::"},
			{ANSI, "ANSI", "███    ██"},
			{ANSIShadow, "ANSIShadow", "███╗   ██╗"},
		}
		for _, l := range logos {
			t.Run(l.name, func(t *testing.T) {
				s := l.logo.String()
				if s == "" {
					t.Error("logo string should not be empty")
				}
				if !strings.Contains(s, l.contains) {
					t.Errorf("logo string does not contain expected content: %q", l.contains)
				}
			})
		}

		t.Run("Colored", func(t *testing.T) {
			withColor(t, func(r *lipgloss.Renderer) {
				logo := Georgia
				plain := logo.String()
				colored := logo.Colored()

				if colored == "" {
					t.Fatal("colored logo is empty")
				}
				if plain == colored {
					t.Error("Colored logo should be different from plain")
				}
				if !strings.Contains(colored, "\u001b[") {
					t.Error("Colored logo should contain ANSI escape codes")
				}
			})
		})

		t.Run("Colored in Viewport", func(t *testing.T) {
			withColor(t, func(r *lipgloss.Renderer) {
				logo := Colossal
				viewport := logo.ColoredInViewport(r)

				if viewport == "" {
					t.Fatal("viewport is empty")
				}

				cleanedViewport := stripAnsi(viewport)
				if !strings.Contains(cleanedViewport, "888") {
					t.Error("Viewport should contain parts of the logo")
				}

				borderChars := []string{"╭", "╮", "╯", "╰"}
				for _, char := range borderChars {
					if !strings.Contains(viewport, char) {
						t.Errorf("Viewport should contain rounded border character %s", char)
					}
				}
			})
		})
	})

}
