/*
TODO: Implement config management
*/
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
)

// CommandGroup represents a group of related CLI commands
type CommandGroup interface {
	Create() *cobra.Command
}

// MovieCommand implements CommandGroup for movie-related commands
type MovieCommand struct {
	handler *handlers.MovieHandler
}

// NewMovieCommand creates a new MovieCommands with the given handler
func NewMovieCommand(handler *handlers.MovieHandler) *MovieCommand {
	return &MovieCommand{handler: handler}
}

func (c *MovieCommand) Create() *cobra.Command {
	root := &cobra.Command{Use: "movie", Short: "Manage movie watch queue"}

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

			return c.handler.SearchAndAddMovie(cmd.Context(), query, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for movie selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--watched|--queued]",
		Short: "List movies in queue with status filtering",
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

			return c.handler.ListMovies(cmd.Context(), status)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "watched [id]",
		Short:   "Mark movie as watched",
		Aliases: []string{"seen"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.MarkMovieWatched(cmd.Context(), args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [id]",
		Short:   "Remove movie from queue",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.RemoveMovie(cmd.Context(), args[0])
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
	root := &cobra.Command{Use: "tv", Short: "Manage TV show watch queue"}

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

			return c.handler.SearchAndAddTV(cmd.Context(), query, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for TV show selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--queued|--watching|--watched]",
		Short: "List TV shows in queue with status filtering",
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

			return c.handler.ListTVShows(cmd.Context(), status)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "watching [id]",
		Short: "Mark TV show as currently watching",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.MarkTVShowWatching(cmd.Context(), args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "watched [id]",
		Short:   "Mark TV show/episodes as watched",
		Aliases: []string{"seen"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.MarkTVShowWatched(cmd.Context(), args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [id]",
		Short:   "Remove TV show from queue",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.RemoveTVShow(cmd.Context(), args[0])
		},
	})

	return root
}

// BookCommand implements CommandGroup for book-related commands
type BookCommand struct {
	handler *handlers.BookHandler
}

// NewBookCommand creates a new BookCommand with the given handler
func NewBookCommand(handler *handlers.BookHandler) *BookCommand {
	return &BookCommand{handler: handler}
}

func (c *BookCommand) Create() *cobra.Command {
	root := &cobra.Command{Use: "book", Short: "Manage reading list"}

	addCmd := &cobra.Command{
		Use:   "add [search query...]",
		Short: "Search and add book to reading list",
		Long: `Search for books and add them to your reading list.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool("interactive")
			return c.handler.SearchAndAddBook(cmd.Context(), args, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for book selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--reading|--finished|--queued]",
		Short: "Show reading queue with progress",
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
			return c.handler.ListBooks(cmd.Context(), status)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "reading <id>",
		Short: "Mark book as currently reading",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateBookStatusByID(cmd.Context(), args[0], "reading")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "finished <id>",
		Short:   "Mark book as completed",
		Aliases: []string{"read"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateBookStatusByID(cmd.Context(), args[0], "finished")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove <id>",
		Short:   "Remove from reading list",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateBookStatusByID(cmd.Context(), args[0], "removed")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "progress <id> <percentage>",
		Short: "Update reading progress percentage (0-100)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			progress, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid progress percentage: %s", args[1])
			}
			return c.handler.UpdateBookProgressByID(cmd.Context(), args[0], progress)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "update <id> <status>",
		Short: "Update book status (queued|reading|finished|removed)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.UpdateBookStatusByID(cmd.Context(), args[0], args[1])
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
	root := &cobra.Command{Use: "note", Short: "Manage notes"}

	createCmd := &cobra.Command{
		Use:     "create [title] [content...]",
		Short:   "Create a new note",
		Aliases: []string{"new"},
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
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note ID: %s", args[0])
			}
			defer c.handler.Close()
			return c.handler.View(cmd.Context(), noteID)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "edit [note-id]",
		Short: "Edit note in configured editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note ID: %s", args[0])
			}
			defer c.handler.Close()
			return c.handler.Edit(cmd.Context(), noteID)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [note-id]",
		Short:   "Permanently removes the note file and metadata",
		Aliases: []string{"rm", "delete", "del"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note ID: %s", args[0])
			}
			defer c.handler.Close()
			return c.handler.Delete(cmd.Context(), noteID)
		},
	})

	return root
}
