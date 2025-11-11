package repo

const (
	noteColumns     = "id, title, content, tags, archived, created, modified, file_path, leaflet_rkey, leaflet_cid, published_at, is_draft"
	queryNoteByID   = "SELECT " + noteColumns + " FROM notes WHERE id = ?"
	queryNoteInsert = `INSERT INTO notes (title, content, tags, archived, created, modified, file_path, leaflet_rkey, leaflet_cid, published_at, is_draft) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	queryNoteUpdate = `UPDATE notes SET title = ?, content = ?, tags = ?, archived = ?, modified = ?, file_path = ?, leaflet_rkey = ?, leaflet_cid = ?, published_at = ?, is_draft = ? WHERE id = ?`
	queryNoteDelete = "DELETE FROM notes WHERE id = ?"
	queryNotesList  = "SELECT " + noteColumns + " FROM notes"
)
const (
	articleColumns     = "id, url, title, author, date, markdown_path, html_path, created, modified"
	queryArticleByID   = "SELECT " + articleColumns + " FROM articles WHERE id = ?"
	queryArticleByURL  = "SELECT " + articleColumns + " FROM articles WHERE url = ?"
	queryArticleInsert = `INSERT INTO articles (url, title, author, date, markdown_path, html_path, created, modified) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	queryArticleUpdate = `UPDATE articles SET title = ?, author = ?, date = ?, markdown_path = ?, html_path = ?, modified = ? WHERE id = ?`
	queryArticleDelete = "DELETE FROM articles WHERE id = ?"
	queryArticlesList  = "SELECT " + articleColumns + " FROM articles"
	queryArticlesCount = "SELECT COUNT(*) FROM articles"
)

const (
	taskColumns     = "id, uuid, description, status, priority, project, context, tags, due, wait, scheduled, entry, modified, end, start, annotations, recur, until, parent_uuid"
	queryTaskByID   = "SELECT " + taskColumns + " FROM tasks WHERE id = ?"
	queryTaskByUUID = "SELECT " + taskColumns + " FROM tasks WHERE uuid = ?"
	queryTaskInsert = `
		INSERT INTO tasks (
			uuid, description, status, priority, project, context,
			tags, due, wait, scheduled, entry, modified, end, start, annotations,
			recur, until, parent_uuid
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	queryTaskUpdate = `
		UPDATE tasks SET
			uuid = ?, description = ?, status = ?, priority = ?, project = ?, context = ?,
			tags = ?, due = ?, wait = ?, scheduled = ?, modified = ?, end = ?, start = ?, annotations = ?,
			recur = ?, until = ?, parent_uuid = ?
		WHERE id = ?`
	queryTaskDelete = "DELETE FROM tasks WHERE id = ?"
	queryTasksList  = "SELECT " + taskColumns + " FROM tasks"
)

type scanner interface {
	Scan(dest ...any) error
}
