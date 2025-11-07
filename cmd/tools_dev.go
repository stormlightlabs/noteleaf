//go:build !prod

package main

import (
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/tools"
)

// registerTools adds development tools to the root command
func registerTools(root *cobra.Command) {
	root.AddCommand(tools.NewToolsCommand(root))
}
