//go:build !prod

package tools

import (
	"context"

	"github.com/spf13/cobra"
)

// NewLexiconsCommand creates a command for fetching Leaflet lexicons
func NewLexiconsCommand() *cobra.Command {
	var sha string
	var output string

	cmd := &cobra.Command{
		Use:   "lexicons",
		Short: "Fetch Leaflet lexicons from GitHub",
		Long: `Fetches Leaflet lexicons from the hyperlink-academy/leaflet repository.

This is a convenience wrapper around gh-repo with pre-configured defaults
for the Leaflet lexicon repository.`,
		Example: `  # Fetch latest lexicons
  noteleaf tools fetch lexicons

  # Fetch from a specific commit
  noteleaf tools fetch lexicons --sha abc123def

  # Fetch to a custom directory
  noteleaf tools fetch lexicons --output ./tmp/lexicons`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := ArchiveConfig{
				Repo:       "hyperlink-academy/leaflet",
				Path:       "lexicons/pub/leaflet/",
				Output:     output,
				SHA:        sha,
				FormatJSON: true,
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			return fetchAndExtractArchive(ctx, config, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&sha, "sha", "", "Specific commit SHA (default: latest)")
	cmd.Flags().StringVar(&output, "output", "lexdocs/leaflet/", "Output directory for lexicons")
	return cmd
}
