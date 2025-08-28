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
		INSERT INTO tasks (uuid, description, status, priority, project, tags, due, entry, modified, end, start, annotations)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		task.UUID, task.Description, task.Status, task.Priority, task.Project,
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
		SELECT id, uuid, description, status, priority, project, tags, due, entry, modified, end, start, annotations
		FROM tasks WHERE id = ?`

	task := &models.Task{}
	var tags, annotations sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.UUID, &task.Description, &task.Status, &task.Priority, &task.Project,
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
		UPDATE tasks SET uuid = ?, description = ?, status = ?, priority = ?, project = ?,
		tags = ?, due = ?, modified = ?, end = ?, start = ?, annotations = ?
		WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query,
		task.UUID, task.Description, task.Status, task.Priority, task.Project,
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
	query := "SELECT id, uuid, description, status, priority, project, tags, due, entry, modified, end, start, annotations FROM tasks"

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
	if !opts.DueAfter.IsZero() {
		args = append(args, opts.DueAfter)
	}
	if !opts.DueBefore.IsZero() {
		args = append(args, opts.DueBefore)
	}

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	return args
}

func (r *TaskRepository) scanTaskRow(rows *sql.Rows, task *models.Task) error {
	var tags, annotations sql.NullString

	if err := rows.Scan(&task.ID, &task.UUID, &task.Description, &task.Status, &task.Priority,
		&task.Project, &tags, &task.Due, &task.Entry, &task.Modified, &task.End, &task.Start, &annotations); err != nil {
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

// GetTasksByTag retrieves all tasks with a specific tag
func (r *TaskRepository) GetTasksByTag(ctx context.Context, tag string) ([]*models.Task, error) {
	query := `
		SELECT tasks.id, tasks.uuid, tasks.description, tasks.status, tasks.priority, tasks.project, tasks.tags, tasks.due, tasks.entry, tasks.modified, tasks.end, tasks.start, tasks.annotations
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
