package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// Setup initializes the application database and configuration
func Setup(ctx context.Context, args []string) error {
	logger := utils.GetLogger()
	logger.Info("Setting up noteleaf")

	configDir, err := store.GetConfigDir()
	if err != nil {
		logger.Error("Failed to get config directory", "error", err)
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	logger.Info("Using config directory", "path", configDir)
	fmt.Printf("Config directory: %s\n", configDir)

	// Load or create config to determine the actual database path
	config, err := store.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine database path using the same logic as NewDatabase
	var dbPath string
	if config.DatabasePath != "" {
		dbPath = config.DatabasePath
	} else if config.DataDir != "" {
		dbPath = filepath.Join(config.DataDir, "noteleaf.db")
	} else {
		dataDir, err := store.GetDataDir()
		if err != nil {
			return fmt.Errorf("failed to get data directory: %w", err)
		}
		dbPath = filepath.Join(dataDir, "noteleaf.db")
	}

	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Database already exists. Use --force to recreate.")
		return nil
	}

	db, err := store.NewDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	fmt.Printf("Database created: %s\n", db.GetPath())

	configPath, err := store.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	fmt.Printf("Configuration created: %s\n", configPath)
	fmt.Printf("Date format: %s\n", config.DateFormat)
	fmt.Printf("Color scheme: %s\n", config.ColorScheme)
	fmt.Printf("Default view: %s\n", config.DefaultView)

	runner := db.NewMigrationRunner()
	migrations, err := runner.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("Applied migrations: %d\n", len(migrations))
	for _, m := range migrations {
		fmt.Printf("  - %s (%s)\n", m.Version, m.AppliedAt)
	}

	fmt.Println("Setup completed successfully!")
	fmt.Println("\nYou can now use noteleaf commands:")
	fmt.Println("  noteleaf add \"Buy groceries\"")
	fmt.Println("  noteleaf list")
	fmt.Println("  noteleaf movie add \"The Matrix\"")

	return nil
}

// Reset recreates the database and configuration (destructive)
func Reset(ctx context.Context, args []string) error {
	fmt.Println("Resetting noteleaf...")

	// Load config to determine the actual database path
	config, err := store.LoadConfig()
	if err != nil {
		// If config doesn't exist, try to determine paths anyway
		config = store.DefaultConfig()
	}

	// Determine database path using the same logic as NewDatabase
	var dbPath string
	if config.DatabasePath != "" {
		dbPath = config.DatabasePath
	} else if config.DataDir != "" {
		dbPath = filepath.Join(config.DataDir, "noteleaf.db")
	} else {
		dataDir, err := store.GetDataDir()
		if err != nil {
			return fmt.Errorf("failed to get data directory: %w", err)
		}
		dbPath = filepath.Join(dataDir, "noteleaf.db")
	}

	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove database: %w", err)
	}

	configPath, err := store.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}

	fmt.Println("Reset completed. Run 'noteleaf setup' to reinitialize.")
	return nil
}

// Status shows the current application status
func Status(ctx context.Context, args []string) error {
	configDir, err := store.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	fmt.Printf("Config directory: %s\n", configDir)

	// Load config to determine the actual database path
	config, err := store.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine database path using the same logic as NewDatabase
	var dbPath string
	if config.DatabasePath != "" {
		dbPath = config.DatabasePath
	} else if config.DataDir != "" {
		dbPath = filepath.Join(config.DataDir, "noteleaf.db")
	} else {
		dataDir, err := store.GetDataDir()
		if err != nil {
			return fmt.Errorf("failed to get data directory: %w", err)
		}
		dbPath = filepath.Join(dataDir, "noteleaf.db")
	}

	if _, err := os.Stat(dbPath); err != nil {
		fmt.Println("Database: Not found")
		fmt.Println("Run 'noteleaf setup' to initialize.")
		return nil
	}

	fmt.Printf("Database: %s\n", dbPath)

	configPath, err := store.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		fmt.Println("Configuration: Not found")
	} else {
		fmt.Printf("Configuration: %s\n", configPath)
	}

	db, err := store.NewDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	runner := db.NewMigrationRunner()
	applied, err := runner.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	available, err := runner.GetAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %w", err)
	}

	fmt.Printf("Migrations: %d/%d applied\n", len(applied), len(available))

	return nil
}
