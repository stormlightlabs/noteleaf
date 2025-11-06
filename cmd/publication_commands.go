package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

// PublicationCommand implements [CommandGroup] for leaflet publication commands
type PublicationCommand struct {
	handler *handlers.PublicationHandler
}

// NewPublicationCommand creates a new [PublicationCommand] with the given handler
func NewPublicationCommand(handler *handlers.PublicationHandler) *PublicationCommand {
	return &PublicationCommand{handler: handler}
}

func (c *PublicationCommand) Create() *cobra.Command {
	root := &cobra.Command{
		Use:   "pub",
		Short: "Manage leaflet publication sync",
		Long: `Sync notes with leaflet.pub (AT Protocol publishing platform).

Authenticate with your BlueSky account to pull drafts and published documents
from leaflet.pub into your local notes. Track publication status and manage
your writing workflow across platforms.

Authentication uses AT Protocol (the same system as BlueSky). You'll need:
- BlueSky handle (e.g., username.bsky.social)
- App password (generated at bsky.app/settings/app-passwords)

Getting Started:
  1. Authenticate: noteleaf pub auth <handle>
  2. Pull documents: noteleaf pub pull
  3. List publications: noteleaf pub list`,
	}

	authCmd := &cobra.Command{
		Use:   "auth [handle]",
		Short: "Authenticate with BlueSky/leaflet",
		Long: `Authenticate with AT Protocol (BlueSky) for leaflet access.

Your handle is typically: username.bsky.social

For the password, use an app password (not your main password):
1. Go to bsky.app/settings/app-passwords
2. Create a new app password named "noteleaf"
3. Use that password here

If credentials are not provided via flags, use the interactive input.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var handle string
			if len(args) > 0 {
				handle = args[0]
			}

			password, _ := cmd.Flags().GetString("password")

			if handle != "" && password != "" {
				defer c.handler.Close()
				return c.handler.Auth(cmd.Context(), handle, password)
			}

			form := ui.NewAuthForm(handle, ui.AuthFormOptions{})
			result, err := form.Run()
			if err != nil {
				return fmt.Errorf("failed to display auth form: %w", err)
			}

			if result.Canceled {
				return fmt.Errorf("authentication canceled")
			}

			defer c.handler.Close()
			return c.handler.Auth(cmd.Context(), result.Handle, result.Password)
		},
	}
	authCmd.Flags().StringP("password", "p", "", "App password (will prompt if not provided)")
	root.AddCommand(authCmd)

	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull documents from leaflet",
		Long: `Fetch all drafts and published documents from leaflet.pub.

This will:
- Connect to your BlueSky/leaflet account
- Fetch all documents in your repository
- Create new notes for documents not yet synced
- Update existing notes that have changed

Notes are matched by their leaflet record key (rkey) stored in the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer c.handler.Close()
			return c.handler.Pull(cmd.Context())
		},
	}
	root.AddCommand(pullCmd)

	listCmd := &cobra.Command{
		Use:     "list [--published|--draft|--all] [--interactive]",
		Short:   "List notes synced with leaflet",
		Aliases: []string{"ls"},
		Long: `Display notes that have been pulled from or pushed to leaflet.

Shows publication metadata including:
- Publication status (draft vs published)
- Published date
- Leaflet record key (rkey)
- Content identifier (cid) for change tracking

Use filters to show specific subsets:
  --published    Show only published documents
  --draft        Show only drafts
  --all          Show all leaflet documents (default)
  --interactive  Open interactive TUI browser with search and preview`,
		RunE: func(cmd *cobra.Command, args []string) error {
			published, _ := cmd.Flags().GetBool("published")
			draft, _ := cmd.Flags().GetBool("draft")
			all, _ := cmd.Flags().GetBool("all")
			interactive, _ := cmd.Flags().GetBool("interactive")

			filter := "all"
			if published {
				filter = "published"
			} else if draft {
				filter = "draft"
			} else if all {
				filter = "all"
			}

			defer c.handler.Close()

			if interactive {
				return c.handler.Browse(cmd.Context(), filter)
			}

			return c.handler.List(cmd.Context(), filter)
		},
	}
	listCmd.Flags().Bool("published", false, "Show only published documents")
	listCmd.Flags().Bool("draft", false, "Show only drafts")
	listCmd.Flags().Bool("all", false, "Show all leaflet documents")
	listCmd.Flags().BoolP("interactive", "i", false, "Open interactive TUI browser")
	root.AddCommand(listCmd)

	readCmd := &cobra.Command{
		Use:   "read [identifier]",
		Short: "Read a publication",
		Long: `Display a publication's content with formatted markdown rendering.

The identifier can be:
- Omitted: Display the newest publication
- Database ID: Display publication by note ID (e.g., 42)
- AT Protocol rkey: Display publication by leaflet rkey

Examples:
  noteleaf pub read                  # Show newest publication
  noteleaf pub read 123              # Show publication with note ID 123
  noteleaf pub read 3jxx...          # Show publication by rkey`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier := ""
			if len(args) > 0 {
				identifier = args[0]
			}

			defer c.handler.Close()
			return c.handler.Read(cmd.Context(), identifier)
		},
	}
	root.AddCommand(readCmd)

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show leaflet authentication status",
		Long:  "Display current authentication status and session information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer c.handler.Close()
			status := c.handler.GetAuthStatus()
			fmt.Println("Leaflet Status:")
			fmt.Printf("  %s\n", status)
			return nil
		},
	}
	root.AddCommand(statusCmd)

	postCmd := &cobra.Command{
		Use:   "post [note-id]",
		Short: "Create a new document on leaflet",
		Long: `Publish a local note to leaflet.pub as a new document.

This command converts your markdown note to leaflet's block format and creates
a new document on the platform. The note will be linked to the leaflet document
for future updates via the patch command.

Examples:
  noteleaf pub post 123                # Publish note 123
  noteleaf pub post 123 --draft        # Create as draft
  noteleaf pub post 123 --preview      # Preview without posting
  noteleaf pub post 123 --validate     # Validate conversion only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID, err := parseNoteID(args[0])
			if err != nil {
				return err
			}

			isDraft, _ := cmd.Flags().GetBool("draft")
			preview, _ := cmd.Flags().GetBool("preview")
			validate, _ := cmd.Flags().GetBool("validate")

			defer c.handler.Close()

			if preview {
				return c.handler.PostPreview(cmd.Context(), noteID, isDraft)
			}

			if validate {
				return c.handler.PostValidate(cmd.Context(), noteID, isDraft)
			}

			return c.handler.Post(cmd.Context(), noteID, isDraft)
		},
	}
	postCmd.Flags().Bool("draft", false, "Create as draft instead of publishing")
	postCmd.Flags().Bool("preview", false, "Show what would be posted without actually posting")
	postCmd.Flags().Bool("validate", false, "Validate markdown conversion without posting")
	root.AddCommand(postCmd)

	patchCmd := &cobra.Command{
		Use:   "patch [note-id]",
		Short: "Update an existing document on leaflet",
		Long: `Update an existing leaflet document from a local note.

This command converts your markdown note to leaflet's block format and updates
the existing document on the platform. The note must have been previously posted
or pulled from leaflet (it needs a leaflet record key).

The document's draft/published status is preserved from the note's current state.

Examples:
  noteleaf pub patch 123            # Update existing document
  noteleaf pub patch 123 --preview  # Preview without updating
  noteleaf pub patch 123 --validate # Validate conversion only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID, err := parseNoteID(args[0])
			if err != nil {
				return err
			}

			preview, _ := cmd.Flags().GetBool("preview")
			validate, _ := cmd.Flags().GetBool("validate")

			defer c.handler.Close()

			if preview {
				return c.handler.PatchPreview(cmd.Context(), noteID)
			}

			if validate {
				return c.handler.PatchValidate(cmd.Context(), noteID)
			}

			return c.handler.Patch(cmd.Context(), noteID)
		},
	}
	patchCmd.Flags().Bool("preview", false, "Show what would be updated without actually patching")
	patchCmd.Flags().Bool("validate", false, "Validate markdown conversion without patching")
	root.AddCommand(patchCmd)

	pushCmd := &cobra.Command{
		Use:   "push [note-ids...]",
		Short: "Create or update multiple documents on leaflet",
		Long: `Batch publish or update multiple local notes to leaflet.pub.

For each note:
- If the note has never been published, creates a new document (like post)
- If the note has been published before, updates the existing document (like patch)

This is useful for bulk operations and continuous publishing workflows.

Examples:
  noteleaf pub push 1 2 3              # Publish/update notes 1, 2, and 3
  noteleaf pub push 42 99 --draft      # Create/update as drafts`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteIDs := make([]int64, len(args))
			for i, arg := range args {
				id, err := parseNoteID(arg)
				if err != nil {
					return err
				}
				noteIDs[i] = id
			}

			isDraft, _ := cmd.Flags().GetBool("draft")

			defer c.handler.Close()
			return c.handler.Push(cmd.Context(), noteIDs, isDraft)
		},
	}
	pushCmd.Flags().Bool("draft", false, "Create/update as drafts instead of publishing")
	root.AddCommand(pushCmd)

	return root
}

func parseNoteID(arg string) (int64, error) {
	noteID, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid note ID '%s': must be a number", arg)
	}
	return noteID, nil
}
