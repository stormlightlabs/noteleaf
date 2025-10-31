-- Add leaflet publication fields to notes table
ALTER TABLE notes ADD COLUMN leaflet_rkey TEXT;
ALTER TABLE notes ADD COLUMN leaflet_cid TEXT;
ALTER TABLE notes ADD COLUMN published_at DATETIME;
ALTER TABLE notes ADD COLUMN is_draft INTEGER DEFAULT 0;

-- Add index for leaflet record key lookups
CREATE INDEX IF NOT EXISTS idx_notes_leaflet_rkey ON notes(leaflet_rkey);

-- Add index for published vs draft queries
CREATE INDEX IF NOT EXISTS idx_notes_is_draft ON notes(is_draft);
