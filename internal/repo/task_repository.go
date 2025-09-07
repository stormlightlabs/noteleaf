package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// TaskListOptions defines options for listing tasks
type TaskListOptions struct {
	Status    string
	Priority  string
	Project   string
	Context   string
	DueAfter  time.Time
	DueBefore time.Time
	Search    string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// ProjectSummary represents a project with its task count
type ProjectSummary struct {
	Name      string `json:"name"`
	TaskCount int    `json:"task_count"`
}

// TagSummary represents a tag with its task count
type TagSummary struct {
	Name      string `json:"name"`
	TaskCount int    `json:"task_count"`
}

// ContextSummary represents a context with its task count
type ContextSummary struct {
	Name      string `json:"name"`
	TaskCount int    `json:"task_count"`
}

// TaskRepository provides database operations for tasks
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create stores a new task and returns its assigned ID
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) (int64, error) {
	now := time.Now()
	task.Entry = now
	task.Modified = now

	tags, err := task.MarshalTags()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal tags: %w", err)
	}

	annotations, err := task.MarshalAnnotations()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal annotations: %w", err)
	}

	query := `
		INSERT INTO tasks (uuid, description, status, priority, project, context, tags, due, entry, modified, end, start, annotations)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		task.UUID, task.Description, task.Status, task.Priority, task.Project, task.Context,
		tags, task.Due, task.Entry, task.Modified, task.End, task.Start, annotations)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = id
	return id, nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, id int64) (*models.Task, error) {
	query := `
		SELECT id, uuid, description, status, priority, project, context, tags, due, entry, modified, end, start, annotations
		FROM tasks WHERE id = ?`

	task := &models.Task{}
	var tags, annotations sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.UUID, &task.Description, &task.Status, &task.Priority, &task.Project, &task.Context,
		&tags, &task.Due, &task.Entry, &task.Modified, &task.End, &task.Start, &annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if tags.Valid {
		if err := task.UnmarshalTags(tags.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	if annotations.Valid {
		if err := task.UnmarshalAnnotations(annotations.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	return task, nil
}

// Update modifies an existing task
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	task.Modified = time.Now()

	tags, err := task.MarshalTags()
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	annotations, err := task.MarshalAnnotations()
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	query := `
		UPDATE tasks SET uuid = ?, description = ?, status = ?, priority = ?, project = ?, context = ?,
		tags = ?, due = ?, modified = ?, end = ?, start = ?, annotations = ?
		WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query,
		task.UUID, task.Description, task.Status, task.Priority, task.Project, task.Context,
		tags, task.Due, task.Modified, task.End, task.Start, annotations, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// Delete removes a task by ID
func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM tasks WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// List retrieves tasks with optional filtering and sorting
func (r *TaskRepository) List(ctx context.Context, opts TaskListOptions) ([]*models.Task, error) {
	query := r.buildListQuery(opts)
	args := r.buildListArgs(opts)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := r.scanTaskRow(rows, task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *TaskRepository) buildListQuery(opts TaskListOptions) string {
	query := "SELECT id, uuid, description, status, priority, project, context, tags, due, entry, modified, end, start, annotations FROM tasks"

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
	}
	if opts.Priority != "" {
		conditions = append(conditions, "priority = ?")
	}
	if opts.Project != "" {
		conditions = append(conditions, "project = ?")
	}
	if opts.Context != "" {
		conditions = append(conditions, "context = ?")
	}
	if !opts.DueAfter.IsZero() {
		conditions = append(conditions, "due >= ?")
	}
	if !opts.DueBefore.IsZero() {
		conditions = append(conditions, "due <= ?")
	}

	if opts.Search != "" {
		searchConditions := []string{
			"description LIKE ?",
			"project LIKE ?",
			"context LIKE ?",
			"tags LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if opts.SortBy != "" {
		order := "ASC"
		if strings.ToUpper(opts.SortOrder) == "DESC" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.SortBy, order)
	} else {
		query += " ORDER BY modified DESC"
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	return query
}

func (r *TaskRepository) buildListArgs(opts TaskListOptions) []any {
	var args []any

	if opts.Status != "" {
		args = append(args, opts.Status)
	}
	if opts.Priority != "" {
		args = append(args, opts.Priority)
	}
	if opts.Project != "" {
		args = append(args, opts.Project)
	}
	if opts.Context != "" {
		args = append(args, opts.Context)
	}
	if !opts.DueAfter.IsZero() {
		args = append(args, opts.DueAfter)
	}
	if !opts.DueBefore.IsZero() {
		args = append(args, opts.DueBefore)
	}

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	return args
}

func (r *TaskRepository) scanTaskRow(rows *sql.Rows, task *models.Task) error {
	var tags, annotations sql.NullString

	if err := rows.Scan(&task.ID, &task.UUID, &task.Description, &task.Status, &task.Priority,
		&task.Project, &task.Context, &tags, &task.Due, &task.Entry, &task.Modified, &task.End, &task.Start, &annotations); err != nil {
		return fmt.Errorf("failed to scan task row: %w", err)
	}

	if tags.Valid {
		if err := task.UnmarshalTags(tags.String); err != nil {
			return fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	if annotations.Valid {
		if err := task.UnmarshalAnnotations(annotations.String); err != nil {
			return fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	return nil
}

// Find retrieves tasks matching specific conditions
func (r *TaskRepository) Find(ctx context.Context, conditions TaskListOptions) ([]*models.Task, error) {
	return r.List(ctx, conditions)
}

// Count returns the number of tasks matching conditions
func (r *TaskRepository) Count(ctx context.Context, opts TaskListOptions) (int64, error) {
	query := "SELECT COUNT(*) FROM tasks"
	args := []any{}

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, opts.Status)
	}
	if opts.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, opts.Priority)
	}
	if opts.Project != "" {
		conditions = append(conditions, "project = ?")
		args = append(args, opts.Project)
	}
	if !opts.DueAfter.IsZero() {
		conditions = append(conditions, "due >= ?")
		args = append(args, opts.DueAfter)
	}
	if !opts.DueBefore.IsZero() {
		conditions = append(conditions, "due <= ?")
		args = append(args, opts.DueBefore)
	}

	if opts.Search != "" {
		searchConditions := []string{
			"description LIKE ?",
			"project LIKE ?",
			"tags LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	return count, nil
}

// GetByUUID retrieves a task by UUID
func (r *TaskRepository) GetByUUID(ctx context.Context, uuid string) (*models.Task, error) {
	query := `
		SELECT id, uuid, description, status, priority, project, tags, due, entry, modified, end, start, annotations
		FROM tasks WHERE uuid = ?`

	task := &models.Task{}
	var tags, annotations sql.NullString

	if err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&task.ID, &task.UUID, &task.Description, &task.Status, &task.Priority, &task.Project,
		&tags, &task.Due, &task.Entry, &task.Modified, &task.End, &task.Start, &annotations); err != nil {
		return nil, fmt.Errorf("failed to get task by UUID: %w", err)
	}

	if tags.Valid {
		if err := task.UnmarshalTags(tags.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	if annotations.Valid {
		if err := task.UnmarshalAnnotations(annotations.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	return task, nil
}

// GetPending retrieves all pending tasks
func (r *TaskRepository) GetPending(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: "pending"})
}

// GetCompleted retrieves all completed tasks
func (r *TaskRepository) GetCompleted(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: "completed"})
}

// GetByProject retrieves all tasks for a specific project
func (r *TaskRepository) GetByProject(ctx context.Context, project string) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Project: project})
}

// GetByContext retrieves all tasks for a specific context
func (r *TaskRepository) GetByContext(ctx context.Context, context string) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Context: context})
}

// GetProjects retrieves all unique project names with their task counts
func (r *TaskRepository) GetProjects(ctx context.Context) ([]ProjectSummary, error) {
	query := `
		SELECT project, COUNT(*) as task_count
		FROM tasks
		WHERE project != '' AND project IS NOT NULL
		GROUP BY project
		ORDER BY project`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer rows.Close()

	var projects []ProjectSummary
	for rows.Next() {
		var project ProjectSummary
		if err := rows.Scan(&project.Name, &project.TaskCount); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// GetTags retrieves all unique tags with their task counts
func (r *TaskRepository) GetTags(ctx context.Context) ([]TagSummary, error) {
	query := `
		SELECT DISTINCT json_each.value as tag, COUNT(tasks.id) as task_count
		FROM tasks, json_each(tasks.tags)
		WHERE tasks.tags != '' AND tasks.tags IS NOT NULL
		GROUP BY tag
		ORDER BY tag`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []TagSummary
	for rows.Next() {
		var tag TagSummary
		if err := rows.Scan(&tag.Name, &tag.TaskCount); err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetContexts retrieves all unique context names with their task counts
func (r *TaskRepository) GetContexts(ctx context.Context) ([]ContextSummary, error) {
	query := `
		SELECT context, COUNT(*) as task_count
		FROM tasks
		WHERE context != '' AND context IS NOT NULL
		GROUP BY context
		ORDER BY context`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contexts: %w", err)
	}
	defer rows.Close()

	var contexts []ContextSummary
	for rows.Next() {
		var context ContextSummary
		if err := rows.Scan(&context.Name, &context.TaskCount); err != nil {
			return nil, fmt.Errorf("failed to scan context row: %w", err)
		}
		contexts = append(contexts, context)
	}

	return contexts, rows.Err()
}

// GetTasksByTag retrieves all tasks with a specific tag
func (r *TaskRepository) GetTasksByTag(ctx context.Context, tag string) ([]*models.Task, error) {
	query := `
		SELECT tasks.id, tasks.uuid, tasks.description, tasks.status, tasks.priority, tasks.project, tasks.context, tasks.tags, tasks.due, tasks.entry, tasks.modified, tasks.end, tasks.start, tasks.annotations
		FROM tasks, json_each(tasks.tags)
		WHERE tasks.tags != '' AND tasks.tags IS NOT NULL AND json_each.value = ?
		ORDER BY tasks.modified DESC`

	rows, err := r.db.QueryContext(ctx, query, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by tag: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := r.scanTaskRow(rows, task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// GetTodo retrieves all tasks with todo status
func (r *TaskRepository) GetTodo(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: models.StatusTodo})
}

// GetInProgress retrieves all tasks with in-progress status
func (r *TaskRepository) GetInProgress(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: models.StatusInProgress})
}

// GetBlocked retrieves all tasks with blocked status
func (r *TaskRepository) GetBlocked(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: models.StatusBlocked})
}

// GetDone retrieves all tasks with done status
func (r *TaskRepository) GetDone(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: models.StatusDone})
}

// GetAbandoned retrieves all tasks with abandoned status
func (r *TaskRepository) GetAbandoned(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Status: models.StatusAbandoned})
}

// GetByPriority retrieves all tasks with a specific priority
//
//	We need special handling for empty priority by using raw SQL
func (r *TaskRepository) GetByPriority(ctx context.Context, priority string) ([]*models.Task, error) {
	if priority == "" {
		query := `
			SELECT id, uuid, description, status, priority, project, context, tags, due, entry, modified, end, start, annotations
			FROM tasks
			WHERE priority = '' OR priority IS NULL
			ORDER BY modified DESC`

		rows, err := r.db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks by empty priority: %w", err)
		}
		defer rows.Close()

		var tasks []*models.Task
		for rows.Next() {
			task := &models.Task{}
			if err := r.scanTaskRow(rows, task); err != nil {
				return nil, err
			}
			tasks = append(tasks, task)
		}

		return tasks, rows.Err()
	}

	return r.List(ctx, TaskListOptions{Priority: priority})
}

// GetHighPriority retrieves all high priority tasks
func (r *TaskRepository) GetHighPriority(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Priority: models.PriorityHigh})
}

// GetMediumPriority retrieves all medium priority tasks
func (r *TaskRepository) GetMediumPriority(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Priority: models.PriorityMedium})
}

// GetLowPriority retrieves all low priority tasks
func (r *TaskRepository) GetLowPriority(ctx context.Context) ([]*models.Task, error) {
	return r.List(ctx, TaskListOptions{Priority: models.PriorityLow})
}

// GetStatusSummary returns a summary of tasks by status
func (r *TaskRepository) GetStatusSummary(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM tasks
		GROUP BY status
		ORDER BY status`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get status summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status summary row: %w", err)
		}
		summary[status] = count
	}

	return summary, rows.Err()
}

// GetPrioritySummary returns a summary of tasks by priority
func (r *TaskRepository) GetPrioritySummary(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT
			CASE
				WHEN priority = '' OR priority IS NULL THEN 'No Priority'
				ELSE priority
			END as priority_group,
			COUNT(*) as count
		FROM tasks
		GROUP BY priority_group
		ORDER BY priority_group`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[string]int64)
	for rows.Next() {
		var priority string
		var count int64
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, fmt.Errorf("failed to scan priority summary row: %w", err)
		}
		summary[priority] = count
	}

	return summary, rows.Err()
}
