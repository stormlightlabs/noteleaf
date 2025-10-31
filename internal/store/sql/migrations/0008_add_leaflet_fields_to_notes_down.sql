-- Remove leaflet fields and indexes
DROP INDEX IF EXISTS idx_notes_is_draft;
DROP INDEX IF EXISTS idx_notes_leaflet_rkey;
ALTER TABLE notes DROP COLUMN is_draft;
ALTER TABLE notes DROP COLUMN published_at;
ALTER TABLE notes DROP COLUMN leaflet_cid;
ALTER TABLE notes DROP COLUMN leaflet_rkey;
