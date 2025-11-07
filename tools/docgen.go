//go:build !prod

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// NewDocGenCommand creates a hidden command for generating CLI documentation
func NewDocGenCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "docgen",
		Short:  "Generate CLI documentation",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("out")
			format, _ := cmd.Flags().GetString("format")
			front, _ := cmd.Flags().GetBool("frontmatter")

			if err := os.MkdirAll(out, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			root.DisableAutoGenTag = true

			switch format {
			case "docusaurus":
				if err := generateDocusaurusDocs(root, out); err != nil {
					return fmt.Errorf("failed to generate docusaurus documentation: %w", err)
				}
			case "markdown":
				if front {
					prep := func(filename string) string {
						base := filepath.Base(filename)
						name := strings.TrimSuffix(base, filepath.Ext(base))
						title := strings.ReplaceAll(name, "_", " ")
						return fmt.Sprintf("---\ntitle: %q\nslug: %q\ndescription: \"CLI reference for %s\"\n---\n\n", title, name, title)
					}
					link := func(name string) string { return strings.ToLower(name) }
					if err := doc.GenMarkdownTreeCustom(root, out, prep, link); err != nil {
						return fmt.Errorf("failed to generate markdown documentation: %w", err)
					}
				} else {
					if err := doc.GenMarkdownTree(root, out); err != nil {
						return fmt.Errorf("failed to generate markdown documentation: %w", err)
					}
				}
			case "man":
				hdr := &doc.GenManHeader{Title: strings.ToUpper(root.Name()), Section: "1"}
				if err := doc.GenManTree(root, hdr, out); err != nil {
					return fmt.Errorf("failed to generate man pages: %w", err)
				}
			case "rest":
				if err := doc.GenReSTTree(root, out); err != nil {
					return fmt.Errorf("failed to generate ReStructuredText documentation: %w", err)
				}
			default:
				return fmt.Errorf("unknown format: %s", format)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Documentation generated in %s\n", out)
			return nil
		},
	}

	cmd.Flags().StringP("out", "o", "./docs/cli", "output directory")
	cmd.Flags().StringP("format", "f", "markdown", "output format (docusaurus|markdown|man|rest)")
	cmd.Flags().Bool("frontmatter", false, "prepend simple YAML front matter to markdown")

	return cmd
}

// CategoryJSON represents the _category_.json structure for Docusaurus
type CategoryJSON struct {
	Label       string `json:"label"`
	Position    int    `json:"position"`
	Link        *Link  `json:"link,omitempty"`
	Collapsed   bool   `json:"collapsed,omitempty"`
	Description string `json:"description,omitempty"`
}

// Link represents a link in _category_.json
type Link struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// generateDocusaurusDocs creates combined, Docusaurus-compatible documentation
func generateDocusaurusDocs(root *cobra.Command, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	category := CategoryJSON{
		Label:    "CLI Reference",
		Position: 3,
		Link: &Link{
			Type:        "generated-index",
			Description: "Complete command-line reference for noteleaf",
		},
	}
	categoryJSON, err := json.MarshalIndent(category, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal category json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "_category_.json"), categoryJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write category json: %w", err)
	}

	indexContent := generateIndexPage(root)
	if err := os.WriteFile(filepath.Join(outDir, "index.md"), []byte(indexContent), 0o644); err != nil {
		return fmt.Errorf("failed to write index.md: %w", err)
	}

	commandGroups := map[string]struct {
		title       string
		position    int
		commands    []string
		description string
	}{
		"tasks": {
			title:       "Task Management",
			position:    1,
			commands:    []string{"todo", "task"},
			description: "Manage tasks with TaskWarrior-inspired features",
		},
		"notes": {
			title:       "Notes",
			position:    2,
			commands:    []string{"note"},
			description: "Create and organize markdown notes",
		},
		"articles": {
			title:       "Articles",
			position:    3,
			commands:    []string{"article"},
			description: "Save and archive web articles",
		},
		"books": {
			title:       "Books",
			position:    4,
			commands:    []string{"media book"},
			description: "Manage reading list and track progress",
		},
		"movies": {
			title:       "Movies",
			position:    5,
			commands:    []string{"media movie"},
			description: "Track movies in watch queue",
		},
		"tv-shows": {
			title:       "TV Shows",
			position:    6,
			commands:    []string{"media tv"},
			description: "Manage TV show watching",
		},
		"configuration": {
			title:       "Configuration",
			position:    7,
			commands:    []string{"config"},
			description: "Manage application configuration",
		},
		"management": {
			title:       "Management",
			position:    8,
			commands:    []string{"status", "setup", "reset"},
			description: "Application management commands",
		},
	}

	for filename, group := range commandGroups {
		content := generateCombinedPage(root, group.title, group.position, group.commands, group.description)
		outputFile := filepath.Join(outDir, filename+".md")
		if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputFile, err)
		}
	}

	return nil
}

// generateIndexPage creates the index/overview page
func generateIndexPage(root *cobra.Command) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("id: index\n")
	b.WriteString("title: CLI Reference\n")
	b.WriteString("sidebar_label: Overview\n")
	b.WriteString("sidebar_position: 0\n")
	b.WriteString("description: Complete command-line reference for noteleaf\n")
	b.WriteString("---\n\n")

	b.WriteString("# noteleaf CLI Reference\n\n")

	if root.Long != "" {
		b.WriteString(root.Long)
		b.WriteString("\n\n")
	} else if root.Short != "" {
		b.WriteString(root.Short)
		b.WriteString("\n\n")
	}

	b.WriteString("## Usage\n\n")
	b.WriteString("```bash\n")
	b.WriteString(root.UseLine())
	b.WriteString("\n```\n\n")

	b.WriteString("## Command Groups\n\n")
	b.WriteString("- **[Task Management](tasks)** - Manage todos, projects, and time tracking\n")
	b.WriteString("- **[Notes](notes)** - Create and organize markdown notes\n")
	b.WriteString("- **[Articles](articles)** - Save and archive web articles\n")
	b.WriteString("- **[Books](books)** - Track reading list and progress\n")
	b.WriteString("- **[Movies](movies)** - Manage movie watch queue\n")
	b.WriteString("- **[TV Shows](tv-shows)** - Track TV show watching\n")
	b.WriteString("- **[Configuration](configuration)** - Manage settings\n")
	b.WriteString("- **[Management](management)** - Application management\n\n")

	return b.String()
}

// generateCombinedPage creates a combined documentation page for a command group
func generateCombinedPage(root *cobra.Command, title string, position int, commandPaths []string, description string) string {
	var b strings.Builder

	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: %s\n", slug))
	b.WriteString(fmt.Sprintf("title: %s\n", title))
	b.WriteString(fmt.Sprintf("sidebar_position: %d\n", position))
	b.WriteString(fmt.Sprintf("description: %s\n", description))
	b.WriteString("---\n\n")

	for _, cmdPath := range commandPaths {
		cmd := findCommand(root, strings.Split(cmdPath, " "))
		if cmd == nil {
			continue
		}

		b.WriteString(fmt.Sprintf("## %s\n\n", cmd.Name()))
		if cmd.Long != "" {
			b.WriteString(cmd.Long)
			b.WriteString("\n\n")
		} else if cmd.Short != "" {
			b.WriteString(cmd.Short)
			b.WriteString("\n\n")
		}

		b.WriteString("```bash\n")
		b.WriteString(cmd.UseLine())
		b.WriteString("\n```\n\n")

		if cmd.HasSubCommands() {
			b.WriteString("### Subcommands\n\n")
			for _, sub := range cmd.Commands() {
				if sub.Hidden {
					continue
				}
				generateSubcommandSection(&b, sub, 4)
			}
		}

		if cmd.HasFlags() {
			b.WriteString("### Options\n\n")
			b.WriteString("```\n")
			b.WriteString(cmd.Flags().FlagUsages())
			b.WriteString("```\n\n")
		}
	}

	return b.String()
}

// generateSubcommandSection generates documentation for a subcommand
func generateSubcommandSection(b *strings.Builder, cmd *cobra.Command, level int) {
	prefix := strings.Repeat("#", level)

	fmt.Fprintf(b, "%s %s\n\n", prefix, cmd.Name())

	if cmd.Long != "" {
		b.WriteString(cmd.Long)
		b.WriteString("\n\n")
	} else if cmd.Short != "" {
		b.WriteString(cmd.Short)
		b.WriteString("\n\n")
	}

	b.WriteString("**Usage:**\n\n")
	b.WriteString("```bash\n")
	b.WriteString(cmd.UseLine())
	b.WriteString("\n```\n\n")

	if cmd.HasLocalFlags() {
		b.WriteString("**Options:**\n\n")
		b.WriteString("```\n")
		b.WriteString(cmd.LocalFlags().FlagUsages())
		b.WriteString("```\n\n")
	}

	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(b, "**Aliases:** %s\n\n", strings.Join(cmd.Aliases, ", "))
	}

	if cmd.HasSubCommands() {
		for _, sub := range cmd.Commands() {
			if sub.Hidden {
				continue
			}
			generateSubcommandSection(b, sub, level+1)
		}
	}
}

// findCommand finds a command by path
func findCommand(root *cobra.Command, path []string) *cobra.Command {
	if len(path) == 0 {
		return root
	}

	for _, cmd := range root.Commands() {
		if cmd.Name() == path[0] || slices.Contains(cmd.Aliases, path[0]) {
			if len(path) == 1 {
				return cmd
			}
			return findCommand(cmd, path[1:])
		}
	}

	return nil
}
