-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    priority TEXT,
    project TEXT,
    tags TEXT, -- JSON array
    due DATETIME,
    entry DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified DATETIME DEFAULT CURRENT_TIMESTAMP,
    end DATETIME,
    start DATETIME,
    annotations TEXT -- JSON array
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project);
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due);
CREATE INDEX IF NOT EXISTS idx_tasks_uuid ON tasks(uuid);

-- Movies table
CREATE TABLE IF NOT EXISTS movies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    year INTEGER,
    status TEXT DEFAULT 'queued',
    rating REAL,
    notes TEXT,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    watched DATETIME
);

CREATE INDEX IF NOT EXISTS idx_movies_status ON movies(status);
CREATE INDEX IF NOT EXISTS idx_movies_title ON movies(title);

-- TV Shows table
CREATE TABLE IF NOT EXISTS tv_shows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    season INTEGER,
    episode INTEGER,
    status TEXT DEFAULT 'queued',
    rating REAL,
    notes TEXT,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_watched DATETIME
);

CREATE INDEX IF NOT EXISTS idx_tv_shows_status ON tv_shows(status);
CREATE INDEX IF NOT EXISTS idx_tv_shows_title ON tv_shows(title);
CREATE INDEX IF NOT EXISTS idx_tv_shows_season_episode ON tv_shows(title, season, episode);

-- Books table
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT,
    status TEXT DEFAULT 'queued',
    progress INTEGER DEFAULT 0,
    pages INTEGER,
    rating REAL,
    notes TEXT,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    started DATETIME,
    finished DATETIME
);

CREATE INDEX IF NOT EXISTS idx_books_status ON books(status);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);