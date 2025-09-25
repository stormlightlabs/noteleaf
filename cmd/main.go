package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
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

	if config, err := store.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	} else {
		return &App{db, config}, nil
	}
}

// Close cleans up application resources
func (app *App) Close() error {
	if app.db != nil {
		return app.db.Close()
	}
	return nil
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

func main() {
	logger := utils.NewLogger("info", "text")
	utils.Logger = logger

	app, err := NewApp()
	if err != nil {
		logger.Fatal("Failed to initialize application", "error", err)
	}
	defer app.Close()

	taskHandler, err := handlers.NewTaskHandler()
	if err != nil {
		log.Fatalf("failed to create task handler: %v", err)
	}

	movieHandler, err := handlers.NewMovieHandler()
	if err != nil {
		log.Fatalf("failed to create movie handler: %v", err)
	}

	tvHandler, err := handlers.NewTVHandler()
	if err != nil {
		log.Fatalf("failed to create TV handler: %v", err)
	}

	noteHandler, err := handlers.NewNoteHandler()
	if err != nil {
		log.Fatalf("failed to create note handler: %v", err)
	}

	bookHandler, err := handlers.NewBookHandler()
	if err != nil {
		log.Fatalf("failed to create book handler: %v", err)
	}

	articleHandler, err := handlers.NewArticleHandler()
	if err != nil {
		log.Fatalf("failed to create article handler: %v", err)
	}

	root := rootCmd()

	coreGroups := []CommandGroup{
		NewTaskCommand(taskHandler),
		NewNoteCommand(noteHandler),
		NewArticleCommand(articleHandler),
	}

	for _, group := range coreGroups {
		cmd := group.Create()
		cmd.GroupID = "core"
		root.AddCommand(cmd)
	}

	mediaCmd := &cobra.Command{Use: "media", Short: "Manage media queues (books, movies, TV shows)"}
	mediaCmd.GroupID = "core"
	mediaCmd.AddCommand(NewMovieCommand(movieHandler).Create())
	mediaCmd.AddCommand(NewTVCommand(tvHandler).Create())
	mediaCmd.AddCommand(NewBookCommand(bookHandler).Create())
	root.AddCommand(mediaCmd)

	mgmt := []func() *cobra.Command{statusCmd, confCmd, setupCmd, resetCmd}
	for _, cmdFunc := range mgmt {
		cmd := cmdFunc()
		cmd.GroupID = "management"
		root.AddCommand(cmd)
	}

	opts := []fang.Option{
		fang.WithVersion("0.1.0"),
		fang.WithoutCompletions(),
		fang.WithColorSchemeFunc(ui.NoteleafColorScheme),
	}

	if err := fang.Execute(context.Background(), root, opts...); err != nil {
		os.Exit(1)
	}
}
