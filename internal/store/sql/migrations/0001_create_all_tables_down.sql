-- Drop books table and indexes
DROP INDEX IF EXISTS idx_books_title;
DROP INDEX IF EXISTS idx_books_author;
DROP INDEX IF EXISTS idx_books_status;
DROP TABLE IF EXISTS books;

-- Drop TV shows table and indexes
DROP INDEX IF EXISTS idx_tv_shows_season_episode;
DROP INDEX IF EXISTS idx_tv_shows_title;
DROP INDEX IF EXISTS idx_tv_shows_status;
DROP TABLE IF EXISTS tv_shows;

-- Drop movies table and indexes
DROP INDEX IF EXISTS idx_movies_title;
DROP INDEX IF EXISTS idx_movies_status;
DROP TABLE IF EXISTS movies;

-- Drop tasks table and indexes
DROP INDEX IF EXISTS idx_tasks_uuid;
DROP INDEX IF EXISTS idx_tasks_due;
DROP INDEX IF EXISTS idx_tasks_project;
DROP INDEX IF EXISTS idx_tasks_status;
DROP TABLE IF EXISTS tasks;