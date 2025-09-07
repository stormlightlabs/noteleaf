-- Add context field to tasks table
ALTER TABLE tasks ADD COLUMN context TEXT;

-- Add index for context queries
CREATE INDEX IF NOT EXISTS idx_tasks_context ON tasks(context);