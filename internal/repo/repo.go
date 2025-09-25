package repo

import (
	"database/sql"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

type Repository interface {
	Validate(models.Model) error
}

// Repositories provides access to all resource repositories
type Repositories struct {
	Tasks       *TaskRepository
	Movies      *MovieRepository
	TV          *TVRepository
	Books       *BookRepository
	Notes       *NoteRepository
	TimeEntries *TimeEntryRepository
	Articles    *ArticleRepository
}

// NewRepositories creates a new set of [Repositories]
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Tasks:       NewTaskRepository(db),
		Movies:      NewMovieRepository(db),
		TV:          NewTVRepository(db),
		Books:       NewBookRepository(db),
		Notes:       NewNoteRepository(db),
		TimeEntries: NewTimeEntryRepository(db),
		Articles:    NewArticleRepository(db),
	}
}
