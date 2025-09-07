package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// TaskRecord adapts models.Task to work with DataTable
type TaskRecord struct {
	*models.Task
}

func (t *TaskRecord) GetField(name string) any {
	switch name {
	case "id":
		return t.ID
	case "uuid":
		return t.UUID
	case "description":
		return t.Description
	case "status":
		return t.Status
	case "priority":
		return t.Priority
	case "project":
		return t.Project
	case "tags":
		return t.Tags
	case "due":
		return t.Due
	case "entry":
		return t.Entry
	case "start":
		return t.Start
	case "end":
		return t.End
	case "modified":
		return t.Modified
	case "annotations":
		return t.Annotations
	default:
		return ""
	}
}

// TaskDataSource adapts TaskRepository to work with DataTable
type TaskDataSource struct {
	repo     utils.TestTaskRepository
	showAll  bool
	status   string
	priority string
	project  string
}

func (t *TaskDataSource) Load(ctx context.Context, opts DataOptions) ([]DataRecord, error) {
	repoOpts := repo.TaskListOptions{
		SortBy:    "modified",
		SortOrder: "DESC",
		Limit:     50,
	}

	if !t.showAll && t.status == "" {
		repoOpts.Status = "pending"
	}
	if t.status != "" {
		repoOpts.Status = t.status
	}
	if t.priority != "" {
		repoOpts.Priority = t.priority
	}
	if t.project != "" {
		repoOpts.Project = t.project
	}

	tasks, err := t.repo.List(ctx, repoOpts)
	if err != nil {
		return nil, err
	}

	records := make([]DataRecord, len(tasks))
	for i, task := range tasks {
		records[i] = &TaskRecord{Task: task}
	}

	return records, nil
}

func (t *TaskDataSource) Count(ctx context.Context, opts DataOptions) (int, error) {
	records, err := t.Load(ctx, opts)
	if err != nil {
		return 0, err
	}
	return len(records), nil
}

func formatTaskForView(task *models.Task) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Task %d\n\n", task.ID))
	content.WriteString(fmt.Sprintf("**UUID:** %s\n", task.UUID))
	content.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))
	content.WriteString(fmt.Sprintf("**Status:** %s\n", task.Status))

	if task.Priority != "" {
		content.WriteString(fmt.Sprintf("**Priority:** %s\n", task.Priority))
	}

	if task.Project != "" {
		content.WriteString(fmt.Sprintf("**Project:** %s\n", task.Project))
	}

	if len(task.Tags) > 0 {
		content.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(task.Tags, ", ")))
	}

	if task.Due != nil {
		content.WriteString(fmt.Sprintf("**Due:** %s\n", task.Due.Format("2006-01-02 15:04")))
	}

	content.WriteString(fmt.Sprintf("**Created:** %s\n", task.Entry.Format("2006-01-02 15:04")))
	content.WriteString(fmt.Sprintf("**Modified:** %s\n", task.Modified.Format("2006-01-02 15:04")))

	if task.Start != nil {
		content.WriteString(fmt.Sprintf("**Started:** %s\n", task.Start.Format("2006-01-02 15:04")))
	}

	if task.End != nil {
		content.WriteString(fmt.Sprintf("**Completed:** %s\n", task.End.Format("2006-01-02 15:04")))
	}

	if len(task.Annotations) > 0 {
		content.WriteString("\n**Annotations:**\n")
		for _, annotation := range task.Annotations {
			content.WriteString(fmt.Sprintf("- %s\n", annotation))
		}
	}

	return content.String()
}

func formatPriorityField(priority string) string {
	if priority == "" {
		return "-"
	}

	titlecase := utils.Titlecase(priority)
	padded := fmt.Sprintf("%-10s", titlecase)

	switch strings.ToLower(priority) {
	case "high", "urgent":
		return PriorityHigh.Render(padded)
	case "medium":
		return PriorityMedium.Render(padded)
	case "low":
		return PriorityLow.Render(padded)
	default:
		return padded
	}
}

// NewTaskDataTable creates a new DataTable for browsing tasks
func NewTaskDataTable(repo utils.TestTaskRepository, opts DataTableOptions, showAll bool, status, priority, project string) *DataTable {
	if opts.Title == "" {
		title := "Tasks"
		if showAll {
			title += " (showing all)"
		} else {
			title += " (pending only)"
		}
		opts.Title = title
	}

	opts.Fields = []Field{
		{Name: "id", Title: "ID", Width: 4},
		{Name: "description", Title: "Description", Width: 40,
			Formatter: func(v any) string {
				desc := fmt.Sprintf("%v", v)
				if len(desc) > 38 {
					return desc[:35] + "..."
				}
				return desc
			}},
		{Name: "status", Title: "Status", Width: 10,
			Formatter: func(v any) string {
				status := fmt.Sprintf("%v", v)
				if len(status) > 8 {
					return status[:8]
				}
				return status
			}},
		{Name: "priority", Title: "Priority", Width: 10,
			Formatter: func(v any) string {
				priority := fmt.Sprintf("%v", v)
				return formatPriorityField(priority)
			}},
		{Name: "project", Title: "Project", Width: 15,
			Formatter: func(v any) string {
				project := fmt.Sprintf("%v", v)
				if project == "" {
					return "-"
				}
				if len(project) > 13 {
					return project[:10] + "..."
				}
				return project
			}},
	}

	if opts.ViewHandler == nil {
		opts.ViewHandler = func(record DataRecord) string {
			if taskRecord, ok := record.(*TaskRecord); ok {
				return formatTaskForView(taskRecord.Task)
			}
			return "Unable to display task"
		}
	}

	if len(opts.Actions) == 0 {
		opts.Actions = []Action{
			{
				Key:         "d",
				Description: "mark done",
				Handler: func(record DataRecord) tea.Cmd {
					return func() tea.Msg {
						if taskRecord, ok := record.(*TaskRecord); ok {
							if taskRecord.Status == "completed" {
								return dataErrorMsg(fmt.Errorf("task already completed"))
							}
							taskRecord.Status = "completed"
							taskRecord.End = &time.Time{}
							*taskRecord.End = time.Now()
							err := repo.Update(context.Background(), taskRecord.Task)
							if err != nil {
								return dataErrorMsg(err)
							}
							return dataLoadedMsg([]DataRecord{})
						}
						return dataErrorMsg(fmt.Errorf("invalid task record"))
					}
				},
			},
		}
	}

	source := &TaskDataSource{
		repo:     repo,
		showAll:  showAll,
		status:   status,
		priority: priority,
		project:  project,
	}

	return NewDataTable(source, opts)
}

// NewTaskListFromTable creates a TaskList-compatible interface using DataTable
func NewTaskListFromTable(repo utils.TestTaskRepository, output io.Writer, input io.Reader, static bool, showAll bool, status, priority, project string) *DataTable {
	opts := DataTableOptions{
		Output: output,
		Input:  input,
		Static: static,
	}
	return NewTaskDataTable(repo, opts, showAll, status, priority, project)
}
