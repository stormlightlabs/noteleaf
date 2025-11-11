package main

import (
	"strings"

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
	root := &cobra.Command{
		Use:     "todo",
		Aliases: []string{"task"},
		Short:   "task management",
		Long: `Manage tasks with TaskWarrior-inspired features.

Track todos with priorities, projects, contexts, and tags. Supports hierarchical
tasks with parent/child relationships, task dependencies, recurring tasks, and
time tracking. Tasks can be filtered by status, priority, project, or context.`,
	}

	root.AddGroup(
		&cobra.Group{ID: "task-ops", Title: "Basic Operations"},
		&cobra.Group{ID: "task-meta", Title: "Metadata"},
		&cobra.Group{ID: "task-tracking", Title: "Tracking"},
		&cobra.Group{ID: "task-reports", Title: "Reports & Views"},
	)

	for _, init := range []func(*handlers.TaskHandler) *cobra.Command{
		addTaskCmd, listTaskCmd, viewTaskCmd, updateTaskCmd, editTaskCmd, deleteTaskCmd,
	} {
		cmd := init(c.handler)
		cmd.GroupID = "task-ops"
		root.AddCommand(cmd)
	}

	for _, init := range []func(*handlers.TaskHandler) *cobra.Command{
		taskProjectsCmd, taskTagsCmd, taskContextsCmd,
	} {
		cmd := init(c.handler)
		cmd.GroupID = "task-meta"
		root.AddCommand(cmd)
	}

	for _, init := range []func(*handlers.TaskHandler) *cobra.Command{
		timesheetViewCmd, taskStartCmd, taskStopCmd, taskCompleteCmd, taskRecurCmd, taskDependCmd,
	} {
		cmd := init(c.handler)
		cmd.GroupID = "task-tracking"
		root.AddCommand(cmd)
	}

	for _, init := range []func(*handlers.TaskHandler) *cobra.Command{
		nextActionsCmd, reportCompletedCmd, reportWaitingCmd, reportBlockedCmd, calendarCmd,
	} {
		cmd := init(c.handler)
		cmd.GroupID = "task-reports"
		root.AddCommand(cmd)
	}

	return root
}

func addTaskCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [description]",
		Short:   "Add a new task",
		Aliases: []string{"create", "new"},
		Long: `Create a new task with description and optional attributes.

Tasks can be created with priority levels (low, medium, high, urgent), assigned
to projects and contexts, tagged for organization, and configured with due dates
and recurrence rules. Dependencies can be established to ensure tasks are
completed in order.

Examples:
  noteleaf todo add "Write documentation" --priority high --project docs
  noteleaf todo add "Weekly review" --recur "FREQ=WEEKLY" --due 2024-01-15`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			description := strings.Join(args, " ")
			priority, _ := c.Flags().GetString("priority")
			project, _ := c.Flags().GetString("project")
			context, _ := c.Flags().GetString("context")
			due, _ := c.Flags().GetString("due")
			wait, _ := c.Flags().GetString("wait")
			scheduled, _ := c.Flags().GetString("scheduled")
			recur, _ := c.Flags().GetString("recur")
			until, _ := c.Flags().GetString("until")
			parent, _ := c.Flags().GetString("parent")
			dependsOn, _ := c.Flags().GetString("depends-on")
			tags, _ := c.Flags().GetStringSlice("tags")

			defer h.Close()
			// TODO: Make a CreateTask struct
			return h.Create(c.Context(), description, priority, project, context, due, wait, scheduled, recur, until, parent, dependsOn, tags)
		},
	}
	addCommonTaskFlags(cmd)
	addDueDateFlag(cmd)
	addWaitScheduledFlags(cmd)
	addRecurrenceFlags(cmd)
	addParentFlag(cmd)
	addDependencyFlags(cmd)

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
			sortBy, _ := c.Flags().GetString("sort")

			defer h.Close()
			// TODO: TaskFilter struct
			return h.List(c.Context(), static, showAll, status, priority, project, context, sortBy)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "Force interactive mode (default)")
	cmd.Flags().Bool("static", false, "Use static text output instead of interactive")
	cmd.Flags().BoolP("all", "a", false, "Show all tasks (default: pending only)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("priority", "", "Filter by priority")
	cmd.Flags().String("project", "", "Filter by project")
	cmd.Flags().String("context", "", "Filter by context")
	cmd.Flags().String("sort", "", "Sort by (urgency)")

	return cmd
}

func viewTaskCmd(handler *handlers.TaskHandler) *cobra.Command {
	viewCmd := &cobra.Command{
		Use:   "view [task-id]",
		Short: "View task by ID",
		Long: `Display detailed information for a specific task.

Shows all task attributes including description, status, priority, project,
context, tags, due date, creation time, and modification history. Use --json
for machine-readable output or --no-metadata to show only the description.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			noMetadata, _ := cmd.Flags().GetBool("no-metadata")

			defer handler.Close()
			return handler.View(cmd.Context(), args, format, jsonOutput, noMetadata)
		},
	}
	addOutputFlags(viewCmd)

	return viewCmd
}

func updateTaskCmd(handler *handlers.TaskHandler) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update [task-id]",
		Short: "Update task properties",
		Long: `Modify attributes of an existing task.

Update any task property including description, status, priority, project,
context, due date, recurrence rule, or parent task. Add or remove tags and
dependencies. Multiple attributes can be updated in a single command.

Examples:
  noteleaf todo update 123 --priority urgent --due tomorrow
  noteleaf todo update 456 --add-tag urgent --project website`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			description, _ := cmd.Flags().GetString("description")
			status, _ := cmd.Flags().GetString("status")
			priority, _ := cmd.Flags().GetString("priority")
			project, _ := cmd.Flags().GetString("project")
			context, _ := cmd.Flags().GetString("context")
			due, _ := cmd.Flags().GetString("due")
			recur, _ := cmd.Flags().GetString("recur")
			until, _ := cmd.Flags().GetString("until")
			parent, _ := cmd.Flags().GetString("parent")
			addTags, _ := cmd.Flags().GetStringSlice("add-tag")
			removeTags, _ := cmd.Flags().GetStringSlice("remove-tag")
			addDeps, _ := cmd.Flags().GetString("add-depends")
			removeDeps, _ := cmd.Flags().GetString("remove-depends")

			defer handler.Close()
			return handler.Update(cmd.Context(), taskID, description, status, priority, project, context, due, recur, until, parent, addTags, removeTags, addDeps, removeDeps)
		},
	}
	updateCmd.Flags().String("description", "", "Update task description")
	updateCmd.Flags().String("status", "", "Update task status")
	addCommonTaskFlags(updateCmd)
	addDueDateFlag(updateCmd)
	addRecurrenceFlags(updateCmd)
	addParentFlag(updateCmd)
	updateCmd.Flags().StringSlice("add-tag", []string{}, "Add tags to task")
	updateCmd.Flags().StringSlice("remove-tag", []string{}, "Remove tags from task")
	updateCmd.Flags().String("add-depends", "", "Add task dependencies (comma-separated UUIDs)")
	updateCmd.Flags().String("remove-depends", "", "Remove task dependencies (comma-separated UUIDs)")

	return updateCmd
}

func taskProjectsCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Short:   "List projects",
		Aliases: []string{"proj"},
		Long: `Display all projects with task counts.

Shows each project used in your tasks along with the number of tasks in each
project. Use --todo-txt to format output with +project syntax for compatibility
with todo.txt tools.`,
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
		Long: `Display all tags used across tasks.

Shows each tag with the number of tasks using it. Tags provide flexible
categorization orthogonal to projects and contexts.`,
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
		Long: `Begin tracking time spent on a task.

Records the start time for a work session. Only one task can be actively
tracked at a time. Use --note to add a description of what you're working on.`,
		Args: cobra.ExactArgs(1),
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
		Long: `End time tracking for the active task.

Records the end time and calculates duration for the current work session.
Duration is added to the task's total time tracked.`,
		Args: cobra.ExactArgs(1),
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
		Long: `Open interactive editor for task modification.

Provides a user-friendly interface with status picker and priority toggle.
Easier than using multiple command-line flags for complex updates.`,
		Args: cobra.ExactArgs(1),
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
		Long: `Permanently remove a task from the database.

This operation cannot be undone. Consider updating the task status to
'deleted' instead if you want to preserve the record for historical purposes.`,
		Args: cobra.ExactArgs(1),
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
		Long: `Display all contexts with task counts.

Contexts represent locations or environments where tasks can be completed (e.g.,
@home, @office, @errands). Use --todo-txt to format output with @context syntax
for compatibility with todo.txt tools.`,
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
		Long: `Mark a task as completed with current timestamp.

Sets the task status to 'completed' and records the completion time. For
recurring tasks, generates the next instance based on the recurrence rule.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.Done(c.Context(), args)
		},
	}
}

func taskRecurCmd(h *handlers.TaskHandler) *cobra.Command {
	root := &cobra.Command{
		Use:     "recur",
		Short:   "Manage task recurrence",
		Aliases: []string{"repeat"},
		Long: `Configure recurring task patterns.

Create tasks that repeat on a schedule using iCalendar recurrence rules (RRULE).
Supports daily, weekly, monthly, and yearly patterns with optional end dates.`,
	}

	setCmd := &cobra.Command{
		Use:   "set [task-id]",
		Short: "Set recurrence rule for a task",
		Long: `Apply a recurrence rule to create repeating task instances.

Uses iCalendar RRULE syntax (e.g., "FREQ=DAILY" for daily tasks, "FREQ=WEEKLY;BYDAY=MO,WE,FR"
for specific weekdays). When a recurring task is completed, the next instance is
automatically generated.

Examples:
  noteleaf todo recur set 123 --rule "FREQ=DAILY"
  noteleaf todo recur set 456 --rule "FREQ=WEEKLY;BYDAY=MO" --until 2024-12-31`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			rule, _ := c.Flags().GetString("rule")
			until, _ := c.Flags().GetString("until")
			defer h.Close()
			return h.SetRecur(c.Context(), args[0], rule, until)
		},
	}
	setCmd.Flags().String("rule", "", "Recurrence rule (e.g., FREQ=DAILY)")
	setCmd.Flags().String("until", "", "Recurrence end date (YYYY-MM-DD)")

	clearCmd := &cobra.Command{
		Use:   "clear [task-id]",
		Short: "Clear recurrence rule from a task",
		Long: `Remove recurrence from a task.

Converts a recurring task to a one-time task. Existing future instances are not
affected.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.ClearRecur(c.Context(), args[0])
		},
	}

	showCmd := &cobra.Command{
		Use:   "show [task-id]",
		Short: "Show recurrence details for a task",
		Long: `Display recurrence rule and schedule information.

Shows the RRULE pattern, next occurrence date, and recurrence end date if
configured.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.ShowRecur(c.Context(), args[0])
		},
	}

	root.AddCommand(setCmd, clearCmd, showCmd)
	return root
}

func nextActionsCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "next",
		Short:   "Show next actions (actionable tasks sorted by urgency)",
		Aliases: []string{"na"},
		Long: `Display actionable tasks sorted by urgency score.

Shows tasks that can be worked on now (not waiting, not blocked, not completed),
ordered by their computed urgency based on priority, due date, age, and other factors.`,
		RunE: func(c *cobra.Command, args []string) error {
			limit, _ := c.Flags().GetInt("limit")
			defer h.Close()
			return h.NextActions(c.Context(), limit)
		},
	}
	cmd.Flags().IntP("limit", "n", 10, "Limit number of tasks shown")
	return cmd
}

func reportCompletedCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completed",
		Short: "Show completed tasks",
		Long:  "Display tasks that have been completed, sorted by completion date.",
		RunE: func(c *cobra.Command, args []string) error {
			limit, _ := c.Flags().GetInt("limit")
			defer h.Close()
			return h.ReportCompleted(c.Context(), limit)
		},
	}
	cmd.Flags().IntP("limit", "n", 20, "Limit number of tasks shown")
	return cmd
}

func reportWaitingCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "waiting",
		Short: "Show waiting tasks",
		Long:  "Display tasks that are waiting for a specific date before becoming actionable.",
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.ReportWaiting(c.Context())
		},
	}
	return cmd
}

func reportBlockedCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocked",
		Short: "Show blocked tasks",
		Long:  "Display tasks that are blocked by dependencies on other tasks.",
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.ReportBlocked(c.Context())
		},
	}
	return cmd
}

func calendarCmd(h *handlers.TaskHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calendar",
		Short:   "Show tasks in calendar view",
		Aliases: []string{"cal"},
		Long: `Display tasks with due dates in a calendar format.

Shows tasks organized by week and day, making it easy to see upcoming deadlines
and plan your work schedule.`,
		RunE: func(c *cobra.Command, args []string) error {
			weeks, _ := c.Flags().GetInt("weeks")
			defer h.Close()
			return h.Calendar(c.Context(), weeks)
		},
	}
	cmd.Flags().IntP("weeks", "w", 4, "Number of weeks to show")
	return cmd
}

func taskDependCmd(h *handlers.TaskHandler) *cobra.Command {
	root := &cobra.Command{
		Use:     "depend",
		Short:   "Manage task dependencies",
		Aliases: []string{"dep", "deps"},
		Long: `Create and manage task dependencies.

Establish relationships where one task must be completed before another can
begin. Useful for multi-step workflows and project management.`,
	}

	addCmd := &cobra.Command{
		Use:   "add [task-id] [depends-on-uuid]",
		Short: "Add a dependency to a task",
		Long: `Make a task dependent on another task's completion.

The first task cannot be started until the second task is completed. Use task
UUIDs to specify dependencies.`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.AddDep(c.Context(), args[0], args[1])
		},
	}

	removeCmd := &cobra.Command{
		Use:     "remove [task-id] [depends-on-uuid]",
		Short:   "Remove a dependency from a task",
		Aliases: []string{"rm"},
		Long:    "Delete a dependency relationship between two tasks.",
		Args:    cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.RemoveDep(c.Context(), args[0], args[1])
		},
	}

	listCmd := &cobra.Command{
		Use:     "list [task-id]",
		Short:   "List dependencies for a task",
		Aliases: []string{"ls"},
		Long:    "Show all tasks that must be completed before this task can be started.",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.ListDeps(c.Context(), args[0])
		},
	}

	blockedByCmd := &cobra.Command{
		Use:   "blocked-by [task-id]",
		Short: "Show tasks blocked by this task",
		Long:  "Display all tasks that depend on this task's completion.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			defer h.Close()
			return h.BlockedByDep(c.Context(), args[0])
		},
	}

	root.AddCommand(addCmd, removeCmd, listCmd, blockedByCmd)
	return root
}
