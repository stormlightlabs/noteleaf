package repo

import (
	"database/sql"
)

// Repositories provides access to all resource repositories
type Repositories struct {
	Tasks  *TaskRepository
	Movies *MovieRepository
	TV     *TVRepository
	Books  *BookRepository
}

// NewRepositories creates a new set of repositories
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Tasks:  NewTaskRepository(db),
		Movies: NewMovieRepository(db),
		TV:     NewTVRepository(db),
		Books:  NewBookRepository(db),
	}
}
