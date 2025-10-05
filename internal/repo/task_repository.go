// TODO: extend queryMany composition for GetTasksBy... methods
package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

var (
	marshalTaskTags          = (*models.Task).MarshalTags
	marshalTaskAnnotations   = (*models.Task).MarshalAnnotations
	unmarshalTaskTags        = (*models.Task).UnmarshalTags
	unmarshalTaskAnnotations = (*models.Task).UnmarshalAnnotations
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

// scanTask scans a database row into a Task model
func (r *TaskRepository) scanTask(s scanner) (*models.Task, error) {
	task := &models.Task{}
	var tags, annotations sql.NullString
	var parentUUID sql.NullString
	var priority, project, context sql.NullString

	if err := s.Scan(
		&task.ID, &task.UUID, &task.Description, &task.Status, &priority,
		&project, &context, &tags,
		&task.Due, &task.Entry, &task.Modified, &task.End, &task.Start, &annotations,
		&task.Recur, &task.Until, &parentUUID,
	); err != nil {
		return nil, err
	}

	if priority.Valid {
		task.Priority = priority.String
	}
	if project.Valid {
		task.Project = project.String
	}
	if context.Valid {
		task.Context = context.String
	}
	if parentUUID.Valid {
		task.ParentUUID = &parentUUID.String
	}

	if tags.Valid {
		if err := unmarshalTaskTags(task, tags.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	if annotations.Valid {
		if err := unmarshalTaskAnnotations(task, annotations.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	return task, nil
}

// queryOne executes a query that returns a single task
func (r *TaskRepository) queryOne(ctx context.Context, query string, args ...any) (*models.Task, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	task, err := r.scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}
	return task, nil
}

// queryMany executes a query that returns multiple tasks
func (r *TaskRepository) queryMany(ctx context.Context, query string, args ...any) ([]*models.Task, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tasks: %w", err)
	}

	return tasks, nil
}

// Create stores a new task and returns its assigned ID
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) (int64, error) {
	now := time.Now()
	task.Entry = now
	task.Modified = now

	tags, err := marshalTaskTags(task)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal tags: %w", err)
	}

	annotations, err := marshalTaskAnnotations(task)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal annotations: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryTaskInsert,
		task.UUID, task.Description, task.Status, task.Priority, task.Project, task.Context,
		tags, task.Due, task.Entry, task.Modified, task.End, task.Start, annotations,
		task.Recur, task.Until, task.ParentUUID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = id

	for _, depUUID := range task.DependsOn {
		if err := r.AddDependency(ctx, task.UUID, depUUID); err != nil {
			return 0, fmt.Errorf("failed to add dependency: %w", err)
		}
	}

	return id, nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, id int64) (*models.Task, error) {
	task, err := r.queryOne(ctx, queryTaskByID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if err := r.PopulateDependencies(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to populate dependencies: %w", err)
	}

	return task, nil
}

// Update modifies an existing task
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	task.Modified = time.Now()

	tags, err := marshalTaskTags(task)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	annotations, err := marshalTaskAnnotations(task)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	if _, err = r.db.ExecContext(ctx, queryTaskUpdate,
		task.UUID, task.Description, task.Status, task.Priority, task.Project, task.Context,
		tags, task.Due, task.Modified, task.End, task.Start, annotations,
		task.Recur, task.Until, task.ParentUUID,
		task.ID,
	); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if err := r.ClearDependencies(ctx, task.UUID); err != nil {
		return fmt.Errorf("failed to clear dependencies: %w", err)
	}

	for _, depUUID := range task.DependsOn {
		if err := r.AddDependency(ctx, task.UUID, depUUID); err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}
	}

	return nil
}

// Delete removes a task by ID
func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, queryTaskDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// List retrieves tasks with optional filtering and sorting
func (r *TaskRepository) List(ctx context.Context, opts TaskListOptions) ([]*models.Task, error) {
	query := r.buildListQuery(opts)
	args := r.buildListArgs(opts)
	return r.queryMany(ctx, query, args...)
}

func (r *TaskRepository) buildListQuery(opts TaskListOptions) string {
	query := queryTasksList
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
	if opts.Context != "" {
		conditions = append(conditions, "context = ?")
		args = append(args, opts.Context)
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
			"context LIKE ?",
			"tags LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
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
	task, err := r.queryOne(ctx, queryTaskByUUID, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get task by UUID: %w", err)
	}

	// Populate dependencies from task_dependencies table
	if err := r.PopulateDependencies(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to populate dependencies: %w", err)
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
		SELECT t.id, t.uuid, t.description, t.status, t.priority, t.project, t.context,
		       t.tags, t.due, t.entry, t.modified, t.end, t.start, t.annotations,
		       t.recur, t.until, t.parent_uuid
		FROM tasks t, json_each(t.tags)
		WHERE t.tags != '' AND t.tags IS NOT NULL AND json_each.value = ?
		ORDER BY t.modified DESC`

	return r.queryMany(ctx, query, tag)
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

// GetByPriority retrieves all tasks with a specific priority with special handling for empty priority by using raw SQL
func (r *TaskRepository) GetByPriority(ctx context.Context, priority string) ([]*models.Task, error) {
	if priority == "" {
		query := "SELECT " + taskColumns + " FROM tasks WHERE priority = '' OR priority IS NULL ORDER BY modified DESC"
		return r.queryMany(ctx, query)
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
	query := `SELECT status, COUNT(*) as count FROM tasks GROUP BY status ORDER BY status`

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
			COUNT(*) as count FROM tasks GROUP BY priority_group ORDER BY priority_group`

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

// AddDependency creates a dependency relationship where taskUUID depends on dependsOnUUID.
func (r *TaskRepository) AddDependency(ctx context.Context, taskUUID, dependsOnUUID string) error {
	if _, err := r.db.ExecContext(ctx, `INSERT INTO task_dependencies (task_uuid, depends_on_uuid) VALUES (?, ?)`, taskUUID, dependsOnUUID); err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}
	return nil
}

// RemoveDependency deletes a specific dependency relationship.
func (r *TaskRepository) RemoveDependency(ctx context.Context, taskUUID, dependsOnUUID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM task_dependencies WHERE task_uuid = ? AND depends_on_uuid = ?`, taskUUID, dependsOnUUID); err != nil {
		return fmt.Errorf("failed to remove dependency: %w", err)
	}
	return nil
}

// ClearDependencies removes all dependencies for a given task.
func (r *TaskRepository) ClearDependencies(ctx context.Context, taskUUID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM task_dependencies WHERE task_uuid = ?`, taskUUID); err != nil {
		return fmt.Errorf("failed to clear dependencies: %w", err)
	}
	return nil
}

// GetDependencies returns the UUIDs of tasks this task depends on.
func (r *TaskRepository) GetDependencies(ctx context.Context, taskUUID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT depends_on_uuid FROM task_dependencies WHERE task_uuid = ?`, taskUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

// PopulateDependencies loads dependency UUIDs from task_dependencies table into task.DependsOn
func (r *TaskRepository) PopulateDependencies(ctx context.Context, task *models.Task) error {
	if deps, err := r.GetDependencies(ctx, task.UUID); err != nil {
		return err
	} else {
		task.DependsOn = deps
	}
	return nil
}

// GetDependents returns tasks that are blocked by a given UUID.
func (r *TaskRepository) GetDependents(ctx context.Context, blockingUUID string) ([]*models.Task, error) {
	query := `
		SELECT t.id, t.uuid, t.description, t.status, t.priority, t.project, t.context,
		       t.tags, t.due, t.entry, t.modified, t.end, t.start, t.annotations, t.recur, t.until, t.parent_uuid
		FROM tasks t JOIN task_dependencies d ON t.uuid = d.task_uuid WHERE d.depends_on_uuid = ?`

	tasks, err := r.queryMany(ctx, query, blockingUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependents: %w", err)
	}

	for _, task := range tasks {
		if err := r.PopulateDependencies(ctx, task); err != nil {
			return nil, fmt.Errorf("failed to populate dependencies: %w", err)
		}
	}
	return tasks, nil
}

// GetBlockedTasks finds tasks that are blocked by a given UUID.
func (r *TaskRepository) GetBlockedTasks(ctx context.Context, blockingUUID string) ([]*models.Task, error) {
	query := `
		SELECT t.id, t.uuid, t.description, t.status, t.priority, t.project, t.context,
		       t.tags, t.due, t.entry, t.modified, t.end, t.start, t.annotations, t.recur, t.until, t.parent_uuid
		FROM tasks t
		JOIN task_dependencies d ON t.uuid = d.task_uuid
		WHERE d.depends_on_uuid = ?`

	tasks, err := r.queryMany(ctx, query, blockingUUID)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if err := r.PopulateDependencies(ctx, task); err != nil {
			return nil, fmt.Errorf("failed to populate dependencies: %w", err)
		}
	}
	return tasks, nil
}
