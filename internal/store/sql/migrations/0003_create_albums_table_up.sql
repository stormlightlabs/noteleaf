CREATE TABLE IF NOT EXISTS albums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    genre TEXT,
    release_year INTEGER,
    tracks TEXT, -- JSON array of track names
    duration_seconds INTEGER,
    album_art_path TEXT,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_albums_modified
    AFTER UPDATE ON albums
    FOR EACH ROW
    BEGIN
        UPDATE albums SET modified = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
