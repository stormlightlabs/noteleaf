// See https://patorjk.com/software/taag/
//
// NOTE: these aren't used anymore but are left in because they're cool
package ui

import (
	_ "embed"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type Logo int

const (
	Colossal Logo = iota
	Georgia
	Alligator
	ANSI
	ANSIShadow
)

const colossal string = `
888b    888          888            888                    .d888
8888b   888          888            888                   d88P"
88888b  888          888            888                   888
888Y88b 888  .d88b.  888888 .d88b.  888  .d88b.   8888b.  888888
888 Y88b888 d88""88b 888   d8P  Y8b 888 d8P  Y8b     "88b 888
888  Y88888 888  888 888   88888888 888 88888888 .d888888 888
888   Y8888 Y88..88P Y88b. Y8b.     888 Y8b.     888  888 888
888    Y888  "Y88P"   "Y888 "Y8888  888  "Y8888  "Y888888 888
`

const alligator string = `
::::    :::  :::::::: ::::::::::: :::::::::: :::        ::::::::::     :::     ::::::::::
:+:+:   :+: :+:    :+:    :+:     :+:        :+:        :+:          :+: :+:   :+:
:+:+:+  +:+ +:+    +:+    +:+     +:+        +:+        +:+         +:+   +:+  +:+
+#+ +:+ +#+ +#+    +:+    +#+     +#++:++#   +#+        +#++:++#   +#++:++#++: :#::+::#
+#+  +#+#+# +#+    +#+    +#+     +#+        +#+        +#+        +#+     +#+ +#+
#+#   #+#+# #+#    #+#    #+#     #+#        #+#        #+#        #+#     #+# #+#
###    ####  ########     ###     ########## ########## ########## ###     ### ###
`

const ansi = `
███    ██  ██████  ████████ ███████ ██      ███████  █████  ███████
████   ██ ██    ██    ██    ██      ██      ██      ██   ██ ██
██ ██  ██ ██    ██    ██    █████   ██      █████   ███████ █████
██  ██ ██ ██    ██    ██    ██      ██      ██      ██   ██ ██
██   ████  ██████     ██    ███████ ███████ ███████ ██   ██ ██
`

const ansiShadow = `
███╗   ██╗ ██████╗ ████████╗███████╗██╗     ███████╗ █████╗ ███████╗
████╗  ██║██╔═══██╗╚══██╔══╝██╔════╝██║     ██╔════╝██╔══██╗██╔════╝
██╔██╗ ██║██║   ██║   ██║   █████╗  ██║     █████╗  ███████║█████╗
██║╚██╗██║██║   ██║   ██║   ██╔══╝  ██║     ██╔══╝  ██╔══██║██╔══╝
██║ ╚████║╚██████╔╝   ██║   ███████╗███████╗███████╗██║  ██║██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝
`

//go:embed art/georgia.txt
var georgia string

func (l Logo) String() string {
	switch l {
	case Colossal:
		return colossal
	case Georgia:
		return georgia
	case Alligator:
		return alligator
	case ANSI:
		return ansi
	case ANSIShadow:
		return ansiShadow
	default:
		return colossal
	}
}

// Colored returns a colored version of the logo using lipgloss with vertical spiral design
// Creates a vertical spiral effect by coloring character by character:
//
//	Combine line position and character position & use modulo to build wave-like transitions
func (l Logo) Colored() string {
	logo := l.String()
	lines := strings.Split(logo, "\n")

	emeraldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10b981"))
	skyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0284c7"))

	var coloredLines []string
	for lineIdx, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for charIdx, char := range line {

			spiralPos := (lineIdx*3 + charIdx) % 8

			if spiralPos < 4 {
				coloredLine.WriteString(emeraldStyle.Render(string(char)))
			} else {
				coloredLine.WriteString(skyStyle.Render(string(char)))
			}
		}

		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

// ColoredInViewport returns the colored logo rendered inside a viewport bubble
func (l Logo) ColoredInViewport(renderer ...*lipgloss.Renderer) string {
	coloredLogo := l.Colored()
	lines := strings.Split(coloredLogo, "\n")

	maxWidth := 0
	for _, line := range lines {
		stripped := lipgloss.Width(line)
		if stripped > maxWidth {
			maxWidth = stripped
		}
	}

	vp := viewport.New(maxWidth+4, len(lines))
	vp.SetContent(coloredLogo)

	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6b7280")).PaddingLeft(1)

	if len(renderer) > 0 && renderer[0] != nil {
		style = style.Renderer(renderer[0])
	}

	return style.Render(vp.View())
}
