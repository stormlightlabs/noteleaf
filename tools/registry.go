//go:build !prod

package tools

import "github.com/spf13/cobra"

// NewToolsCommand creates a parent command for all development tools
func NewToolsCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Development and maintenance tools",
		Long: `Development tools for documentation generation, data synchronization,
and maintenance tasks. These commands are only available in dev builds.`,
	}
	cmd.AddCommand(NewDocGenCommand(root))
	cmd.AddCommand(NewFetchCommand())

	return cmd
}
