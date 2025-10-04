package handlers

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

type MarkdownRenderer interface {
	Render(string) (string, error)
}

var newRenderer = func() (MarkdownRenderer, error) {
	return glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(80))
}

func renderMarkdown(content string) (string, error) {
	renderer, err := newRenderer()
	if err != nil {
		return "", fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}
	return rendered, nil
}
