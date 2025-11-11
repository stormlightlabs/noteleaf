package main

import "github.com/spf13/cobra"

func addCommonTaskFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("priority", "p", "", "Set task priority")
	cmd.Flags().String("project", "", "Set task project")
	cmd.Flags().StringP("context", "c", "", "Set task context")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Add tags to task")
}

func addRecurrenceFlags(cmd *cobra.Command) {
	cmd.Flags().String("recur", "", "Set recurrence rule (e.g., FREQ=DAILY)")
	cmd.Flags().String("until", "", "Set recurrence end date (YYYY-MM-DD)")
}

func addDependencyFlags(cmd *cobra.Command) {
	cmd.Flags().String("depends-on", "", "Set task dependencies (comma-separated UUIDs)")
}

func addParentFlag(cmd *cobra.Command) {
	cmd.Flags().String("parent", "", "Set parent task UUID")
}

func addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().String("format", "detailed", "Output format (detailed, brief)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("no-metadata", false, "Hide creation/modification timestamps")
}

func addDueDateFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("due", "d", "", "Set due date (YYYY-MM-DD)")
}

func addWaitScheduledFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("wait", "w", "", "Task not actionable until date (YYYY-MM-DD)")
	cmd.Flags().StringP("scheduled", "s", "", "Task scheduled to start on date (YYYY-MM-DD)")
}
