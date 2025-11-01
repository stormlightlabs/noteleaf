// TODO: add context field to table in [TaskHandler.listTasksInteractive]
package handlers

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

// TaskHandler handles all task-related commands
type TaskHandler struct {
	db     *store.Database
	config *store.Config
	repos  *repo.Repositories
}

// NewTaskHandler creates a new task handler
func NewTaskHandler() (*TaskHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)

	return &TaskHandler{
		db:     db,
		config: config,
		repos:  repos,
	}, nil
}

// Close cleans up resources
func (h *TaskHandler) Close() error {
	return h.db.Close()
}

// Create creates a new task
func (h *TaskHandler) Create(ctx context.Context, description, priority, project, context, due, recur, until, parentUUID, dependsOn string, tags []string) error {
	if description == "" {
		return fmt.Errorf("task description required")
	}

	parsed := parseDescription(description)

	if project != "" {
		parsed.Project = project
	}
	if context != "" {
		parsed.Context = context
	}
	if due != "" {
		parsed.Due = due
	}
	if recur != "" {
		parsed.Recur = recur
	}
	if until != "" {
		parsed.Until = until
	}
	if parentUUID != "" {
		parsed.ParentUUID = parentUUID
	}
	if dependsOn != "" {
		parsed.DependsOn = strings.Split(dependsOn, ",")
	}
	if len(tags) > 0 {
		parsed.Tags = append(parsed.Tags, tags...)
	}

	task := &models.Task{
		UUID:        uuid.New().String(),
		Description: parsed.Description,
		Status:      "pending",
		Priority:    priority,
		Project:     parsed.Project,
		Context:     parsed.Context,
		Tags:        parsed.Tags,
		Recur:       models.RRule(parsed.Recur),
		DependsOn:   parsed.DependsOn,
	}

	if parsed.Due != "" {
		if dueTime, err := time.Parse("2006-01-02", parsed.Due); err == nil {
			task.Due = &dueTime
		} else {
			return fmt.Errorf("invalid due date format, use YYYY-MM-DD: %w", err)
		}
	}

	if parsed.Until != "" {
		if untilTime, err := time.Parse("2006-01-02", parsed.Until); err == nil {
			task.Until = &untilTime
		} else {
			return fmt.Errorf("invalid until date format, use YYYY-MM-DD: %w", err)
		}
	}

	if parsed.ParentUUID != "" {
		task.ParentUUID = &parsed.ParentUUID
	}

	id, err := h.repos.Tasks.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Task created (ID: %d, UUID: %s): %s\n", id, task.UUID, task.Description)

	if priority != "" {
		fmt.Printf("Priority: %s\n", priority)
	}
	if task.Project != "" {
		fmt.Printf("Project: %s\n", task.Project)
	}
	if task.Context != "" {
		fmt.Printf("Context: %s\n", task.Context)
	}
	if len(task.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(task.Tags, ", "))
	}
	if task.Due != nil {
		fmt.Printf("Due: %s\n", task.Due.Format("2006-01-02"))
	}
	if task.Recur != "" {
		fmt.Printf("Recur: %s\n", task.Recur)
	}
	if task.Until != nil {
		fmt.Printf("Until: %s\n", task.Until.Format("2006-01-02"))
	}
	if task.ParentUUID != nil {
		fmt.Printf("Parent: %s\n", *task.ParentUUID)
	}
	if len(task.DependsOn) > 0 {
		fmt.Printf("Depends on: %s\n", strings.Join(task.DependsOn, ", "))
	}

	return nil
}

// List lists all tasks with optional filtering
func (h *TaskHandler) List(ctx context.Context, static, showAll bool, status, priority, project, context string) error {
	if static {
		return h.listTasksStatic(ctx, showAll, status, priority, project, context)
	}

	return h.listTasksInteractive(ctx, showAll, status, priority, project, context)
}

func (h *TaskHandler) listTasksStatic(ctx context.Context, showAll bool, status, priority, project, context string) error {
	opts := repo.TaskListOptions{
		Status:   status,
		Priority: priority,
		Project:  project,
		Context:  context,
	}

	if !showAll && opts.Status == "" {
		opts.Status = "pending"
	}

	tasks, err := h.repos.Tasks.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Printf("No tasks found matching criteria\n")
		return nil
	}

	fmt.Printf("Found %d task(s):\n\n", len(tasks))
	for _, task := range tasks {
		printTask(task)
	}

	return nil
}

func (h *TaskHandler) listTasksInteractive(ctx context.Context, showAll bool, status, priority, project, _ string) error {
	taskTable := ui.NewTaskListFromTable(h.repos.Tasks, os.Stdout, os.Stdin, false, showAll, status, priority, project)
	return taskTable.Browse(ctx)
}

// Update updates a task using parsed flag values
func (h *TaskHandler) Update(ctx context.Context, taskID, description, status, priority, project, context, due, recur, until, parentUUID string, addTags, removeTags []string, addDeps, removeDeps string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	if description != "" {
		task.Description = description
	}
	if status != "" {
		task.Status = status
	}
	if priority != "" {
		task.Priority = priority
	}
	if project != "" {
		task.Project = project
	}
	if context != "" {
		task.Context = context
	}
	if due != "" {
		if dueTime, err := time.Parse("2006-01-02", due); err == nil {
			task.Due = &dueTime
		} else {
			return fmt.Errorf("invalid due date format, use YYYY-MM-DD: %w", err)
		}
	}
	if recur != "" {
		task.Recur = models.RRule(recur)
	}
	if until != "" {
		if untilTime, err := time.Parse("2006-01-02", until); err == nil {
			task.Until = &untilTime
		} else {
			return fmt.Errorf("invalid until date format, use YYYY-MM-DD: %w", err)
		}
	}
	if parentUUID != "" {
		task.ParentUUID = &parentUUID
	}

	for _, tag := range addTags {
		if !slices.Contains(task.Tags, tag) {
			task.Tags = append(task.Tags, tag)
		}
	}

	for _, tag := range removeTags {
		task.Tags = removeString(task.Tags, tag)
	}

	if addDeps != "" {
		deps := strings.SplitSeq(addDeps, ",")
		for dep := range deps {
			dep = strings.TrimSpace(dep)
			if dep != "" && !slices.Contains(task.DependsOn, dep) {
				task.DependsOn = append(task.DependsOn, dep)
			}
		}
	}

	if removeDeps != "" {
		deps := strings.SplitSeq(removeDeps, ",")
		for dep := range deps {
			dep = strings.TrimSpace(dep)
			task.DependsOn = removeString(task.DependsOn, dep)
		}
	}

	err = h.repos.Tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Task updated (ID: %d): %s\n", task.ID, task.Description)
	return nil
}

// EditInteractive opens an interactive task editor with status picker and priority toggle
func (h *TaskHandler) EditInteractive(ctx context.Context, taskID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	editor := ui.NewTaskEditor(task, h.repos.Tasks, ui.TaskEditOptions{})
	updated, err := editor.Edit(ctx)
	if err != nil {
		if err.Error() == "edit cancelled" {
			fmt.Println("Task edit cancelled")
			return nil
		}
		return fmt.Errorf("failed to edit task: %w", err)
	}

	fmt.Printf("Task updated (ID: %d): %s\n", updated.ID, updated.Description)
	fmt.Printf("Status: %s\n", ui.FormatStatusWithText(updated.Status))
	if updated.Priority != "" {
		fmt.Printf("Priority: %s\n", ui.FormatPriorityWithText(updated.Priority))
	}
	if updated.Project != "" {
		fmt.Printf("Project: %s\n", updated.Project)
	}

	return nil
}

// Delete deletes a task
func (h *TaskHandler) Delete(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("task ID required")
	}

	taskID := args[0]
	var task *models.Task
	var err error

	if id, parseErr := strconv.ParseInt(taskID, 10, 64); parseErr == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to find task: %w", err)
		}

		err = h.repos.Tasks.Delete(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to find task: %w", err)
		}

		err = h.repos.Tasks.Delete(ctx, task.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Task deleted (ID: %d): %s\n", task.ID, task.Description)
	return nil
}

// View displays a single task
func (h *TaskHandler) View(ctx context.Context, args []string, format string, jsonOutput, noMetadata bool) error {
	if len(args) < 1 {
		return fmt.Errorf("task ID required")
	}

	taskID := args[0]
	var task *models.Task
	var err error

	if id, parseErr := strconv.ParseInt(taskID, 10, 64); parseErr == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	if jsonOutput {
		return printTaskJSON(task)
	}

	if format == "brief" {
		printTask(task)
	} else {
		printTaskDetail(task, noMetadata)
	}
	return nil
}

// Start starts time tracking for a task
func (h *TaskHandler) Start(ctx context.Context, taskID string, description string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	active, err := h.repos.TimeEntries.GetActiveByTaskID(ctx, task.ID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("failed to check active time entry: %w", err)
	}
	if active != nil {
		duration := time.Since(active.StartTime)
		fmt.Printf("Task already started %s ago: %s\n", formatDuration(duration), task.Description)
		return nil
	}

	_, err = h.repos.TimeEntries.Start(ctx, task.ID, description)
	if err != nil {
		return fmt.Errorf("failed to start time tracking: %w", err)
	}

	fmt.Printf("Started task (ID: %d): %s\n", task.ID, task.Description)
	if description != "" {
		fmt.Printf("Note: %s\n", description)
	}

	return nil
}

// Stop stops time tracking for a task
func (h *TaskHandler) Stop(ctx context.Context, taskID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	entry, err := h.repos.TimeEntries.StopActiveByTaskID(ctx, task.ID)
	if err != nil {
		if err.Error() == "no active time entry found for task" {
			fmt.Printf("No active time tracking for task: %s\n", task.Description)
			return nil
		}
		return fmt.Errorf("failed to stop time tracking: %w", err)
	}

	fmt.Printf("Stopped task (ID: %d): %s\n", task.ID, task.Description)
	fmt.Printf("Time tracked: %s\n", formatDuration(entry.GetDuration()))

	return nil
}

// Timesheet shows time tracking summary
func (h *TaskHandler) Timesheet(ctx context.Context, days int, taskID string) error {
	var entries []*models.TimeEntry
	var err error

	if taskID != "" {
		var task *models.Task
		if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
			task, err = h.repos.Tasks.Get(ctx, id)
		} else {
			task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
		}

		if err != nil {
			return fmt.Errorf("failed to find task: %w", err)
		}

		entries, err = h.repos.TimeEntries.GetByTaskID(ctx, task.ID)
		if err != nil {
			return fmt.Errorf("failed to get time entries: %w", err)
		}

		fmt.Printf("Timesheet for task: %s\n\n", task.Description)
	} else {
		end := time.Now()
		start := end.AddDate(0, 0, -days)

		entries, err = h.repos.TimeEntries.GetByDateRange(ctx, start, end)
		if err != nil {
			return fmt.Errorf("failed to get time entries: %w", err)
		}

		fmt.Printf("Timesheet for last %d days:\n\n", days)
	}

	if len(entries) == 0 {
		fmt.Printf("No time entries found\n")
		return nil
	}

	taskTotals := make(map[int64]time.Duration)
	dayTotals := make(map[string]time.Duration)
	totalTime := time.Duration(0)

	fmt.Printf("%-20s %-10s %-12s %-40s %s\n", "Date", "Duration", "Status", "Task", "Note")
	fmt.Printf("%s\n", strings.Repeat("-", 95))

	for _, entry := range entries {
		task, err := h.repos.Tasks.Get(ctx, entry.TaskID)
		if err != nil {
			continue
		}

		duration := entry.GetDuration()
		day := entry.StartTime.Format("2006-01-02")
		status := "completed"
		if entry.IsActive() {
			status = "active"
		}

		taskTotals[entry.TaskID] += duration
		dayTotals[day] += duration
		totalTime += duration

		note := entry.Description
		if len(note) > 35 {
			note = note[:32] + "..."
		}

		taskDesc := task.Description
		if len(taskDesc) > 37 {
			taskDesc = taskDesc[:34] + "..."
		}

		fmt.Printf("%-20s %-10s %-12s %-40s %s\n",
			day,
			formatDuration(duration),
			status,
			fmt.Sprintf("[%d] %s", task.ID, taskDesc),
			note,
		)
	}

	fmt.Printf("%s\n", strings.Repeat("-", 95))
	fmt.Printf("Total time: %s\n", formatDuration(totalTime))

	return nil
}

// Done marks a task as completed
func (h *TaskHandler) Done(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("task ID required")
	}

	taskID := args[0]
	var task *models.Task
	var err error

	if id, parseErr := strconv.ParseInt(taskID, 10, 64); parseErr == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	if task.Status == "completed" {
		fmt.Printf("Task already completed: %s\n", task.Description)
		return nil
	}

	now := time.Now()
	task.Status = "completed"
	task.End = &now

	err = h.repos.Tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Task completed (ID: %d): %s\n", task.ID, task.Description)
	return nil
}

// ListProjects lists all projects with their task counts
func (h *TaskHandler) ListProjects(ctx context.Context, static bool, todoTxt ...bool) error {
	useTodoTxt := len(todoTxt) > 0 && todoTxt[0]
	if static {
		return h.listProjectsStatic(ctx, useTodoTxt)
	}
	return h.listProjectsInteractive(ctx, useTodoTxt)
}

func (h *TaskHandler) listProjectsStatic(ctx context.Context, todoTxt bool) error {
	tasks, err := h.repos.Tasks.List(ctx, repo.TaskListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list tasks for projects: %w", err)
	}

	projectCounts := make(map[string]int)
	for _, task := range tasks {
		if task.Project != "" {
			projectCounts[task.Project]++
		}
	}

	if len(projectCounts) == 0 {
		fmt.Printf("No projects found\n")
		return nil
	}

	projects := make([]string, 0, len(projectCounts))
	for project := range projectCounts {
		projects = append(projects, project)
	}
	slices.Sort(projects)

	fmt.Printf("Found %d project(s):\n\n", len(projects))
	for _, project := range projects {
		count := projectCounts[project]
		if todoTxt {
			fmt.Printf("+%s (%d task%s)\n", project, count, pluralize(count))
		} else {
			fmt.Printf("%s (%d task%s)\n", project, count, pluralize(count))
		}
	}

	return nil
}

// TODO: Add todo.txt format support to interactive mode
func (h *TaskHandler) listProjectsInteractive(ctx context.Context, _ bool) error {
	projectTable := ui.NewProjectListFromTable(h.repos.Tasks, nil, nil, false)
	return projectTable.Browse(ctx)
}

// ListTags lists all tags with their task counts
func (h *TaskHandler) ListTags(ctx context.Context, static bool) error {
	if static {
		return h.listTagsStatic(ctx)
	}

	return h.listTagsInteractive(ctx)
}

// ListContexts lists all contexts with their task counts
func (h *TaskHandler) ListContexts(ctx context.Context, static bool, todoTxt ...bool) error {
	useTodoTxt := len(todoTxt) > 0 && todoTxt[0]
	if static {
		return h.listContextsStatic(ctx, useTodoTxt)
	}
	return h.listContextsInteractive(ctx, useTodoTxt)
}

func (h *TaskHandler) listContextsStatic(ctx context.Context, todoTxt bool) error {
	tasks, err := h.repos.Tasks.List(ctx, repo.TaskListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list tasks for contexts: %w", err)
	}

	contextCounts := make(map[string]int)
	for _, task := range tasks {
		if task.Context != "" {
			contextCounts[task.Context]++
		}
	}

	if len(contextCounts) == 0 {
		fmt.Printf("No contexts found\n")
		return nil
	}

	contexts := make([]string, 0, len(contextCounts))
	for context := range contextCounts {
		contexts = append(contexts, context)
	}
	slices.Sort(contexts)

	fmt.Printf("Found %d context(s):\n\n", len(contexts))
	for _, context := range contexts {
		count := contextCounts[context]
		if todoTxt {
			fmt.Printf("@%s (%d task%s)\n", context, count, pluralize(count))
		} else {
			fmt.Printf("%s (%d task%s)\n", context, count, pluralize(count))
		}
	}

	return nil
}

func (h *TaskHandler) listContextsInteractive(ctx context.Context, todoTxt bool) error {
	fmt.Println("Interactive context listing not implemented yet - using static mode")
	return h.listContextsStatic(ctx, todoTxt)
}

func (h *TaskHandler) listTagsStatic(ctx context.Context) error {
	tasks, err := h.repos.Tasks.List(ctx, repo.TaskListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list tasks for tags: %w", err)
	}

	tagCounts := make(map[string]int)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}
	}

	if len(tagCounts) == 0 {
		fmt.Printf("No tags found\n")
		return nil
	}

	tags := make([]string, 0, len(tagCounts))
	for tag := range tagCounts {
		tags = append(tags, tag)
	}
	slices.Sort(tags)

	fmt.Printf("Found %d tag(s):\n\n", len(tags))
	for _, tag := range tags {
		count := tagCounts[tag]
		fmt.Printf("%s (%d task%s)\n", tag, count, pluralize(count))
	}

	return nil
}

func (h *TaskHandler) listTagsInteractive(ctx context.Context) error {
	tagTable := ui.NewTagListFromTable(h.repos.Tasks, nil, nil, false)
	return tagTable.Browse(ctx)
}

// SetRecur sets the recurrence rule for a task
func (h *TaskHandler) SetRecur(ctx context.Context, taskID, rule, until string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	if rule != "" {
		task.Recur = models.RRule(rule)
	}

	if until != "" {
		if untilTime, err := time.Parse("2006-01-02", until); err == nil {
			task.Until = &untilTime
		} else {
			return fmt.Errorf("invalid until date format, use YYYY-MM-DD: %w", err)
		}
	}

	err = h.repos.Tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task recurrence: %w", err)
	}

	fmt.Printf("Recurrence set for task (ID: %d): %s\n", task.ID, task.Description)
	if task.Recur != "" {
		fmt.Printf("Rule: %s\n", task.Recur)
	}
	if task.Until != nil {
		fmt.Printf("Until: %s\n", task.Until.Format("2006-01-02"))
	}

	return nil
}

// ClearRecur clears the recurrence rule from a task
func (h *TaskHandler) ClearRecur(ctx context.Context, taskID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	task.Recur = ""
	task.Until = nil

	err = h.repos.Tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to clear task recurrence: %w", err)
	}

	fmt.Printf("Recurrence cleared for task (ID: %d): %s\n", task.ID, task.Description)
	return nil
}

// ShowRecur displays the recurrence details for a task
func (h *TaskHandler) ShowRecur(ctx context.Context, taskID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	fmt.Printf("Task (ID: %d): %s\n", task.ID, task.Description)
	if task.Recur != "" {
		fmt.Printf("Recurrence rule: %s\n", task.Recur)
		if task.Until != nil {
			fmt.Printf("Recurrence until: %s\n", task.Until.Format("2006-01-02"))
		} else {
			fmt.Printf("Recurrence until: (no end date)\n")
		}
	} else {
		fmt.Printf("No recurrence set\n")
	}

	return nil
}

// AddDep adds a dependency to a task
func (h *TaskHandler) AddDep(ctx context.Context, taskID, dependsOnUUID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	if _, err := h.repos.Tasks.GetByUUID(ctx, dependsOnUUID); err != nil {
		return fmt.Errorf("dependency task not found: %w", err)
	}

	err = h.repos.Tasks.AddDependency(ctx, task.UUID, dependsOnUUID)
	if err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}

	fmt.Printf("Dependency added to task (ID: %d): %s\n", task.ID, task.Description)
	fmt.Printf("Now depends on: %s\n", dependsOnUUID)

	return nil
}

// RemoveDep removes a dependency from a task
func (h *TaskHandler) RemoveDep(ctx context.Context, taskID, dependsOnUUID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	err = h.repos.Tasks.RemoveDependency(ctx, task.UUID, dependsOnUUID)
	if err != nil {
		return fmt.Errorf("failed to remove dependency: %w", err)
	}

	fmt.Printf("Dependency removed from task (ID: %d): %s\n", task.ID, task.Description)
	fmt.Printf("No longer depends on: %s\n", dependsOnUUID)

	return nil
}

// ListDeps lists all dependencies for a task
func (h *TaskHandler) ListDeps(ctx context.Context, taskID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	fmt.Printf("Task (ID: %d): %s\n", task.ID, task.Description)

	if len(task.DependsOn) == 0 {
		fmt.Printf("No dependencies\n")
		return nil
	}

	fmt.Printf("Depends on %d task(s):\n", len(task.DependsOn))
	for _, depUUID := range task.DependsOn {
		depTask, err := h.repos.Tasks.GetByUUID(ctx, depUUID)
		if err != nil {
			fmt.Printf("  - %s (not found)\n", depUUID)
			continue
		}
		fmt.Printf("  - [%d] %s (UUID: %s)\n", depTask.ID, depTask.Description, depTask.UUID)
	}

	return nil
}

// BlockedByDep shows tasks that are blocked by the given task
func (h *TaskHandler) BlockedByDep(ctx context.Context, taskID string) error {
	var task *models.Task
	var err error

	if id, err_ := strconv.ParseInt(taskID, 10, 64); err_ == nil {
		task, err = h.repos.Tasks.Get(ctx, id)
	} else {
		task, err = h.repos.Tasks.GetByUUID(ctx, taskID)
	}

	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	fmt.Printf("Task (ID: %d): %s\n", task.ID, task.Description)

	dependents, err := h.repos.Tasks.GetDependents(ctx, task.UUID)
	if err != nil {
		return fmt.Errorf("failed to get dependent tasks: %w", err)
	}

	if len(dependents) == 0 {
		fmt.Printf("No tasks are blocked by this task\n")
		return nil
	}

	fmt.Printf("Blocks %d task(s):\n", len(dependents))
	for _, dep := range dependents {
		fmt.Printf("  - [%d] %s\n", dep.ID, dep.Description)
	}

	return nil
}
