package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/handlers"
)

func createTestPublicationHandler(t *testing.T) (*handlers.PublicationHandler, func()) {
	cleanup := setupCommandTest(t)
	handler, err := handlers.NewPublicationHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create test publication handler: %v", err)
	}
	return handler, func() {
		handler.Close()
		cleanup()
	}
}

func TestPublicationCommand(t *testing.T) {
	t.Run("CommandGroup Interface", func(t *testing.T) {
		handler, cleanup := createTestPublicationHandler(t)
		defer cleanup()

		var _ CommandGroup = NewPublicationCommand(handler)
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("creates command with correct structure", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()

			if cmd == nil {
				t.Fatal("Create returned nil")
			}
			if cmd.Use != "pub" {
				t.Errorf("Expected Use to be 'pub', got '%s'", cmd.Use)
			}
			if cmd.Short != "Manage leaflet publication sync" {
				t.Errorf("Expected Short to be 'Manage leaflet publication sync', got '%s'", cmd.Short)
			}
			if !cmd.HasSubCommands() {
				t.Error("Expected command to have subcommands")
			}
		})

		t.Run("has all expected subcommands", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			subcommands := cmd.Commands()
			subcommandNames := make([]string, len(subcommands))
			for i, subcmd := range subcommands {
				subcommandNames[i] = subcmd.Use
			}

			expectedSubcommands := []string{
				"auth [handle]",
				"pull",
				"list [--published|--draft|--all]",
				"status",
			}

			for _, expected := range expectedSubcommands {
				if !findSubcommand(subcommandNames, expected) {
					t.Errorf("Expected subcommand '%s' not found in %v", expected, subcommandNames)
				}
			}
		})
	})

	t.Run("Status Command", func(t *testing.T) {
		t.Run("shows not authenticated initially", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"status"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("status command failed: %v", err)
			}
		})
	})

	t.Run("List Command", func(t *testing.T) {
		t.Run("default filter", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"list"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("list command failed: %v", err)
			}
		})

		t.Run("with published flag", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--published"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("list --published failed: %v", err)
			}
		})

		t.Run("with draft flag", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--draft"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("list --draft failed: %v", err)
			}
		})

		t.Run("with all flag", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--all"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("list --all failed: %v", err)
			}
		})

		t.Run("published takes precedence over draft", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"list", "--published", "--draft"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("list with multiple flags failed: %v", err)
			}
		})
	})

	t.Run("Pull Command", func(t *testing.T) {
		t.Run("fails when not authenticated", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"pull"})
			err := cmd.Execute()

			if err == nil {
				t.Error("Expected pull to fail when not authenticated")
			}
			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got: %v", err)
			}
		})
	})

	t.Run("Command Help", func(t *testing.T) {
		t.Run("root help", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"help"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("help command failed: %v", err)
			}
		})

		t.Run("auth help", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"auth", "--help"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("auth help failed: %v", err)
			}
		})
	})

	t.Run("Command Aliases", func(t *testing.T) {
		t.Run("list alias ls works", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			cmd := NewPublicationCommand(handler).Create()
			cmd.SetArgs([]string{"ls"})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("list alias 'ls' failed: %v", err)
			}
		})
	})

	t.Run("Handler Validation", func(t *testing.T) {
		t.Run("auth validates empty handle", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			ctx := context.Background()
			err := handler.Auth(ctx, "", "password")

			if err == nil {
				t.Error("Expected error for empty handle")
			}
			if !strings.Contains(err.Error(), "handle is required") {
				t.Errorf("Expected 'handle is required' error, got: %v", err)
			}
		})

		t.Run("auth validates empty password", func(t *testing.T) {
			handler, cleanup := createTestPublicationHandler(t)
			defer cleanup()

			ctx := context.Background()
			err := handler.Auth(ctx, "test.bsky.social", "")

			if err == nil {
				t.Error("Expected error for empty password")
			}
			if !strings.Contains(err.Error(), "password is required") {
				t.Errorf("Expected 'password is required' error, got: %v", err)
			}
		})
	})
}
