-- Add recurrence fields directly to tasks
ALTER TABLE tasks ADD COLUMN recur TEXT;   -- e.g. "daily", "weekly", ISO8601 rule
ALTER TABLE tasks ADD COLUMN until DATETIME; -- optional end date for recurrence
ALTER TABLE tasks ADD COLUMN parent_uuid TEXT; -- parent/template task UUID

-- Create dependencies table
CREATE TABLE IF NOT EXISTS task_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_uuid TEXT NOT NULL,     -- the dependent task
    depends_on_uuid TEXT NOT NULL, -- the blocking task
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY(task_uuid) REFERENCES tasks(uuid) ON DELETE CASCADE,
    FOREIGN KEY(depends_on_uuid) REFERENCES tasks(uuid) ON DELETE CASCADE
);

-- Indexes for faster dependency lookups
CREATE INDEX IF NOT EXISTS idx_task_dependencies_task_uuid ON task_dependencies(task_uuid);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on_uuid ON task_dependencies(depends_on_uuid);
