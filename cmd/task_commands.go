package main

import (
	"github.com/spf13/cobra"
	"github.com/stormlightlabs/noteleaf/internal/handlers"
)

// TaskCommand implements CommandGroup for task-related commands
type TaskCommand struct {
	handler *handlers.TaskHandler
}

// NewTaskCommand creates a new TaskCommands with the given handler
func NewTaskCommand(handler *handlers.TaskHandler) *TaskCommand {
	return &TaskCommand{handler: handler}
}

func (c *TaskCommand) Create() *cobra.Command {
	root := &cobra.Command{Use: "todo", Aliases: []string{"task"}, Short: "task management"}

	for _, init := range []func(*handlers.TaskHandler) *cobra.Command{
		addTaskCmd, listTaskCmd, viewTaskCmd, updateTaskCmd, editTaskCmd,
		deleteTaskCmd, taskProjectsCmd, taskTagsCmd, taskContextsCmd,
		taskCompleteCmd, taskStartCmd, taskStopCmd, timesheetViewCmd,
	} {
		cmd := init(c.handler)
		root.AddCommand(cmd)
	}

	return root
}

func addTaskCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [description]",
		Short:   "Add a new task",
		Aliases: []string{"create", "new"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			priority, _ := c.Flags().GetString("priority")
			project, _ := c.Flags().GetString("project")
			context, _ := c.Flags().GetString("context")
			due, _ := c.Flags().GetString("due")
			tags, _ := c.Flags().GetStringSlice("tags")

			defer h.Close()
			return h.Create(c.Context(), args, priority, project, context, due, tags)
		},
	}
	cmd.Flags().StringP("priority", "p", "", "Set task priority")
	cmd.Flags().String("project", "", "Set task project")
	cmd.Flags().StringP("context", "c", "", "Set task context")
	cmd.Flags().StringP("due", "d", "", "Set due date (YYYY-MM-DD)")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Add tags to task")

	return cmd
}

func listTaskCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		Aliases: []string{"ls"},
		Long: `List tasks with optional filtering and display modes.

By default, shows tasks in an interactive TaskWarrior-like interface.
Use --static to show a simple text list instead.
Use --all to show all tasks, otherwise only pending tasks are shown.`,
		RunE: func(c *cobra.Command, args []string) error {
			static, _ := c.Flags().GetBool("static")
			showAll, _ := c.Flags().GetBool("all")
			status, _ := c.Flags().GetString("status")
			priority, _ := c.Flags().GetString("priority")
			project, _ := c.Flags().GetString("project")
			context, _ := c.Flags().GetString("context")

			defer h.Close()
			return h.List(c.Context(), static, showAll, status, priority, project, context)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "Force interactive mode (default)")
	cmd.Flags().Bool("static", false, "Use static text output instead of interactive")
	cmd.Flags().BoolP("all", "a", false, "Show all tasks (default: pending only)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("priority", "", "Filter by priority")
	cmd.Flags().String("project", "", "Filter by project")
	cmd.Flags().String("context", "", "Filter by context")

	return cmd
}

func viewTaskCmd(handler *handlers.TaskHandler) *cobra.Command {
	viewCmd := &cobra.Command{
		Use:   "view [task-id]",
		Short: "View task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			noMetadata, _ := cmd.Flags().GetBool("no-metadata")

			defer handler.Close()
			return handler.View(cmd.Context(), args, format, jsonOutput, noMetadata)
		},
	}
	viewCmd.Flags().String("format", "detailed", "Output format (detailed, brief)")
	viewCmd.Flags().Bool("json", false, "Output as JSON")
	viewCmd.Flags().Bool("no-metadata", false, "Hide creation/modification timestamps")

	return viewCmd
}

func updateTaskCmd(handler *handlers.TaskHandler) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update [task-id]",
		Short: "Update task properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			description, _ := cmd.Flags().GetString("description")
			status, _ := cmd.Flags().GetString("status")
			priority, _ := cmd.Flags().GetString("priority")
			project, _ := cmd.Flags().GetString("project")
			context, _ := cmd.Flags().GetString("context")
			due, _ := cmd.Flags().GetString("due")
			addTags, _ := cmd.Flags().GetStringSlice("add-tag")
			removeTags, _ := cmd.Flags().GetStringSlice("remove-tag")

			defer handler.Close()
			return handler.Update(cmd.Context(), taskID, description, status, priority, project, context, due, addTags, removeTags)
		},
	}
	updateCmd.Flags().String("description", "", "Update task description")
	updateCmd.Flags().String("status", "", "Update task status")
	updateCmd.Flags().StringP("priority", "p", "", "Update task priority")
	updateCmd.Flags().String("project", "", "Update task project")
	updateCmd.Flags().StringP("context", "c", "", "Update task context")
	updateCmd.Flags().StringP("due", "d", "", "Update due date (YYYY-MM-DD)")
	updateCmd.Flags().StringSlice("add-tag", []string{}, "Add tags to task")
	updateCmd.Flags().StringSlice("remove-tag", []string{}, "Remove tags from task")

	return updateCmd
}

func taskProjectsCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Short:   "List projects",
		Aliases: []string{"proj"},
		RunE: func(c *cobra.Command, args []string) error {
			static, _ := c.Flags().GetBool("static")
			todoTxt, _ := c.Flags().GetBool("todo-txt")

			defer h.Close()
			return h.ListProjects(c.Context(), static, todoTxt)
		},
	}
	cmd.Flags().Bool("static", false, "Use static text output instead of interactive")
	cmd.Flags().Bool("todo-txt", false, "Format output with +project prefix for todo.txt compatibility")

	return cmd
}

func taskTagsCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags",
		Short:   "List tags",
		Aliases: []string{"t"},
		RunE: func(c *cobra.Command, args []string) error {
			static, _ := c.Flags().GetBool("static")
			defer h.Close()
			return h.ListTags(c.Context(), static)
		},
	}
	cmd.Flags().Bool("static", false, "Use static text output instead of interactive")
	return cmd
}

func taskStartCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [task-id]",
		Short: "Start time tracking for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			taskID := args[0]
			description, _ := c.Flags().GetString("note")

			defer h.Close()
			return h.Start(c.Context(), taskID, description)
		},
	}
	cmd.Flags().StringP("note", "n", "", "Add a note to the time entry")
	return cmd
}

func taskStopCmd(h *handlers.TaskHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [task-id]",
		Short: "Stop time tracking for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			taskID := args[0]
			defer h.Close()
			return h.Stop(c.Context(), taskID)
		},
	}
}

func timesheetViewCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timesheet",
		Short: "Show time tracking summary",
		Long: `Show time tracking summary for tasks.

By default shows time entries for the last 7 days.
Use --task to show timesheet for a specific task.
Use --days to change the date range.`,
		RunE: func(c *cobra.Command, args []string) error {
			days, _ := c.Flags().GetInt("days")
			taskID, _ := c.Flags().GetString("task")

			defer h.Close()
			return h.Timesheet(c.Context(), days, taskID)
		},
	}
	cmd.Flags().IntP("days", "d", 7, "Number of days to show in timesheet")
	cmd.Flags().StringP("task", "t", "", "Show timesheet for specific task ID")
	return cmd
}

func editTaskCmd(h *handlers.TaskHandler) *cobra.Command {
	return &cobra.Command{
		Use:     "edit [task-id]",
		Short:   "Edit task interactively with status picker and priority toggle",
		Aliases: []string{"e"},
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			taskID := args[0]
			defer h.Close()
			return h.EditInteractive(c.Context(), taskID)
		},
	}
}

func deleteTaskCmd(h *handlers.TaskHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [task-id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.Delete(c.Context(), args)
		},
	}
}

func taskContextsCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contexts",
		Short:   "List contexts (locations)",
		Aliases: []string{"con", "loc", "ctx", "locations"},
		RunE: func(c *cobra.Command, args []string) error {
			static, _ := c.Flags().GetBool("static")
			todoTxt, _ := c.Flags().GetBool("todo-txt")

			defer h.Close()
			return h.ListContexts(c.Context(), static, todoTxt)
		},
	}
	cmd.Flags().Bool("static", false, "Use static text output instead of interactive")
	cmd.Flags().Bool("todo-txt", false, "Format output with @context prefix for todo.txt compatibility")
	return cmd
}

func taskCompleteCmd(h *handlers.TaskHandler) *cobra.Command {
	return &cobra.Command{
		Use:     "done [task-id]",
		Short:   "Mark task as completed",
		Aliases: []string{"complete"},
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.Done(c.Context(), args)
		},
	}
}
