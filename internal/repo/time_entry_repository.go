package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// TimeEntryRepository provides database operations for time entries
type TimeEntryRepository struct {
	db *sql.DB
}

// NewTimeEntryRepository creates a new time entry repository
func NewTimeEntryRepository(db *sql.DB) *TimeEntryRepository {
	return &TimeEntryRepository{db: db}
}

// Start creates a new active time entry for a task
func (r *TimeEntryRepository) Start(ctx context.Context, taskID int64, description string) (*models.TimeEntry, error) {
	active, err := r.GetActiveByTaskID(ctx, taskID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for active time entry: %w", err)
	}
	if active != nil {
		return nil, fmt.Errorf("task already has an active time entry")
	}

	now := time.Now()
	entry := &models.TimeEntry{
		TaskID:      taskID,
		StartTime:   now,
		Description: description,
		Created:     now,
		Modified:    now,
	}

	query := `
		INSERT INTO time_entries (task_id, start_time, description, created, modified)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, entry.TaskID, entry.StartTime, entry.Description, entry.Created, entry.Modified)
	if err != nil {
		return nil, fmt.Errorf("failed to create time entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry ID: %w", err)
	}

	entry.ID = id
	return entry, nil
}

// Stop stops an active time entry by ID
func (r *TimeEntryRepository) Stop(ctx context.Context, id int64) (*models.TimeEntry, error) {
	entry, err := r.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}

	if !entry.IsActive() {
		return nil, fmt.Errorf("time entry is not active")
	}

	entry.Stop()

	query := `
		UPDATE time_entries
		SET end_time = ?, duration_seconds = ?, modified = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query, entry.EndTime, entry.DurationSeconds, entry.Modified, entry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to stop time entry: %w", err)
	}

	return entry, nil
}

// StopActiveByTaskID stops the active time entry for a task
func (r *TimeEntryRepository) StopActiveByTaskID(ctx context.Context, taskID int64) (*models.TimeEntry, error) {
	active, err := r.GetActiveByTaskID(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active time entry found for task")
		}
		return nil, fmt.Errorf("failed to get active time entry: %w", err)
	}

	return r.Stop(ctx, active.ID)
}

// Get retrieves a time entry by ID
func (r *TimeEntryRepository) Get(ctx context.Context, id int64) (*models.TimeEntry, error) {
	query := `
		SELECT id, task_id, start_time, end_time, duration_seconds, description, created, modified
		FROM time_entries
		WHERE id = ?
	`

	entry := &models.TimeEntry{}
	var durationSeconds sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.TaskID,
		&entry.StartTime,
		&entry.EndTime,
		&durationSeconds,
		&entry.Description,
		&entry.Created,
		&entry.Modified,
	)

	if durationSeconds.Valid {
		entry.DurationSeconds = durationSeconds.Int64
	}

	if err != nil {
		return nil, err
	}

	return entry, nil
}

// GetActiveByTaskID retrieves the active time entry for a task (if any)
func (r *TimeEntryRepository) GetActiveByTaskID(ctx context.Context, taskID int64) (*models.TimeEntry, error) {
	query := `
		SELECT id, task_id, start_time, end_time, duration_seconds, description, created, modified
		FROM time_entries
		WHERE task_id = ? AND end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`

	entry := &models.TimeEntry{}
	var durationSeconds sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(
		&entry.ID,
		&entry.TaskID,
		&entry.StartTime,
		&entry.EndTime,
		&durationSeconds,
		&entry.Description,
		&entry.Created,
		&entry.Modified,
	)

	if durationSeconds.Valid {
		entry.DurationSeconds = durationSeconds.Int64
	}

	if err != nil {
		return nil, err
	}

	return entry, nil
}

// GetByTaskID retrieves all time entries for a task
func (r *TimeEntryRepository) GetByTaskID(ctx context.Context, taskID int64) ([]*models.TimeEntry, error) {
	query := `
		SELECT id, task_id, start_time, end_time, duration_seconds, description, created, modified
		FROM time_entries
		WHERE task_id = ?
		ORDER BY start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query time entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.TimeEntry
	for rows.Next() {
		entry := &models.TimeEntry{}
		var durationSeconds sql.NullInt64
		err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&entry.StartTime,
			&entry.EndTime,
			&durationSeconds,
			&entry.Description,
			&entry.Created,
			&entry.Modified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time entry: %w", err)
		}
		if durationSeconds.Valid {
			entry.DurationSeconds = durationSeconds.Int64
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate time entries: %w", err)
	}

	return entries, nil
}

// GetByDateRange retrieves time entries within a date range
func (r *TimeEntryRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*models.TimeEntry, error) {
	query := `
		SELECT id, task_id, start_time, end_time, duration_seconds, description, created, modified
		FROM time_entries
		WHERE start_time >= ? AND start_time <= ?
		ORDER BY start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query time entries by date range: %w", err)
	}
	defer rows.Close()

	var entries []*models.TimeEntry
	for rows.Next() {
		entry := &models.TimeEntry{}
		var durationSeconds sql.NullInt64

		if err := rows.Scan(&entry.ID, &entry.TaskID, &entry.StartTime,
			&entry.EndTime, &durationSeconds, &entry.Description,
			&entry.Created, &entry.Modified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan time entry: %w", err)
		}

		if durationSeconds.Valid {
			entry.DurationSeconds = durationSeconds.Int64
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate time entries: %w", err)
	}

	return entries, nil
}

// GetTotalTimeByTaskID calculates total time spent on a task
func (r *TimeEntryRepository) GetTotalTimeByTaskID(ctx context.Context, taskID int64) (time.Duration, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE
				WHEN end_time IS NULL THEN
					(strftime('%s', 'now') - strftime('%s', start_time))
				ELSE
					duration_seconds
			END
		), 0) as total_seconds
		FROM time_entries
		WHERE task_id = ?
	`

	var totalSeconds int64
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&totalSeconds)
	if err != nil {
		return 0, fmt.Errorf("failed to get total time: %w", err)
	}

	return time.Duration(totalSeconds) * time.Second, nil
}

// Delete removes a time entry
func (r *TimeEntryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM time_entries WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete time entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("time entry not found")
	}

	return nil
}
