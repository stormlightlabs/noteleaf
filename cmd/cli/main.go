package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
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
		return &App{db: db, config: config}, nil
	}
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

	root := rootCmd()
	commands := []func() *cobra.Command{
		setupCmd, resetCmd, statusCmd, todoCmd,
		movieCmd, noteCmd, tvCmd, bookCmd, confCmd,
	}

	for _, cmdFunc := range commands {
		cmd := cmdFunc()
		root.AddCommand(cmd)
	}

	options := []fang.Option{
		fang.WithVersion("0.1.0"),
		fang.WithoutCompletions(),
		fang.WithColorSchemeFunc(ui.NoteleafColorScheme),
	}

	if err := fang.Execute(context.Background(), root, options...); err != nil {
		os.Exit(1)
	}
}
