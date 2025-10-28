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
	"github.com/stormlightlabs/noteleaf/tools"
)

var (
	newTaskHandler    = handlers.NewTaskHandler
	newMovieHandler   = handlers.NewMovieHandler
	newTVHandler      = handlers.NewTVHandler
	newNoteHandler    = handlers.NewNoteHandler
	newBookHandler    = handlers.NewBookHandler
	newArticleHandler = handlers.NewArticleHandler
	exc               = fang.Execute
)

// App represents the main CLI application
type App struct {
	db     *store.Database
	config *store.Config
}

// NewApp creates a new CLI application instance ([App])
func NewApp() (*App, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &App{db, config}, nil
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
		Long: `Display comprehensive application status information.

Shows database location, configuration file path, data directories, and current
settings. Use this command to verify your noteleaf installation and diagnose
configuration issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Status(cmd.Context(), args, cmd.OutOrStdout())
		},
	}
}

func resetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset the application (removes all data)",
		Long: `Remove all application data and return to initial state.

This command deletes the database, all media files, notes, and articles. The
configuration file is preserved. Use with caution as this operation cannot be
undone. You will be prompted for confirmation before deletion proceeds.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Reset(cmd.Context(), args)
		},
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "noteleaf",
		Short: "A TaskWarrior-inspired CLI with notes, media queues and reading lists",
		Long: `noteleaf - personal information manager for the command line

A comprehensive CLI tool for managing tasks, notes, articles, and media queues.
Inspired by TaskWarrior, noteleaf combines todo management with reading lists,
watch queues, and a personal knowledge base.

Core features include hierarchical tasks with dependencies, recurring tasks,
time tracking, markdown notes with tags, article archiving, and media queue
management for books, movies, and TV shows.`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.Help()
			}

			output := strings.Join(args, " ")
			fmt.Fprintln(c.OutOrStdout(), output)
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
		Long: `Initialize noteleaf for first use.

Creates the database, configuration file, and required data directories. Run
this command after installing noteleaf or when setting up a new environment.
Safe to run multiple times as it will skip existing resources.`,
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
	handler, err := handlers.NewConfigHandler()
	if err != nil {
		log.Fatalf("failed to create config handler: %v", err)
	}
	return NewConfigCommand(handler).Create()
}

func run() int {
	logger := utils.NewLogger("info", "text")
	utils.Logger = logger

	app, err := NewApp()
	if err != nil {
		logger.Error("Failed to initialize application", "error", err)
		return 1
	}
	defer app.Close()

	taskHandler, err := newTaskHandler()
	if err != nil {
		log.Error("failed to create task handler", "err", err)
		return 1
	}

	movieHandler, err := newMovieHandler()
	if err != nil {
		log.Error("failed to create movie handler", "err", err)
		return 1
	}

	tvHandler, err := newTVHandler()
	if err != nil {
		log.Error("failed to create TV handler", "err", err)
		return 1
	}

	noteHandler, err := newNoteHandler()
	if err != nil {
		log.Error("failed to create note handler", "err", err)
		return 1
	}

	bookHandler, err := newBookHandler()
	if err != nil {
		log.Error("failed to create book handler", "err", err)
		return 1
	}

	articleHandler, err := newArticleHandler()
	if err != nil {
		log.Error("failed to create article handler", "err", err)
		return 1
	}

	root := rootCmd()

	coreGroups := []CommandGroup{
		NewTaskCommand(taskHandler), NewNoteCommand(noteHandler), NewArticleCommand(articleHandler),
	}

	for _, group := range coreGroups {
		cmd := group.Create()
		cmd.GroupID = "core"
		root.AddCommand(cmd)
	}

	mediaCmd := &cobra.Command{
		Use:   "media",
		Short: "Manage media queues (books, movies, TV shows)",
		Long: `Track and manage reading lists and watch queues.

Organize books, movies, and TV shows you want to consume. Search external
databases to add items, track reading/watching progress, and maintain a
history of completed media.`,
	}
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

	root.AddCommand(tools.NewDocGenCommand(root))

	opts := []fang.Option{
		fang.WithVersion("0.1.0"),
		fang.WithoutCompletions(),
		fang.WithColorSchemeFunc(ui.NoteleafColorScheme),
	}

	if err := exc(context.Background(), root, opts...); err != nil {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
