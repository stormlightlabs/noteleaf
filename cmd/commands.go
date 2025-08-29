/*
TODO: Implement movie addition
TODO: Implement movie listing
TODO: Implement movie watched status
TODO: Implement movie removal
TODO: Implement TV show addition
TODO: Implement TV show listing
TODO: Implement TV show watched status
TODO: Implement TV show removal
TODO: Implement config management
*/
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "noteleaf",
		Long:  ui.Georgia.ColoredInViewport(),
		Short: "A TaskWarrior-inspired CLI with notes, media queues and reading lists",
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.Help()
			}

			output := strings.Join(args, " ")
			fmt.Println(output)
			return nil
		},
	}

	root.SetHelpCommand(&cobra.Command{Hidden: true})
	cobra.EnableCommandSorting = false

	root.AddGroup(&cobra.Group{ID: "core", Title: "Core Commands:"})
	root.AddGroup(&cobra.Group{ID: "management", Title: "Management Commands:"})
	return root
}

func todoCmd() *cobra.Command {
	root := &cobra.Command{Use: "todo", Aliases: []string{"task"}, Short: "task management"}

	handler, err := handlers.NewTaskHandler()
	if err != nil {
		log.Fatalf("failed to create task handler: %v", err)
	}

	for _, init := range []func(*handlers.TaskHandler) *cobra.Command{
		addTaskCmd, listTaskCmd, viewTaskCmd, updateTaskCmd, editTaskCmd,
		deleteTaskCmd, taskProjectsCmd, taskTagsCmd, taskContextsCmd,
		taskCompleteCmd, taskStartCmd, taskStopCmd, timesheetViewCmd,
	} {
		cmd := init(handler)
		root.AddCommand(cmd)
	}

	return root
}

func mediaCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "media", Short: "Manage media queues (books, movies, TV shows)"}
	for _, init := range []func() *cobra.Command{bookMediaCmd, movieMediaCmd, tvMediaCmd} {
		cmd.AddCommand(init())
	}
	return cmd
}

func movieMediaCmd() *cobra.Command {
	root := &cobra.Command{Use: "movie", Short: "Manage movie watch queue"}

	root.AddCommand(&cobra.Command{
		Use:   "add [title]",
		Short: "Add movie to watch queue",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			fmt.Printf("Adding movie: %s\n", title)
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List movies in queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing movies...")
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "watched [id]",
		Short:   "Mark movie as watched",
		Aliases: []string{"seen"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Marking movie %s as watched\n", args[0])
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [id]",
		Short:   "Remove movie from queue",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Removing movie %s from queue\n", args[0])
			return nil
		},
	})

	return root
}

func tvMediaCmd() *cobra.Command {
	root := &cobra.Command{Use: "tv", Short: "Manage TV show watch queue"}

	root.AddCommand(&cobra.Command{
		Use:   "add [title]",
		Short: "Add TV show to watch queue",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			fmt.Printf("Adding TV show: %s\n", title)
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List TV shows in queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing TV shows...")
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "watched [id]",
		Short:   "Mark TV show/episodes as watched",
		Aliases: []string{"seen"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Marking TV show %s as watched\n", args[0])
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove [id]",
		Short:   "Remove TV show from queue",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Removing TV show %s from queue\n", args[0])
			return nil
		},
	})

	return root
}

func bookMediaCmd() *cobra.Command {
	root := &cobra.Command{Use: "book", Short: "Manage reading list"}

	addCmd := &cobra.Command{
		Use:   "add [search query...]",
		Short: "Search and add book to reading list",
		Long: `Search for books and add them to your reading list.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool("interactive")
			return handlers.SearchAndAddWithOptions(cmd.Context(), args, interactive)
		},
	}
	addCmd.Flags().BoolP("interactive", "i", false, "Use interactive interface for book selection")
	root.AddCommand(addCmd)

	root.AddCommand(&cobra.Command{
		Use:   "list [--all|--reading|--finished|--queued]",
		Short: "Show reading queue with progress",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.ListBooks(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "reading <id>",
		Short: "Mark book as currently reading",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.UpdateBookStatus(cmd.Context(), []string{args[0], "reading"})
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "finished <id>",
		Short:   "Mark book as completed",
		Aliases: []string{"read"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.UpdateBookStatus(cmd.Context(), []string{args[0], "finished"})
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "remove <id>",
		Short:   "Remove from reading list",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.UpdateBookStatus(cmd.Context(), []string{args[0], "removed"})
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "progress <id> <percentage>",
		Short: "Update reading progress percentage (0-100)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.UpdateBookProgress(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "update <id> <status>",
		Short: "Update book status (queued|reading|finished|removed)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.UpdateBookStatus(cmd.Context(), args)
		},
	})

	return root
}

func noteCmd() *cobra.Command {
	root := &cobra.Command{Use: "note", Short: "Manage notes"}

	handler, err := handlers.NewNoteHandler()
	if err != nil {
		log.Fatalf("failed to instantiate note handler: %v", err)
	}

	createCmd := &cobra.Command{
		Use:     "create [title] [content...]",
		Short:   "Create a new note",
		Aliases: []string{"new"},
		RunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool("interactive")
			filePath, _ := cmd.Flags().GetString("file")

			var title, content string
			if len(args) > 0 {
				title = args[0]
			}
			if len(args) > 1 {
				content = strings.Join(args[1:], " ")
			}

			if err != nil {
				return err
			}
			defer handler.Close()
			return handler.Create(cmd.Context(), title, content, filePath, interactive)
		},
	}
	createCmd.Flags().BoolP("interactive", "i", false, "Open interactive editor")
	createCmd.Flags().StringP("file", "f", "", "Create note from markdown file")
	root.AddCommand(createCmd)

	listCmd := &cobra.Command{
		Use:     "list [--archived] [--tags=tag1,tag2]",
		Short:   "Opens interactive TUI browser for navigating and viewing notes",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			archived, _ := cmd.Flags().GetBool("archived")
			tagsStr, _ := cmd.Flags().GetString("tags")

			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}

			handler, err := handlers.NewNoteHandler()
			if err != nil {
				return err
			}
			defer handler.Close()
			return handler.List(cmd.Context(), false, archived, tags)
		},
	}
	listCmd.Flags().BoolP("archived", "a", false, "Show archived notes")
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
			handler, err := handlers.NewNoteHandler()
			if err != nil {
				return err
			}
			defer handler.Close()
			return handler.View(cmd.Context(), noteID)
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
			handler, err := handlers.NewNoteHandler()
			if err != nil {
				return err
			}
			defer handler.Close()
			return handler.Edit(cmd.Context(), noteID)
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
			handler, err := handlers.NewNoteHandler()
			if err != nil {
				return err
			}
			defer handler.Close()
			return handler.Delete(cmd.Context(), noteID)
		},
	})

	return root
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show application status and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Status(cmd.Context(), args)
		},
	}
}

func resetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset the application (removes all data)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Reset(cmd.Context(), args)
		},
	}
}

func setupCmd() *cobra.Command {
	handler, err := handlers.NewSeedHandler()
	if err != nil {
		log.Fatalf("failed to instantiate seed handler: %v", err)
	}

	root := &cobra.Command{
		Use:   "setup",
		Short: "Initialize and manage application setup",
		RunE: func(c *cobra.Command, args []string) error {
			return handlers.Setup(c.Context(), args)
		},
	}

	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Populate database with test data",
		Long:  "Add sample tasks, books, and notes to the database for testing and demonstration purposes",
		RunE: func(c *cobra.Command, args []string) error {
			force, _ := c.Flags().GetBool("force")
			return handler.Seed(c.Context(), force)
		},
	}
	seedCmd.Flags().BoolP("force", "f", false, "Clear existing data and re-seed")

	root.AddCommand(seedCmd)
	return root
}

func confCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config [key] [value]",
		Short: "Manage configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			fmt.Printf("Setting config %s = %s\n", key, value)
			return nil
		},
	}
}
