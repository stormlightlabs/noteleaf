-- Remove recurrence fields
ALTER TABLE tasks DROP COLUMN recur;
ALTER TABLE tasks DROP COLUMN until;
ALTER TABLE tasks DROP COLUMN parent_uuid;

-- Drop dependencies table
DROP TABLE IF EXISTS task_dependencies;
