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
		Use:     "add [description]",
		Short:   "Add a new task",
		Aliases: []string{"create", "new"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.CreateTask(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.ListTasks(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "view [task-id]",
		Short: "View task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.ViewTask(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "update [task-id] [options...]",
		Short: "Update task properties",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.UpdateTask(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "delete [task-id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.DeleteTask(cmd.Context(), args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "projects",
		Short:   "List projects",
		Aliases: []string{"proj"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing projects...")
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "tags",
		Short:   "List tags",
		Aliases: []string{"t"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing tags...")
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "contexts",
		Short:   "List contexts (locations)",
		Aliases: []string{"loc", "ctx", "locations"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing task contexts...")
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:     "done [task-id]",
		Short:   "Mark task as completed",
		Aliases: []string{"complete"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.DoneTask(cmd.Context(), args)
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
