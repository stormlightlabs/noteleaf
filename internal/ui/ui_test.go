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
			{Cherry, "#FF388B"},
			{Julep, "#00FFB2"},
			{Butter, "#FFFAF1"},
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
			{Cherry, "Cherry"},
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
			{Collosal, "Collosal", "888b    888"},
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
				logo := Collosal
				viewport := logo.ColoredInViewport(r)
				t.Logf("viewport output:\n%s", viewport)

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
