package main

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
)

// SearchCommand implements [CommandGroup] for document search commands
type SearchCommand struct {
	handler *handlers.DocumentHandler
}

// NewSearchCommand creates a new search command group
func NewSearchCommand(handler *handlers.DocumentHandler) *SearchCommand {
	return &SearchCommand{handler: handler}
}

func (c *SearchCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "search",
		Short: "Search notes using TF-IDF",
		Long: `Full-text search for notes using Term Frequency-Inverse Document Frequency (TF-IDF) ranking.

The search engine tokenizes text, builds an inverted index, and ranks results by relevance.
Results are sorted by TF-IDF score, with higher scores indicating better matches.`,
	}

	queryCmd := &cobra.Command{
		Use:   "query [search terms...]",
		Short: "Search for documents matching query terms",
		Long: `Search for documents using TF-IDF ranking.

Examples:
  noteleaf search query go programming
  noteleaf search query "machine learning" --limit 5`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			limit, _ := cmd.Flags().GetInt("limit")

			return c.handler.Search(cmd.Context(), query, limit)
		},
	}
	queryCmd.Flags().IntP("limit", "l", 10, "Maximum number of results to return")
	root.AddCommand(queryCmd)

	rebuildCmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild search index from notes",
		Long: `Rebuild the search index from all notes in the database.

This command:
1. Clears the existing document index
2. Copies all notes to the documents table
3. Builds a new TF-IDF search index

Run this after adding, updating, or deleting notes to refresh the search index.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handler.RebuildIndex(cmd.Context())
		},
	}
	root.AddCommand(rebuildCmd)

	return root
}
