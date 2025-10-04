package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) string {
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("command %q failed: %v", args, err)
	}
	return buf.String()
}

func TestNewApp(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		origDB := store.NewDatabase
		origConfig := store.LoadConfig
		defer func() {
			store.NewDatabase = origDB
			store.LoadConfig = origConfig
		}()

		store.NewDatabase = func() (*store.Database, error) { return &store.Database{}, nil }
		store.LoadConfig = func() (*store.Config, error) { return &store.Config{}, nil }

		app, err := NewApp()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if app.db == nil || app.config == nil {
			t.Fatalf("expected db and config to be initialized")
		}
	})

	t.Run("DBError", func(t *testing.T) {
		origDB := store.NewDatabase
		defer func() { store.NewDatabase = origDB }()
		store.NewDatabase = func() (*store.Database, error) { return nil, errors.New("db boom") }

		_, err := NewApp()
		if err == nil || !strings.Contains(err.Error(), "failed to initialize database") {
			t.Errorf("expected db init error, got %v", err)
		}
	})

	t.Run("ConfigError", func(t *testing.T) {
		origDB := store.NewDatabase
		origConfig := store.LoadConfig
		defer func() {
			store.NewDatabase = origDB
			store.LoadConfig = origConfig
		}()
		store.NewDatabase = func() (*store.Database, error) { return &store.Database{}, nil }
		store.LoadConfig = func() (*store.Config, error) { return nil, errors.New("config boom") }

		_, err := NewApp()
		if err == nil || !strings.Contains(err.Error(), "failed to load configuration") {
			t.Errorf("expected config load error, got %v", err)
		}
	})
}

func TestRootCmd(t *testing.T) {
	t.Run("Help", func(t *testing.T) {
		root := rootCmd()
		root.SetArgs([]string{})
		if err := root.Execute(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("PrintArgs", func(t *testing.T) {
		root := rootCmd()
		root.SetArgs([]string{"hello", "world"})
		buf := &bytes.Buffer{}
		root.SetOut(buf)

		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := buf.String(); !strings.Contains(got, "hello world") {
			t.Errorf("expected output to contain 'hello world', got %q", got)
		}
	})
}

func TestStatusCmd(t *testing.T) {
	output := executeCommand(t, statusCmd(), nil...)
	if output == "" {
		t.Errorf("expected some status output, got empty string")
	}
}

func TestResetCmd(t *testing.T) {
	_ = executeCommand(t, resetCmd(), nil...)
}

func TestSetupCmd(t *testing.T) {
	_ = executeCommand(t, setupCmd(), nil...)
}

func TestConfCmd(t *testing.T) {
	cmd := confCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"path"})

	if err := cmd.Execute(); err != nil {
		t.Errorf("config path command failed: %v", err)
	}
}

func TestRun(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		code := run()
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	t.Run("TaskHandlerError", func(t *testing.T) {
		orig := newTaskHandler
		defer func() { newTaskHandler = orig }()
		newTaskHandler = func() (*handlers.TaskHandler, error) { return nil, errors.New("boom") }

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("MovieHandlerError", func(t *testing.T) {
		orig := newMovieHandler
		defer func() { newMovieHandler = orig }()
		newMovieHandler = func() (*handlers.MovieHandler, error) { return nil, errors.New("boom") }

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("TVHandlerError", func(t *testing.T) {
		orig := newTVHandler
		defer func() { newTVHandler = orig }()
		newTVHandler = func() (*handlers.TVHandler, error) { return nil, errors.New("boom") }

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("NoteHandlerError", func(t *testing.T) {
		orig := newNoteHandler
		defer func() { newNoteHandler = orig }()
		newNoteHandler = func() (*handlers.NoteHandler, error) { return nil, errors.New("boom") }

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("BookHandlerError", func(t *testing.T) {
		orig := newBookHandler
		defer func() { newBookHandler = orig }()
		newBookHandler = func() (*handlers.BookHandler, error) { return nil, errors.New("boom") }

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("ArticleHandlerError", func(t *testing.T) {
		orig := newArticleHandler
		defer func() { newArticleHandler = orig }()
		newArticleHandler = func() (*handlers.ArticleHandler, error) { return nil, errors.New("boom") }

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})

	t.Run("FangExecuteError", func(t *testing.T) {
		orig := exc
		defer func() { exc = orig }()
		exc = func(ctx context.Context, cmd *cobra.Command, opts ...fang.Option) error {
			return errors.New("fang failed")
		}

		if code := run(); code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})
}
