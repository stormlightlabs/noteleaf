package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/cmd/handlers"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// App represents the main CLI application
type App struct {
	db     *store.Database
	config *store.Config
}

// NewApp creates a new CLI application instance
func NewApp() (*App, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &App{
		db:     db,
		config: config,
	}, nil
}

// Close cleans up application resources
func (app *App) Close() error {
	if app.db != nil {
		return app.db.Close()
	}
	return nil
}

func main() {
	logger := utils.NewLogger("info", "text")
	utils.Logger = logger

	app, err := NewApp()
	if err != nil {
		logger.Fatal("Failed to initialize application", "error", err)
	}
	defer app.Close()

	rootCmd := &cobra.Command{
		Use:   "noteleaf",
		Short: "A TaskWarrior-inspired CLI with media queues and reading lists",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println(ui.Collosal.ColoredInViewport())
				cmd.Help()
				return
			}

			output := strings.Join(args, " ")
			fmt.Println(output)
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Initialize the application database and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Setup(cmd.Context(), args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset the application (removes all data)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Reset(cmd.Context(), args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show application status and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Status(cmd.Context(), args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
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

	rootCmd.AddCommand(&cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing tasks...")
			// TODO: Implement task listing
			return nil
		},
	})

	rootCmd.AddCommand(&cobra.Command{
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

	movieCmd := &cobra.Command{
		Use:   "movie",
		Short: "Manage movie watch queue",
	}

	movieCmd.AddCommand(&cobra.Command{
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

	movieCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List movies in queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing movies...")
			// TODO: Implement movie listing
			return nil
		},
	})

	rootCmd.AddCommand(movieCmd)

	tvCmd := &cobra.Command{
		Use:   "tv",
		Short: "Manage TV show watch queue",
	}

	tvCmd.AddCommand(&cobra.Command{
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

	tvCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List TV shows in queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing TV shows...")
			// TODO: Implement TV show listing
			return nil
		},
	})

	rootCmd.AddCommand(tvCmd)

	bookCmd := &cobra.Command{
		Use:   "book",
		Short: "Manage reading list",
	}

	bookCmd.AddCommand(&cobra.Command{
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

	bookCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List books in reading list",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing books...")
			// TODO: Implement book listing
			return nil
		},
	})

	rootCmd.AddCommand(bookCmd)

	noteCmd := &cobra.Command{
		Use:   "note",
		Short: "Manage notes",
	}

	noteCmd.AddCommand(&cobra.Command{
		Use:     "create [title] [content...]",
		Short:   "Create a new note",
		Aliases: []string{"new"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Create(cmd.Context(), args)
		},
	})

	rootCmd.AddCommand(noteCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "config [key] [value]",
		Short: "Manage configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			fmt.Printf("Setting config %s = %s\n", key, value)
			// TODO: Implement config management
			return nil
		},
	})

	if err := fang.Execute(context.Background(), rootCmd, fang.WithVersion("0.1.0")); err != nil {
		os.Exit(1)
	}
}
