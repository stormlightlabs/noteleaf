-- Drop time tracking table
DROP INDEX IF EXISTS idx_time_entries_task_id;
DROP INDEX IF EXISTS idx_time_entries_start_time;
DROP INDEX IF EXISTS idx_time_entries_end_time;
DROP TABLE IF EXISTS time_entries;