// TODO: implement prompt for password
//
//	See: https://github.com/charmbracelet/bubbletea/blob/main/examples/textinputs/main.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
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

The password will be prompted securely if not provided via flag.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var handle string
			if len(args) > 0 {
				handle = args[0]
			}

			password, _ := cmd.Flags().GetString("password")

			if handle == "" {
				return fmt.Errorf("handle is required")
			}

			if password == "" {
				return fmt.Errorf("password is required (use --password flag)")
			}

			defer c.handler.Close()
			return c.handler.Auth(cmd.Context(), handle, password)
		},
	}
	authCmd.Flags().StringP("password", "p", "", "App password (will be prompted if not provided)")
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

	// List command
	listCmd := &cobra.Command{
		Use:     "list [--published|--draft|--all]",
		Short:   "List notes synced with leaflet",
		Aliases: []string{"ls"},
		Long: `Display notes that have been pulled from or pushed to leaflet.

Shows publication metadata including:
- Publication status (draft vs published)
- Published date
- Leaflet record key (rkey)
- Content identifier (cid) for change tracking

Use filters to show specific subsets:
  --published  Show only published documents
  --draft      Show only drafts
  --all        Show all leaflet documents (default)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			published, _ := cmd.Flags().GetBool("published")
			draft, _ := cmd.Flags().GetBool("draft")
			all, _ := cmd.Flags().GetBool("all")

			filter := "all"
			if published {
				filter = "published"
			} else if draft {
				filter = "draft"
			} else if all {
				filter = "all"
			}

			defer c.handler.Close()
			return c.handler.List(cmd.Context(), filter)
		},
	}
	listCmd.Flags().Bool("published", false, "Show only published documents")
	listCmd.Flags().Bool("draft", false, "Show only drafts")
	listCmd.Flags().Bool("all", false, "Show all leaflet documents")
	root.AddCommand(listCmd)

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

	return root
}
