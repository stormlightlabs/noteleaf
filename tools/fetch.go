//go:build !prod

package tools

import "github.com/spf13/cobra"

// NewFetchCommand creates a parent command for fetching remote resources
func NewFetchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch remote resources",
		Long: `Fetch and synchronize remote resources from GitHub repositories.

Includes commands for fetching lexicons, schemas, and other data files.`,
	}

	cmd.AddCommand(NewGHRepoCommand())
	cmd.AddCommand(NewLexiconsCommand())

	return cmd
}
