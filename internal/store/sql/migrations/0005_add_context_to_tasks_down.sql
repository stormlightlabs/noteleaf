-- Remove context field and index
DROP INDEX IF EXISTS idx_tasks_context;
ALTER TABLE tasks DROP COLUMN context;