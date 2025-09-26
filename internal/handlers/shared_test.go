package handlers

import (
	"strings"
	"testing"
)

type renderMarkdownTC struct {
	name     string
	content  string
	err      bool
	contains []string
}

func TestRenderMarkdown(t *testing.T) {
	tt := []renderMarkdownTC{
		{name: "simple text", content: "Hello, world!", err: false, contains: []string{"Hello, world!"}},
		{name: "markdown heading", content: "# Main Title", err: false, contains: []string{"Main Title"}},
		{name: "markdown with emphasis", content: "This is **bold** and *italic* text", err: false, contains: []string{"bold", "italic"}},
		{name: "markdown list", content: "- Item 1\n- Item 2\n- Item 3", err: false, contains: []string{"Item 1", "Item 2", "Item 3"}},
		{name: "code block", content: "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```", err: false, contains: []string{"main", "fmt.Println"}},
		{name: "empty string", content: "", err: false, contains: []string{}},
		{name: "only whitespace", content: "   \n\t  \n   ", err: false, contains: []string{}},
		{name: "mixed content", content: "# Title\n\nSome **bold** text and a [link](https://example.com)\n\n- List item", err: false,
			contains: []string{"Title", "bold", "example.com", "List item"}}}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderMarkdown(tt.content)
			if tt.err && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.contains {
				if strings.Contains(result, want) {
					continue
				}
				t.Fatalf("result should contain %q, got:\n%s", want, result)
			}
		})
	}

	t.Run("WordWrap", func(t *testing.T) {
		text := strings.Repeat("This is a very long line that should be wrapped at 80 characters. ", 5)
		result, err := renderMarkdown(text)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(result, "\n")
		for i, line := range lines {
			cleaned := removeANSI(line)
			if len(cleaned) > 85 {
				t.Fatalf("Line at index %d is too long (%d chars): %q", i, len(cleaned), cleaned)
			}
		}
	})
}

func removeANSI(s string) string {
	result := ""
	inEscapeCh := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscapeCh = true
			i++
			continue
		}
		if inEscapeCh {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscapeCh = false
			}
			continue
		}
		result += string(s[i])
	}
	return result
}
