package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
)

func parseID(k string, args []string) (int64, error) {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return id, fmt.Errorf("invalid %v ID: %s", k, args[0])
	}

	return id, err
}

// CommandGroup represents a group of related CLI commands
type CommandGroup interface {
	Create() *cobra.Command
}

// MovieCommand implements [CommandGroup] for movie-related commands
type MovieCommand struct {
	handler *handlers.MovieHandler
}

// NewMovieCommand creates a new MovieCommands with the given handler
func NewMovieCommand(handler *handlers.MovieHandler) *MovieCommand {
	return &MovieCommand{handler: handler}
}

func (c *MovieCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "movie",
		Short: "Manage movie watch queue",
		Long: `Track movies you want to watch.

Search TMDB for movies and add them to your queue. Mark movies as watched when
completed. Maintains a history of your movie watching activity.`,
	}

	addCmd := &cobra.Command{
		Use:   "add [search query...]",
		Short: "Search and add movie to watch queue",
		Long: `Search for movies and add them to your watch queue.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("search query cannot be empty")
			}
			interactive, _ := cmd.Flags().GetBool("interactive")
			query := strings.Join(args, " ")

			return c.handler.SearchAndAdd(cmd.Context(), query, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for movie selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--watched|--queued]",
		Short: "List movies in queue with status filtering",
		Long: `Display movies in your queue with optional status filters.

Shows movie titles, release years, and current status. Filter by --all to show
everything, --watched for completed movies, or --queued for unwatched items.
Default shows queued movies only.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var status string
			if len(args) > 0 {
				switch args[0] {
				case "--all":
					status = ""
				case "--watched":
					status = "watched"
				case "--queued":
					status = "queued"
				default:
					return fmt.Errorf("invalid status filter: %s (use: --all, --watched, --queued)", args[0])
				}
			}

			return c.handler.List(cmd.Context(), status)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "watched [id]",
		Short:   "Mark movie as watched",
		Aliases: []string{"seen"},
		Long:    "Mark a movie as watched with current timestamp. Moves the movie from queued to watched status.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.MarkWatched(cmd.Context(), args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [id]",
		Short:   "Remove movie from queue",
		Aliases: []string{"rm"},
		Long:    "Remove a movie from your watch queue. Use this for movies you no longer want to track.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.Remove(cmd.Context(), args[0])
		},
	})

	return root
}

// TVCommand implements [CommandGroup] for TV show-related commands
type TVCommand struct {
	handler *handlers.TVHandler
}

// NewTVCommand creates a new [TVCommand] with the given handler
func NewTVCommand(handler *handlers.TVHandler) *TVCommand {
	return &TVCommand{handler: handler}
}

func (c *TVCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "tv",
		Short: "Manage TV show watch queue",
		Long: `Track TV shows and episodes.

Search TMDB for TV shows and add them to your queue. Track which shows you're
currently watching, mark episodes as watched, and maintain a complete history
of your viewing activity.`,
	}

	addCmd := &cobra.Command{
		Use:   "add [search query...]",
		Short: "Search and add TV show to watch queue",
		Long: `Search for TV shows and add them to your watch queue.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("search query cannot be empty")
			}
			interactive, _ := cmd.Flags().GetBool("interactive")
			query := strings.Join(args, " ")

			return c.handler.SearchAndAdd(cmd.Context(), query, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for TV show selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--queued|--watching|--watched]",
		Short: "List TV shows in queue with status filtering",
		Long: `Display TV shows in your queue with optional status filters.

Shows show titles, air dates, and current status. Filter by --all, --queued,
--watching for shows in progress, or --watched for completed series. Default
shows queued shows only.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var status string
			if len(args) > 0 {
				switch args[0] {
				case "--all":
					status = ""
				case "--queued":
					status = "queued"
				case "--watching":
					status = "watching"
				case "--watched":
					status = "watched"
				default:
					return fmt.Errorf("invalid status filter: %s (use: --all, --queued, --watching, --watched)", args[0])
				}
			}

			return c.handler.List(cmd.Context(), status)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "watching [id]",
		Short: "Mark TV show as currently watching",
		Long:  "Mark a TV show as currently watching. Use this when you start watching a series.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.MarkTVShowWatching(cmd.Context(), args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "watched [id]",
		Short:   "Mark TV show/episodes as watched",
		Aliases: []string{"seen"},
		Long: `Mark TV show episodes or entire series as watched.

Updates episode tracking and completion status. Can mark individual episodes
or complete seasons/series depending on ID format.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.MarkWatched(cmd.Context(), args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [id]",
		Short:   "Remove TV show from queue",
		Aliases: []string{"rm"},
		Long:    "Remove a TV show from your watch queue. Use this for shows you no longer want to track.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.Remove(cmd.Context(), args[0])
		},
	})

	return root
}

// BookCommand implements [CommandGroup] for book-related commands
type BookCommand struct {
	handler *handlers.BookHandler
}

// NewBookCommand creates a new [BookCommand] with the given handler
func NewBookCommand(handler *handlers.BookHandler) *BookCommand {
	return &BookCommand{handler: handler}
}

func (c *BookCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "book",
		Short: "Manage reading list",
		Long: `Track books and reading progress.

Search Google Books API to add books to your reading list. Track which books
you're reading, update progress percentages, and maintain a history of finished
books.`,
	}

	addCmd := &cobra.Command{
		Use:   "add [search query...]",
		Short: "Search and add book to reading list",
		Long: `Search for books and add them to your reading list.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool("interactive")
			query := strings.Join(args, " ")
			return c.handler.SearchAndAdd(cmd.Context(), query, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for book selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--reading|--finished|--queued]",
		Short: "Show reading queue with progress",
		Long: `Display books in your reading list with progress indicators.

Shows book titles, authors, and reading progress percentages. Filter by --all,
--reading for books in progress, --finished for completed books, or --queued
for books not yet started. Default shows queued books only.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var status string
			if len(args) > 0 {
				switch args[0] {
				case "--all":
					status = ""
				case "--reading":
					status = "reading"
				case "--finished":
					status = "finished"
				case "--queued":
					status = "queued"
				default:
					return fmt.Errorf("invalid status filter: %s (use: --all, --reading, --finished, --queued)", args[0])
				}
			}
			return c.handler.List(cmd.Context(), status)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "reading <id>",
		Short: "Mark book as currently reading",
		Long:  "Mark a book as currently reading. Use this when you start a book from your queue.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateStatus(cmd.Context(), args[0], "reading")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "finished <id>",
		Short:   "Mark book as completed",
		Aliases: []string{"read"},
		Long:    "Mark a book as finished with current timestamp. Sets reading progress to 100%.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateStatus(cmd.Context(), args[0], "finished")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove <id>",
		Short:   "Remove from reading list",
		Aliases: []string{"rm"},
		Long:    "Remove a book from your reading list. Use this for books you no longer want to track.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateStatus(cmd.Context(), args[0], "removed")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "progress <id> <percentage>",
		Short: "Update reading progress percentage (0-100)",
		Long: `Set reading progress for a book.

Specify a percentage value between 0 and 100 to indicate how far you've
progressed through the book. Automatically updates status to 'reading' if not
already set.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			progress, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid progress percentage: %s", args[1])
			}
			return c.handler.UpdateProgress(cmd.Context(), args[0], progress)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "update <id> <status>",
		Short: "Update book status (queued|reading|finished|removed)",
		Long: `Change a book's status directly.

Valid statuses are: queued (not started), reading (in progress), finished
(completed), or removed (no longer tracking).`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateStatus(cmd.Context(), args[0], args[1])
		},
	})

	return root
}

// NoteCommand implements [CommandGroup] for note-related commands
type NoteCommand struct {
	handler *handlers.NoteHandler
}

// NewNoteCommand creates a new NoteCommand with the given handler
func NewNoteCommand(handler *handlers.NoteHandler) *NoteCommand {
	return &NoteCommand{handler: handler}
}

func (c *NoteCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "note",
		Short: "Manage notes",
		Long: `Create and organize markdown notes with tags.

Write notes in markdown format, organize them with tags, browse them in an
interactive TUI, and edit them in your preferred editor. Notes are stored as
files on disk with metadata tracked in the database.`,
	}

	createCmd := &cobra.Command{
		Use:     "create [title] [content...]",
		Short:   "Create a new note",
		Aliases: []string{"new"},
		Long: `Create a new markdown note.

Provide a title and optional content inline, or use --interactive to open an
editor. Use --file to import content from an existing markdown file. Notes
support tags for organization and full-text search.

Examples:
  noteleaf note create "Meeting notes" "Discussed project timeline"
  noteleaf note create -i
  noteleaf note create --file ~/documents/draft.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool("interactive")
			editor, _ := cmd.Flags().GetBool("editor")
			filePath, _ := cmd.Flags().GetString("file")

			var title, content string
			if len(args) > 0 {
				title = args[0]
			}
			if len(args) > 1 {
				content = strings.Join(args[1:], " ")
			}

			defer c.handler.Close()
			return c.handler.CreateWithOptions(cmd.Context(), title, content, filePath, interactive, editor)
		},
	}
	createCmd.Flags().BoolP("interactive", "i", false, "Open interactive editor")
	createCmd.Flags().BoolP("editor", "e", false, "Prompt to open note in editor after creation")
	createCmd.Flags().StringP("file", "f", "", "Create note from markdown file")
	root.AddCommand(createCmd)

	listCmd := &cobra.Command{
		Use:     "list [--archived] [--static] [--tags=tag1,tag2]",
		Short:   "Opens interactive TUI browser for navigating and viewing notes",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			archived, _ := cmd.Flags().GetBool("archived")
			static, _ := cmd.Flags().GetBool("static")
			tagsStr, _ := cmd.Flags().GetString("tags")

			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}

			defer c.handler.Close()
			return c.handler.List(cmd.Context(), static, archived, tags)
		},
	}
	listCmd.Flags().BoolP("archived", "a", false, "Show archived notes")
	listCmd.Flags().BoolP("static", "s", false, "Show static list instead of interactive TUI")
	listCmd.Flags().String("tags", "", "Filter by tags (comma-separated)")
	root.AddCommand(listCmd)

	root.AddCommand(&cobra.Command{
		Use:     "read [note-id]",
		Short:   "Display formatted note content with syntax highlighting",
		Aliases: []string{"view"},
		Long: `Display note content with formatted markdown rendering.

Shows the note with syntax highlighting, proper formatting, and metadata.
Useful for quick viewing without opening an editor.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if noteID, err := parseID("note", args); err != nil {
				return err
			} else {
				defer c.handler.Close()
				return c.handler.View(cmd.Context(), noteID)
			}
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "edit [note-id]",
		Short: "Edit note in configured editor",
		Long: `Open note in your configured text editor.

Uses the editor specified in your noteleaf configuration or the EDITOR
environment variable. Changes are automatically saved when you close the
editor.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if noteID, err := parseID("note", args); err != nil {
				return err
			} else {
				defer c.handler.Close()
				return c.handler.Edit(cmd.Context(), noteID)
			}
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [note-id]",
		Short:   "Permanently removes the note file and metadata",
		Aliases: []string{"rm", "delete", "del"},
		Long: `Delete a note permanently.

Removes both the markdown file and database metadata. This operation cannot be
undone. You will be prompted for confirmation before deletion.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if noteID, err := parseID("note", args); err != nil {
				return err
			} else {
				defer c.handler.Close()
				return c.handler.Delete(cmd.Context(), noteID)
			}
		},
	})

	return root
}

// ArticleCommand implements [CommandGroup] for article-related commands
type ArticleCommand struct {
	handler *handlers.ArticleHandler
}

// NewArticleCommand creates a new ArticleCommand with the given handler
func NewArticleCommand(handler *handlers.ArticleHandler) *ArticleCommand {
	return &ArticleCommand{handler: handler}
}

func (c *ArticleCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "article",
		Short: "Manage saved articles",
		Long: `Save and archive web articles locally.

Parse articles from supported websites, extract clean content, and save as
both markdown and HTML. Maintains a searchable archive of articles with
metadata including author, title, and publication date.`,
	}

	addCmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Parse and save article from URL",
		Long: `Parse and save article content from a supported website.

The article will be parsed using domain-specific XPath rules and saved
as both Markdown and HTML files. Article metadata is stored in the database.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			defer c.handler.Close()
			return c.handler.Add(cmd.Context(), args[0])
		},
	}
	root.AddCommand(addCmd)

	listCmd := &cobra.Command{
		Use:     "list [query]",
		Short:   "List saved articles",
		Aliases: []string{"ls"},
		Long: `List saved articles with optional filtering.

Use query to filter by title, or use flags for more specific filtering.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			author, _ := cmd.Flags().GetString("author")
			limit, _ := cmd.Flags().GetInt("limit")

			var query string
			if len(args) > 0 {
				query = strings.Join(args, " ")
			}

			defer c.handler.Close()
			return c.handler.List(cmd.Context(), query, author, limit)
		},
	}
	listCmd.Flags().String("author", "", "Filter by author")
	listCmd.Flags().IntP("limit", "l", 0, "Limit number of results (0 = no limit)")
	root.AddCommand(listCmd)

	viewCmd := &cobra.Command{
		Use:     "view <id>",
		Short:   "View article details and content preview",
		Aliases: []string{"show"},
		Long: `Display article metadata and summary.

Shows article title, author, publication date, URL, and a brief content
preview. Use 'read' command to view the full article content.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if articleID, err := parseID("article", args); err != nil {
				return err
			} else {
				defer c.handler.Close()
				return c.handler.View(cmd.Context(), articleID)
			}
		},
	}
	root.AddCommand(viewCmd)

	readCmd := &cobra.Command{
		Use:   "read <id>",
		Short: "Read article content with formatted markdown",
		Long: `Read the full markdown content of an article with beautiful formatting.

This displays the complete article content using syntax highlighting and proper formatting.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if articleID, err := parseID("article", args); err != nil {
				return err
			} else {
				defer c.handler.Close()
				return c.handler.Read(cmd.Context(), articleID)
			}
		},
	}
	root.AddCommand(readCmd)

	removeCmd := &cobra.Command{
		Use:     "remove <id>",
		Short:   "Remove article and associated files",
		Aliases: []string{"rm", "delete"},
		Long: `Delete an article and its files permanently.

Removes the article metadata from the database and deletes associated markdown
and HTML files. This operation cannot be undone.`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if articleID, err := parseID("article", args); err != nil {
				return err
			} else {
				defer c.handler.Close()
				return c.handler.Remove(cmd.Context(), articleID)
			}
		},
	}
	root.AddCommand(removeCmd)

	originalHelpFunc := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		originalHelpFunc(cmd, args)

		fmt.Println()
		defer c.handler.Close()
		c.handler.Help()
	})

	return root
}

// ConfigCommand implements [CommandGroup] for configuration management commands
type ConfigCommand struct {
	handler *handlers.ConfigHandler
}

// NewConfigCommand creates a new [ConfigCommand] with the given handler
func NewConfigCommand(handler *handlers.ConfigHandler) *ConfigCommand {
	return &ConfigCommand{handler: handler}
}

func (c *ConfigCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "config",
		Short: "Manage noteleaf configuration",
	}

	root.AddCommand(&cobra.Command{
		Use:   "get [key]",
		Short: "Get configuration value(s)",
		Long: `Display configuration values.

If no key is provided, displays all configuration values.
Otherwise, displays the value for the specified key.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var key string
			if len(args) > 0 {
				key = args[0]
			}
			return c.handler.Get(key)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration value",
		Long: `Update a configuration value.

Available keys:
  database_path      - Custom database file path
  data_dir           - Custom data directory
  date_format        - Date format string (default: 2006-01-02)
  color_scheme       - Color scheme (default: default)
  default_view       - Default view mode (default: list)
  default_priority   - Default task priority
  editor             - Preferred text editor
  articles_dir       - Articles storage directory
  notes_dir          - Notes storage directory
  auto_archive       - Auto-archive completed items (true/false)
  sync_enabled       - Enable synchronization (true/false)
  sync_endpoint      - Synchronization endpoint URL
  sync_token         - Synchronization token
  export_format      - Default export format (default: json)
  movie_api_key      - API key for movie database
  book_api_key       - API key for book database`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.Set(args[0], args[1])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  "Display the path to the configuration file being used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.Path()
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  "Reset all configuration values to their defaults.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.Reset()
		},
	})

	return root
}
