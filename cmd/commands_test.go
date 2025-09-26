package main

import (
	"context"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/handlers"
	"github.com/stormlightlabs/noteleaf/internal/services"
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

func createTestArticleHandler(t *testing.T) (*handlers.ArticleHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewArticleHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test article handler: %v", err)
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

		articleHandler, articleCleanup := createTestArticleHandler(t)
		defer articleCleanup()

		var _ CommandGroup = NewTaskCommand(taskHandler)
		var _ CommandGroup = NewMovieCommand(movieHandler)
		var _ CommandGroup = NewTVCommand(tvHandler)
		var _ CommandGroup = NewNoteCommand(noteHandler)
		var _ CommandGroup = NewBookCommand(bookHandler)
		var _ CommandGroup = NewArticleCommand(articleHandler)
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

		t.Run("ArticleCommand", func(t *testing.T) {
			handler, cleanup := createTestArticleHandler(t)
			defer cleanup()

			commands := NewArticleCommand(handler)
			cmd := commands.Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "article" {
				t.Errorf("Expected Use to be 'article', got '%s'", cmd.Use)
			}
			if cmd.Short != "Manage saved articles" {
				t.Errorf("Expected Short to be 'Manage saved articles', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}

			subcommands := cmd.Commands()
			subcommandNames := make([]string, len(subcommands))
			for i, subcmd := range subcommands {
				subcommandNames[i] = subcmd.Use
			}

			for _, expected := range []string{"add <url>", "list [query]", "view <id>", "remove <id>"} {
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

			articleHandler, articleCleanup := createTestArticleHandler(t)
			defer articleCleanup()

			groups := []CommandGroup{
				NewTaskCommand(taskHandler),
				NewMovieCommand(movieHandler),
				NewTVCommand(tvHandler),
				NewNoteCommand(noteHandler),
				NewBookCommand(bookHandler),
				NewArticleCommand(articleHandler),
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

		t.Run("add command with valid args - successful search", func(t *testing.T) {
			cleanup := services.SetupSuccessfulMovieMocks(t)
			defer cleanup()

			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"add", "Fantastic Four"})
			err := cmd.Execute()

			// NOTE: The command will find results but fail due to no user input in test environment
			if err == nil {
				t.Error("expected movie add command to fail due to no user input in test environment")
			}
			if !strings.Contains(err.Error(), "invalid input") {
				t.Errorf("expected 'invalid input' error, got: %v", err)
			}
		})

		t.Run("add command with valid args - search failure", func(t *testing.T) {
			cleanup := services.SetupFailureMocks(t, "search failed")
			defer cleanup()

			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"add", "some movie"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected movie add command to fail when search fails")
			}
			services.AssertErrorContains(t, err, "search failed")
		})

		t.Run("remove command with non-existent movie ID", func(t *testing.T) {
			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "999"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected movie remove command to fail with non-existent ID")
			}
		})

		t.Run("remove command with non-numeric ID", func(t *testing.T) {
			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected movie remove command to fail with non-numeric ID")
			}
		})

		t.Run("watched command", func(t *testing.T) {
			handler, cleanup := createTestMovieHandler(t)
			defer cleanup()

			cmd := NewMovieCommand(handler).Create()
			cmd.SetArgs([]string{"watched", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected movie watched command to fail with non-existent ID")
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

		t.Run("add command with valid args - successful search", func(t *testing.T) {
			cleanup := services.SetupSuccessfulTVMocks(t)
			defer cleanup()

			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"add", "Peacemaker"})
			err := cmd.Execute()

			// NOTE: The command will find results but fail due to no user input in test environment
			if err == nil {
				t.Error("expected tv add command to fail due to no user input in test environment")
			}
			if !strings.Contains(err.Error(), "invalid input") {
				t.Errorf("expected 'invalid input' error, got: %v", err)
			}
		})

		t.Run("add command with valid args - search failure", func(t *testing.T) {
			cleanup := services.SetupFailureMocks(t, "tv search failed")
			defer cleanup()

			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"add", "some show"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected tv add command to fail when search fails")
			}
			services.AssertErrorContains(t, err, "tv search failed")
		})

		t.Run("remove command with non-existent TV show ID", func(t *testing.T) {
			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "999"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected tv remove command to fail with non-existent ID")
			}
		})

		t.Run("remove command with non-numeric ID", func(t *testing.T) {
			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected tv remove command to fail with non-numeric ID")
			}
		})

		t.Run("watching command", func(t *testing.T) {
			handler, cleanup := createTestTVHandler(t)
			defer cleanup()

			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"watching", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected tv watching command to fail with non-existent ID")
			}
		})

		t.Run("watched command", func(t *testing.T) {
			handler, cleanup := createTestTVHandler(t)
			defer cleanup()

			cmd := NewTVCommand(handler).Create()
			cmd.SetArgs([]string{"watched", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected tv watched command to fail with non-existent ID")
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

		t.Run("remove command with non-existent book ID", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "999"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book remove command to fail with non-existent ID")
			}
		})

		t.Run("remove command with non-numeric ID", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book remove command to fail with non-numeric ID")
			}
		})

		t.Run("update command with removed status", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"update", "999", "removed"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book update command to fail with non-existent ID")
			}
		})

		t.Run("update command with invalid status", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"update", "1", "invalid_status"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book update command to fail with invalid status")
			}
		})

		t.Run("reading command", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"reading", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book reading command to fail with non-existent ID")
			}
		})

		t.Run("finished command", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"finished", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book finished command to fail with non-existent ID")
			}
		})

		t.Run("progress command", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"progress", "1", "50"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book progress command to fail with non-existent ID")
			}
		})

		t.Run("progress command with invalid percentage", func(t *testing.T) {
			cmd := NewBookCommand(handler).Create()
			cmd.SetArgs([]string{"progress", "1", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected book progress command to fail with invalid percentage")
			}
		})
	})

	t.Run("Article Commands", func(t *testing.T) {
		handler, cleanup := createTestArticleHandler(t)
		defer cleanup()

		t.Run("list command - default", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"list"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("article list command failed: %v", err)
			}
		})

		t.Run("help command", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"help"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("article help command failed: %v", err)
			}
		})

		t.Run("add command with empty args", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"add"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article add command to fail with empty args")
			}
		})

		t.Run("add command with invalid URL", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"add", "not-a-url"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article add command to fail with invalid URL")
			}
		})

		t.Run("view command with non-existent article ID", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"view", "999"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article view command to fail with non-existent ID")
			}
		})

		t.Run("view command with non-numeric ID", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"view", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article view command to fail with non-numeric ID")
			}
		})

		t.Run("read command with non-existent article ID", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"read", "999"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article read command to fail with non-existent ID")
			}
		})

		t.Run("read command with non-numeric ID", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"read", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article read command to fail with non-numeric ID")
			}
		})

		t.Run("remove command with non-existent article ID", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "999"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article remove command to fail with non-existent ID")
			}
		})

		t.Run("remove command with non-numeric ID", func(t *testing.T) {
			cmd := NewArticleCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected article remove command to fail with non-numeric ID")
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

		t.Run("read command with valid note ID", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			err := handler.CreateWithOptions(context.Background(), "test note", "test content", "", false, false)
			if err != nil {
				t.Fatalf("failed to create test note: %v", err)
			}

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"read", "1"})
			err = cmd.Execute()
			if err != nil {
				t.Errorf("note read command failed: %v", err)
			}
		})

		t.Run("edit command with valid note ID", func(t *testing.T) {
			t.Skip("edit command requires interactive editor")
		})

		t.Run("remove command with valid note ID", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			err := handler.CreateWithOptions(context.Background(), "test note", "test content", "", false, false)
			if err != nil {
				t.Fatalf("failed to create test note: %v", err)
			}

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "1"})
			err = cmd.Execute()
			if err != nil {
				t.Errorf("note remove command failed: %v", err)
			}
		})

		t.Run("edit command with invalid ID", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"edit", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected note edit command to fail with invalid ID")
			}
		})

		t.Run("remove command with invalid ID", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"remove", "invalid"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected note remove command to fail with invalid ID")
			}
		})

		t.Run("list command with static flag", func(t *testing.T) {
			handler, cleanup := createTestNoteHandler(t)
			defer cleanup()

			cmd := NewNoteCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--static", "test query"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("note list command with query failed: %v", err)
			}
		})
	})

	t.Run("Task Commands", func(t *testing.T) {
		t.Run("list command - static", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--static"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("task list command failed: %v", err)
			}
		})

		t.Run("add command with valid args", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"add", "test task"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("task add command failed: %v", err)
			}
		})

		t.Run("projects command - static", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"projects", "--static"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("task projects command failed: %v", err)
			}
		})

		t.Run("tags command - static", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"tags", "--static"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("task tags command failed: %v", err)
			}
		})

		t.Run("contexts command - static", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"contexts", "--static"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("task contexts command failed: %v", err)
			}
		})

		t.Run("timesheet command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"timesheet"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("task timesheet command failed: %v", err)
			}
		})

		t.Run("view command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"view", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task view command to fail with non-existent ID")
			}
		})

		t.Run("update command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"update", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task update command to fail with non-existent ID")
			}
		})

		t.Run("start command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"start", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task start command to fail with non-existent ID")
			}
		})

		t.Run("stop command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"stop", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task stop command to fail with non-existent ID")
			}
		})

		t.Run("edit command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"edit", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task edit command to fail with non-existent ID")
			}
		})

		t.Run("delete command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"delete", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task delete command to fail with non-existent ID")
			}
		})

		t.Run("done command", func(t *testing.T) {
			handler, cleanup := createTestTaskHandler(t)
			defer cleanup()

			cmd := NewTaskCommand(handler).Create()
			cmd.SetArgs([]string{"done", "1"})
			err := cmd.Execute()
			if err == nil {
				t.Error("expected task done command to fail with non-existent ID")
			}
		})
	})
}
