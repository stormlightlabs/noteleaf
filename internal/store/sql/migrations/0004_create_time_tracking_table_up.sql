-- Time tracking table for tasks
CREATE TABLE IF NOT EXISTS time_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration_seconds INTEGER,
    description TEXT,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_time_entries_task_id ON time_entries(task_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_start_time ON time_entries(start_time);
CREATE INDEX IF NOT EXISTS idx_time_entries_end_time ON time_entries(end_time);