package handlers

import (
	"context"
	"fmt"
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

// CreateTask creates a new task
func CreateTask(ctx context.Context, args []string) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	return handler.createTask(ctx, args)
}

func (h *TaskHandler) createTask(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("task description required")
	}

	description := strings.Join(args, " ")

	task := &models.Task{
		UUID:        uuid.New().String(),
		Description: description,
		Status:      "pending",
	}

	id, err := h.repos.Tasks.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Task created (ID: %d, UUID: %s): %s\n", id, task.UUID, task.Description)
	return nil
}

// ListTasks lists all tasks with optional filtering
func ListTasks(ctx context.Context, static, showAll bool, status, priority, project string) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	if static {
		return handler.listTasksStatic(ctx, showAll, status, priority, project)
	}

	return handler.listTasksInteractive(ctx, showAll, status, priority, project)
}

func (h *TaskHandler) listTasksStatic(ctx context.Context, showAll bool, status, priority, project string) error {
	opts := repo.TaskListOptions{
		Status:   status,
		Priority: priority,
		Project:  project,
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

func (h *TaskHandler) listTasksInteractive(ctx context.Context, showAll bool, status, priority, project string) error {
	taskList := ui.NewTaskList(h.repos.Tasks, ui.TaskListOptions{
		ShowAll:  showAll,
		Status:   status,
		Priority: priority,
		Project:  project,
		Static:   false,
	})

	return taskList.Browse(ctx)
}

// UpdateTask updates an existing task
func UpdateTask(ctx context.Context, args []string) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	return handler.updateTask(ctx, args)
}

func (h *TaskHandler) updateTask(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("task ID required")
	}

	taskID := args[0]
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

	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--description" && i+1 < len(args):
			task.Description = args[i+1]
			i++
		case arg == "--status" && i+1 < len(args):
			task.Status = args[i+1]
			i++
		case arg == "--priority" && i+1 < len(args):
			task.Priority = args[i+1]
			i++
		case arg == "--project" && i+1 < len(args):
			task.Project = args[i+1]
			i++
		case arg == "--due" && i+1 < len(args):
			if dueTime, err := time.Parse("2006-01-02", args[i+1]); err == nil {
				task.Due = &dueTime
			}
			i++
		case strings.HasPrefix(arg, "--add-tag="):
			tag := strings.TrimPrefix(arg, "--add-tag=")
			if !slices.Contains(task.Tags, tag) {
				task.Tags = append(task.Tags, tag)
			}
		case strings.HasPrefix(arg, "--remove-tag="):
			tag := strings.TrimPrefix(arg, "--remove-tag=")
			task.Tags = removeString(task.Tags, tag)
		}
	}

	err = h.repos.Tasks.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Task updated (ID: %d): %s\n", task.ID, task.Description)
	return nil
}

// DeleteTask deletes a task
func DeleteTask(ctx context.Context, args []string) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	return handler.deleteTask(ctx, args)
}

func (h *TaskHandler) deleteTask(ctx context.Context, args []string) error {
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

// ViewTask displays a single task
func ViewTask(ctx context.Context, args []string) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	return handler.viewTask(ctx, args)
}

func (h *TaskHandler) viewTask(ctx context.Context, args []string) error {
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

	h.printTaskDetail(task)
	return nil
}

// DoneTask marks a task as completed
func DoneTask(ctx context.Context, args []string) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	return handler.doneTask(ctx, args)
}

func (h *TaskHandler) doneTask(ctx context.Context, args []string) error {
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
func ListProjects(ctx context.Context, static bool) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	if static {
		return handler.listProjectsStatic(ctx)
	}

	return handler.listProjectsInteractive(ctx)
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
	projectList := ui.NewProjectList(h.repos.Tasks, ui.ProjectListOptions{})
	return projectList.Browse(ctx)
}

// ListTags lists all tags with their task counts
func ListTags(ctx context.Context, static bool) error {
	handler, err := NewTaskHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize task handler: %w", err)
	}
	defer handler.Close()

	if static {
		return handler.listTagsStatic(ctx)
	}

	return handler.listTagsInteractive(ctx)
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
	tagList := ui.NewTagList(h.repos.Tasks, ui.TagListOptions{})
	return tagList.Browse(ctx)
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

	if len(task.Tags) > 0 {
		fmt.Printf(" @%s", strings.Join(task.Tags, " @"))
	}

	if task.Due != nil {
		fmt.Printf(" (due: %s)", task.Due.Format("2006-01-02"))
	}

	fmt.Println()
}

func (h *TaskHandler) printTaskDetail(task *models.Task) {
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

	if len(task.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(task.Tags, ", "))
	}

	if task.Due != nil {
		fmt.Printf("Due: %s\n", task.Due.Format("2006-01-02 15:04"))
	}

	fmt.Printf("Created: %s\n", task.Entry.Format("2006-01-02 15:04"))
	fmt.Printf("Modified: %s\n", task.Modified.Format("2006-01-02 15:04"))

	if task.Start != nil {
		fmt.Printf("Started: %s\n", task.Start.Format("2006-01-02 15:04"))
	}

	if task.End != nil {
		fmt.Printf("Completed: %s\n", task.End.Format("2006-01-02 15:04"))
	}

	if len(task.Annotations) > 0 {
		fmt.Printf("Annotations:\n")
		for _, annotation := range task.Annotations {
			fmt.Printf("  - %s\n", annotation)
		}
	}
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
	if count == 1 {
		return ""
	}
	return "s"
}
