//go:build prod

package main

import "github.com/spf13/cobra"

// registerTools is a no-op in production builds
func registerTools(*cobra.Command) {}
