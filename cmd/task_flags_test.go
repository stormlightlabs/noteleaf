package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTaskFlags(t *testing.T) {
	t.Run("AddCommonTaskFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		addCommonTaskFlags(cmd)

		if cmd.Flags().Lookup("priority") == nil {
			t.Error("Expected priority flag to be defined")
		}
		if cmd.Flags().Lookup("project") == nil {
			t.Error("Expected project flag to be defined")
		}
		if cmd.Flags().Lookup("context") == nil {
			t.Error("Expected context flag to be defined")
		}
		if cmd.Flags().Lookup("tags") == nil {
			t.Error("Expected tags flag to be defined")
		}

		if cmd.Flags().ShorthandLookup("p") == nil {
			t.Error("Expected 'p' shorthand for priority")
		}
		if cmd.Flags().ShorthandLookup("c") == nil {
			t.Error("Expected 'c' shorthand for context")
		}
		if cmd.Flags().ShorthandLookup("t") == nil {
			t.Error("Expected 't' shorthand for tags")
		}
	})

	t.Run("AddRecurrenceFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		addRecurrenceFlags(cmd)

		if cmd.Flags().Lookup("recur") == nil {
			t.Error("Expected recur flag to be defined")
		}
		if cmd.Flags().Lookup("until") == nil {
			t.Error("Expected until flag to be defined")
		}
	})

	t.Run("AddDependencyFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		addDependencyFlags(cmd)

		if cmd.Flags().Lookup("depends-on") == nil {
			t.Error("Expected depends-on flag to be defined")
		}
	})

	t.Run("AddParentFlag", func(t *testing.T) {
		cmd := &cobra.Command{}
		addParentFlag(cmd)

		if cmd.Flags().Lookup("parent") == nil {
			t.Error("Expected parent flag to be defined")
		}
	})

	t.Run("AddOutputFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		addOutputFlags(cmd)

		if cmd.Flags().Lookup("format") == nil {
			t.Error("Expected format flag to be defined")
		}
		if cmd.Flags().Lookup("json") == nil {
			t.Error("Expected json flag to be defined")
		}
		if cmd.Flags().Lookup("no-metadata") == nil {
			t.Error("Expected no-metadata flag to be defined")
		}

		format, _ := cmd.Flags().GetString("format")
		if format != "detailed" {
			t.Errorf("Expected format default to be 'detailed', got '%s'", format)
		}
	})

	t.Run("AddDueDateFlag", func(t *testing.T) {
		cmd := &cobra.Command{}
		addDueDateFlag(cmd)

		if cmd.Flags().Lookup("due") == nil {
			t.Error("Expected due flag to be defined")
		}

		if cmd.Flags().ShorthandLookup("d") == nil {
			t.Error("Expected 'd' shorthand for due")
		}
	})
}
