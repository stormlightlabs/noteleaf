package main

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/handlers"
)

func setupCommandTest(t *testing.T) func() {
	tempDir, err := os.MkdirTemp("", "noteleaf-cmd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = handlers.Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	return cleanup
}

func createTestTaskHandler(t *testing.T) (*handlers.TaskHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewTaskHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test task handler: %v", err)
	}
	return handler, func() {
		handler.Close()
		cleanup()
	}
}

func createTestMovieHandler(t *testing.T) (*handlers.MovieHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewMovieHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test movie handler: %v", err)
	}
	return handler, func() {
		handler.Close()
		cleanup()
	}
}

func createTestTVHandler(t *testing.T) (*handlers.TVHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewTVHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test TV handler: %v", err)
	}
	return handler, func() {
		handler.Close()
		cleanup()
	}
}

func createTestNoteHandler(t *testing.T) (*handlers.NoteHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewNoteHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test note handler: %v", err)
	}
	return handler, func() {
		handler.Close()
		cleanup()
	}
}

func createTestBookHandler(t *testing.T) (*handlers.BookHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewBookHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test book handler: %v", err)
	}
	return handler, func() {
		handler.Close()
		cleanup()
	}
}

func findSubcommand(commands []string, target string) bool {
	return slices.Contains(commands, target)
}

func TestCommandGroup(t *testing.T) {
	t.Run("Interface Implementations", func(t *testing.T) {
		taskHandler, taskCleanup := createTestTaskHandler(t)
		defer taskCleanup()

		movieHandler, movieCleanup := createTestMovieHandler(t)
		defer movieCleanup()

		tvHandler, tvCleanup := createTestTVHandler(t)
		defer tvCleanup()

		noteHandler, noteCleanup := createTestNoteHandler(t)
		defer noteCleanup()

		bookHandler, bookCleanup := createTestBookHandler(t)
		defer bookCleanup()

		var _ CommandGroup = NewTaskCommand(taskHandler)
		var _ CommandGroup = NewMovieCommand(movieHandler)
		var _ CommandGroup = NewTVCommand(tvHandler)
		var _ CommandGroup = NewNoteCommand(noteHandler)
		var _ CommandGroup = NewBookCommand(bookHandler)
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("TaskCommand", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			commands := NewTaskCommand(handler)
			cmd := commands.Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "todo" {
				t.Errorf("Expected Use to be 'todo', got '%s'", cmd.Use)
			}
			if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "task" {
				t.Errorf("Expected aliases to be ['task'], got %v", cmd.Aliases)
			}
			if cmd.Short != "task management" {
				t.Errorf("Expected Short to be 'task management', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}
		})

		t.Run("MovieCommand", func(t *testing.T) {
			handler, cleanup := createTestMovieHandler(t)
			defer cleanup()

			commands := NewMovieCommand(handler)
			cmd := commands.Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "movie" {
				t.Errorf("Expected Use to be 'movie', got '%s'", cmd.Use)
			}
			if cmd.Short != "Manage movie watch queue" {
				t.Errorf("Expected Short to be 'Manage movie watch queue', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}

			subcommands := cmd.Commands()
			subcommandNames := make([]string, len(subcommands))
			for i, subcmd := range subcommands {
				subcommandNames[i] = subcmd.Use
			}

			expectedSubcommands := []string{
				"add [search query...]",
				"list [--all|--watched|--queued]",
				"watched [id]",
				"remove [id]",
			}

			for _, expected := range expectedSubcommands {
				if !findSubcommand(subcommandNames, expected) {
					t.Errorf("Expected subcommand '%s' not found in %v", expected, subcommandNames)
				}
			}
		})

		t.Run("TVCommand", func(t *testing.T) {
			handler, cleanup := createTestTVHandler(t)
			defer cleanup()

			commands := NewTVCommand(handler)
			cmd := commands.Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "tv" {
				t.Errorf("Expected Use to be 'tv', got '%s'", cmd.Use)
			}
			if cmd.Short != "Manage TV show watch queue" {
				t.Errorf("Expected Short to be 'Manage TV show watch queue', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}

			subcommands := cmd.Commands()
			subcommandNames := make([]string, len(subcommands))
			for i, subcmd := range subcommands {
				subcommandNames[i] = subcmd.Use
			}

			expectedSubcommands := []string{
				"add [search query...]",
				"list [--all|--queued|--watching|--watched]",
				"watching [id]",
				"watched [id]",
				"remove [id]",
			}

			for _, expected := range expectedSubcommands {
				if !findSubcommand(subcommandNames, expected) {
					t.Errorf("Expected subcommand '%s' not found in %v", expected, subcommandNames)
				}
			}
		})

		t.Run("NoteCommand", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			commands := NewNoteCommand(handler)
			cmd := commands.Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "note" {
				t.Errorf("Expected Use to be 'note', got '%s'", cmd.Use)
			}
			if cmd.Short != "Manage notes" {
				t.Errorf("Expected Short to be 'Manage notes', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}

			subcommands := cmd.Commands()
			subcommandNames := make([]string, len(subcommands))
			for i, subcmd := range subcommands {
				subcommandNames[i] = subcmd.Use
			}

			expectedSubcommands := []string{
				"create [title] [content...]",
				"list [--archived] [--static] [--tags=tag1,tag2]",
				"read [note-id]",
				"edit [note-id]",
				"remove [note-id]",
			}

			for _, expected := range expectedSubcommands {
				if !findSubcommand(subcommandNames, expected) {
					t.Errorf("Expected subcommand '%s' not found in %v", expected, subcommandNames)
				}
			}
		})

		t.Run("BookCommand", func(t *testing.T) {
			handler, cleanup := createTestBookHandler(t)
			defer cleanup()

			commands := NewBookCommand(handler)
			cmd := commands.Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "book" {
				t.Errorf("Expected Use to be 'book', got '%s'", cmd.Use)
			}
			if cmd.Short != "Manage reading list" {
				t.Errorf("Expected Short to be 'Manage reading list', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}

			subcommands := cmd.Commands()
			subcommandNames := make([]string, len(subcommands))
			for i, subcmd := range subcommands {
				subcommandNames[i] = subcmd.Use
			}

			expectedSubcommands := []string{
				"add [search query...]",
				"list [--all|--reading|--finished|--queued]",
				"reading <id>",
				"finished <id>",
				"remove <id>",
				"progress <id> <percentage>",
				"update <id> <status>",
			}

			for _, expected := range expectedSubcommands {
				if !findSubcommand(subcommandNames, expected) {
					t.Errorf("Expected subcommand '%s' not found in %v", expected, subcommandNames)
				}
			}
		})

		t.Run("all command groups implement Create", func(t *testing.T) {
			taskHandler, taskCleanup := createTestTaskHandler(t)
			defer taskCleanup()

			movieHandler, movieCleanup := createTestMovieHandler(t)
			defer movieCleanup()

			tvHandler, tvCleanup := createTestTVHandler(t)
			defer tvCleanup()

			noteHandler, noteCleanup := createTestNoteHandler(t)
			defer noteCleanup()

			bookHandler, bookCleanup := createTestBookHandler(t)
			defer bookCleanup()

			groups := []CommandGroup{
				NewTaskCommand(taskHandler),
				NewMovieCommand(movieHandler),
				NewTVCommand(tvHandler),
				NewNoteCommand(noteHandler),
				NewBookCommand(bookHandler),
			}

			for i, group := range groups {
				cmd := group.Create()
				if cmd == nil {
					t.Errorf("CommandGroup %d returned nil from Create()", i)
					continue
				}
				if cmd.Use == "" {
					t.Errorf("CommandGroup %d returned command with empty Use", i)
				}
			}
		})
	})

}

func TestCommandExecution(t *testing.T) {
	t.Run("Movie Commands", func(t *testing.T) {
		handler, cleanup := createTestMovieHandler(t)
		defer cleanup()

		t.Run("list command - default", func(t *testing.T) {
			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"list"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("movie list command failed: %v", err)
			}
		})

		t.Run("add command with empty args", func(t *testing.T) {
			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"add"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected movie add command to fail with empty args")
			}
		})
	})

	t.Run("TV Commands", func(t *testing.T) {
		handler, cleanup := createTestTVHandler(t)
		defer cleanup()

		t.Run("list command - default", func(t *testing.T) {
			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"list"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("tv list command failed: %v", err)
			}
		})

		t.Run("add command with empty args", func(t *testing.T) {
			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"add"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected tv add command to fail with empty args")
			}
		})
	})

	t.Run("Book Commands", func(t *testing.T) {
		handler, cleanup := createTestBookHandler(t)
		defer cleanup()

		t.Run("list command - default", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"list"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("book list command failed: %v", err)
			}
		})
	})

	t.Run("Note Commands", func(t *testing.T) {

		t.Run("create command - non-interactive", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"create", "test title", "test content"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("note create command failed: %v", err)
			}
		})

		t.Run("list command - static mode", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--static"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("note list command failed: %v", err)
			}
		})
	})
}
