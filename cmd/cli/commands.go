package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/cmd/handlers"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "noteleaf",
		Long:  ui.Colossal.ColoredInViewport(),
		Short: "A TaskWarrior-inspired CLI with notes, media queues and reading lists",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			} else {
				output := strings.Join(args, " ")
				fmt.Println(output)
			}
		},
	}
}

func todoCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "todo",
		Short: "task management",
	}

	root.AddCommand(&cobra.Command{
		Use:   "add [description]",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			description := args[0]
			fmt.Printf("Adding task: %s\n", description)
			// TODO: Implement task creation
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing tasks...")
			// TODO: Implement task listing
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "done [task-id]",
		Short: "Mark task as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			fmt.Printf("Marking task %s as done\n", taskID)
			// TODO: Implement task completion
			return nil
		},
	})

	return root
}

func movieCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "movie",
		Short: "Manage movie watch queue",
	}

	root.AddCommand(&cobra.Command{
		Use:   "add [title]",
		Short: "Add movie to watch queue",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			fmt.Printf("Adding movie: %s\n", title)
			// TODO: Implement movie addition
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List movies in queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing movies...")
			// TODO: Implement movie listing
			return nil
		},
	})

	return root
}

func tvCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "tv",
		Short: "Manage TV show watch queue",
	}

	root.AddCommand(&cobra.Command{
		Use:   "add [title]",
		Short: "Add TV show to watch queue",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			fmt.Printf("Adding TV show: %s\n", title)
			// TODO: Implement TV show addition
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List TV shows in queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing TV shows...")
			// TODO: Implement TV show listing
			return nil
		},
	})

	return root
}

func bookCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "book",
		Short: "Manage reading list",
	}

	root.AddCommand(&cobra.Command{
		Use:   "add [title]",
		Short: "Add book to reading list",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := args[0]
			fmt.Printf("Adding book: %s\n", title)
			// TODO: Implement book addition
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List books in reading list",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing books...")
			// TODO: Implement book listing
			return nil
		},
	})

	return root
}

func noteCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "note",
		Short: "Manage notes",
	}

	root.AddCommand(&cobra.Command{
		Use:     "create [title] [content...]",
		Short:   "Create a new note",
		Aliases: []string{"new"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Create(cmd.Context(), args)
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
	return &cobra.Command{
		Use:   "setup",
		Short: "Initialize the application database and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Setup(cmd.Context(), args)
		},
	}
}

func confCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config [key] [value]",
		Short: "Manage configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			fmt.Printf("Setting config %s = %s\n", key, value)
			// TODO: Implement config management
			return nil
		},
	}
}
