package handlers

import (
	"context"
	"encoding/json"
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
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
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
func (h *TaskHandler) Create(ctx context.Context, desc []string, priority, project, context, due string, tags []string) error {
	if len(desc) < 1 {
		return fmt.Errorf("task description required")
	}

	description := strings.Join(desc, " ")

	task := &models.Task{
		UUID:        uuid.New().String(),
		Description: description,
		Status:      "pending",
		Priority:    priority,
		Project:     project,
		Context:     context,
		Tags:        tags,
	}

	if due != "" {
		if dueTime, err := time.Parse("2006-01-02", due); err == nil {
			task.Due = &dueTime
		} else {
			return fmt.Errorf("invalid due date format, use YYYY-MM-DD: %w", err)
		}
	}

	id, err := h.repos.Tasks.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Task created (ID: %d, UUID: %s): %s\n", id, task.UUID, task.Description)

	if priority != "" {
		fmt.Printf("Priority: %s\n", priority)
	}
	if project != "" {
		fmt.Printf("Project: %s\n", project)
	}
	if context != "" {
		fmt.Printf("Context: %s\n", context)
	}
	if len(tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}
	if task.Due != nil {
		fmt.Printf("Due: %s\n", task.Due.Format("2006-01-02"))
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
		h.printTask(task)
	}

	return nil
}

func (h *TaskHandler) listTasksInteractive(ctx context.Context, showAll bool, status, priority, project, context string) error {
	taskTable := ui.NewTaskListFromTable(h.repos.Tasks, os.Stdout, os.Stdin, false, showAll, status, priority, project)
	return taskTable.Browse(ctx)
}

// Update updates a task using parsed flag values
func (h *TaskHandler) Update(ctx context.Context, taskID, description, status, priority, project, context, due string, addTags, removeTags []string) error {
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

	for _, tag := range addTags {
		if !slices.Contains(task.Tags, tag) {
			task.Tags = append(task.Tags, tag)
		}
	}

	for _, tag := range removeTags {
		task.Tags = removeString(task.Tags, tag)
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
		return h.printTaskJSON(task)
	}

	if format == "brief" {
		h.printTask(task)
	} else {
		h.printTaskDetail(task, noMetadata)
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
func (h *TaskHandler) ListProjects(ctx context.Context, static bool) error {
	if static {
		return h.listProjectsStatic(ctx)
	}
	return h.listProjectsInteractive(ctx)
}

func (h *TaskHandler) listProjectsStatic(ctx context.Context) error {
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
		fmt.Printf("%s (%d task%s)\n", project, count, pluralize(count))
	}

	return nil
}

func (h *TaskHandler) listProjectsInteractive(ctx context.Context) error {
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
func (h *TaskHandler) ListContexts(ctx context.Context, static bool) error {
	if static {
		return h.listContextsStatic(ctx)
	}
	return h.listContextsInteractive(ctx)
}

func (h *TaskHandler) listContextsStatic(ctx context.Context) error {
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
		fmt.Printf("%s (%d task%s)\n", context, count, pluralize(count))
	}

	return nil
}

func (h *TaskHandler) listContextsInteractive(ctx context.Context) error {
	fmt.Println("Interactive context listing not implemented yet - using static mode")
	return h.listContextsStatic(ctx)
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

func (h *TaskHandler) printTask(task *models.Task) {
	fmt.Printf("[%d] %s", task.ID, task.Description)

	if task.Status != "pending" {
		fmt.Printf(" (%s)", task.Status)
	}

	if task.Priority != "" {
		fmt.Printf(" [%s]", task.Priority)
	}

	if task.Project != "" {
		fmt.Printf(" +%s", task.Project)
	}

	if task.Context != "" {
		fmt.Printf(" @%s", task.Context)
	}

	if len(task.Tags) > 0 {
		fmt.Printf(" #%s", strings.Join(task.Tags, " #"))
	}

	if task.Due != nil {
		fmt.Printf(" (due: %s)", task.Due.Format("2006-01-02"))
	}

	fmt.Println()
}

func (h *TaskHandler) printTaskDetail(task *models.Task, noMetadata bool) {
	fmt.Printf("Task ID: %d\n", task.ID)
	fmt.Printf("UUID: %s\n", task.UUID)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Status: %s\n", task.Status)

	if task.Priority != "" {
		fmt.Printf("Priority: %s\n", task.Priority)
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
		fmt.Printf("Due: %s\n", task.Due.Format("2006-01-02 15:04"))
	}

	if !noMetadata {
		fmt.Printf("Created: %s\n", task.Entry.Format("2006-01-02 15:04"))
		fmt.Printf("Modified: %s\n", task.Modified.Format("2006-01-02 15:04"))

		if task.Start != nil {
			fmt.Printf("Started: %s\n", task.Start.Format("2006-01-02 15:04"))
		}

		if task.End != nil {
			fmt.Printf("Completed: %s\n", task.End.Format("2006-01-02 15:04"))
		}
	}

	if len(task.Annotations) > 0 {
		fmt.Printf("Annotations:\n")
		for _, annotation := range task.Annotations {
			fmt.Printf("  - %s\n", annotation)
		}
	}
}

func (h *TaskHandler) printTaskJSON(task *models.Task) error {
	jsonData, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func removeString(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func pluralize(count int) string {
	rule := plural.Cardinal.MatchPlural(language.English, count, 0, 0, 0, 0)
	switch rule {
	case plural.One:
		return ""
	default:
		return "s"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	hours := d.Hours()
	if hours < 24 {
		return fmt.Sprintf("%.1fh", hours)
	}
	days := int(hours / 24)
	remainingHours := hours - float64(days*24)
	if remainingHours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %.1fh", days, remainingHours)
}
