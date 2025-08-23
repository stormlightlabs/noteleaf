package models

import (
	"encoding/json"
	"time"
)

// Model defines the common interface that all domain models must implement
type Model interface {
	// GetID returns the primary key identifier
	GetID() int64

	// SetID sets the primary key identifier
	SetID(id int64)

	// GetTableName returns the database table name for this model
	GetTableName() string

	// GetCreatedAt returns when the model was created
	GetCreatedAt() time.Time

	// SetCreatedAt sets when the model was created
	SetCreatedAt(t time.Time)

	// GetUpdatedAt returns when the model was last updated
	GetUpdatedAt() time.Time

	// SetUpdatedAt sets when the model was last updated
	SetUpdatedAt(t time.Time)
}

// Task represents a task item with TaskWarrior-inspired fields
type Task struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Description string `json:"description"`
	// pending, completed, deleted
	Status string `json:"status"`
	// A-Z or empty
	Priority string     `json:"priority,omitempty"`
	Project  string     `json:"project,omitempty"`
	Tags     []string   `json:"tags,omitempty"`
	Due      *time.Time `json:"due,omitempty"`
	Entry    time.Time  `json:"entry"`
	Modified time.Time  `json:"modified"`
	// completion time
	End *time.Time `json:"end,omitempty"`
	// when task was started
	Start       *time.Time `json:"start,omitempty"`
	Annotations []string   `json:"annotations,omitempty"`
}

// Movie represents a movie in the watch queue
type Movie struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Year  int    `json:"year,omitempty"`
	// queued, watched, removed
	Status  string     `json:"status"`
	Rating  float64    `json:"rating,omitempty"`
	Notes   string     `json:"notes,omitempty"`
	Added   time.Time  `json:"added"`
	Watched *time.Time `json:"watched,omitempty"`
}

// TVShow represents a TV show in the watch queue
type TVShow struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Season  int    `json:"season,omitempty"`
	Episode int    `json:"episode,omitempty"`
	// queued, watching, watched, removed
	Status      string     `json:"status"`
	Rating      float64    `json:"rating,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	Added       time.Time  `json:"added"`
	LastWatched *time.Time `json:"last_watched,omitempty"`
}

// Book represents a book in the reading list
type Book struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author,omitempty"`
	// queued, reading, finished, removed
	Status string `json:"status"`
	// percentage 0-100
	Progress int        `json:"progress"`
	Pages    int        `json:"pages,omitempty"`
	Rating   float64    `json:"rating,omitempty"`
	Notes    string     `json:"notes,omitempty"`
	Added    time.Time  `json:"added"`
	Started  *time.Time `json:"started,omitempty"`
	Finished *time.Time `json:"finished,omitempty"`
}

// Note represents a markdown note
type Note struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Tags     []string  `json:"tags,omitempty"`
	Archived bool      `json:"archived"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	FilePath string    `json:"file_path,omitempty"`
}

// MarshalTags converts tags slice to JSON string for database storage
func (t *Task) MarshalTags() (string, error) {
	if len(t.Tags) == 0 {
		return "", nil
	}
	data, err := json.Marshal(t.Tags)
	return string(data), err
}

// UnmarshalTags converts JSON string from database to tags slice
func (t *Task) UnmarshalTags(data string) error {
	if data == "" {
		t.Tags = nil
		return nil
	}
	return json.Unmarshal([]byte(data), &t.Tags)
}

// MarshalAnnotations converts annotations slice to JSON string for database storage
func (t *Task) MarshalAnnotations() (string, error) {
	if len(t.Annotations) == 0 {
		return "", nil
	}
	data, err := json.Marshal(t.Annotations)
	return string(data), err
}

// UnmarshalAnnotations converts JSON string from database to annotations slice
func (t *Task) UnmarshalAnnotations(data string) error {
	if data == "" {
		t.Annotations = nil
		return nil
	}
	return json.Unmarshal([]byte(data), &t.Annotations)
}

// IsCompleted returns true if the task is marked as completed
func (t *Task) IsCompleted() bool {
	return t.Status == "completed"
}

// IsPending returns true if the task is pending
func (t *Task) IsPending() bool {
	return t.Status == "pending"
}

// IsDeleted returns true if the task is deleted
func (t *Task) IsDeleted() bool {
	return t.Status == "deleted"
}

// HasPriority returns true if the task has a priority set
func (t *Task) HasPriority() bool {
	return t.Priority != ""
}

// IsWatched returns true if the movie has been watched
func (m *Movie) IsWatched() bool {
	return m.Status == "watched"
}

// IsQueued returns true if the movie is in the queue
func (m *Movie) IsQueued() bool {
	return m.Status == "queued"
}

// IsWatching returns true if the TV show is currently being watched
func (tv *TVShow) IsWatching() bool {
	return tv.Status == "watching"
}

// IsWatched returns true if the TV show has been watched
func (tv *TVShow) IsWatched() bool {
	return tv.Status == "watched"
}

// IsQueued returns true if the TV show is in the queue
func (tv *TVShow) IsQueued() bool {
	return tv.Status == "queued"
}

// IsReading returns true if the book is currently being read
func (b *Book) IsReading() bool {
	return b.Status == "reading"
}

// IsFinished returns true if the book has been finished
func (b *Book) IsFinished() bool {
	return b.Status == "finished"
}

// IsQueued returns true if the book is in the queue
func (b *Book) IsQueued() bool {
	return b.Status == "queued"
}

// ProgressPercent returns the reading progress as a percentage
func (b *Book) ProgressPercent() int {
	return b.Progress
}

func (t *Task) GetID() int64                { return t.ID }
func (t *Task) SetID(id int64)              { t.ID = id }
func (t *Task) GetTableName() string        { return "tasks" }
func (t *Task) GetCreatedAt() time.Time     { return t.Entry }
func (t *Task) SetCreatedAt(time time.Time) { t.Entry = time }
func (t *Task) GetUpdatedAt() time.Time     { return t.Modified }
func (t *Task) SetUpdatedAt(time time.Time) { t.Modified = time }

func (m *Movie) GetID() int64                { return m.ID }
func (m *Movie) SetID(id int64)              { m.ID = id }
func (m *Movie) GetTableName() string        { return "movies" }
func (m *Movie) GetCreatedAt() time.Time     { return m.Added }
func (m *Movie) SetCreatedAt(time time.Time) { m.Added = time }
func (m *Movie) GetUpdatedAt() time.Time     { return m.Added }
func (m *Movie) SetUpdatedAt(time time.Time) { m.Added = time }

func (tv *TVShow) GetID() int64                { return tv.ID }
func (tv *TVShow) SetID(id int64)              { tv.ID = id }
func (tv *TVShow) GetTableName() string        { return "tv_shows" }
func (tv *TVShow) GetCreatedAt() time.Time     { return tv.Added }
func (tv *TVShow) SetCreatedAt(time time.Time) { tv.Added = time }
func (tv *TVShow) GetUpdatedAt() time.Time     { return tv.Added }
func (tv *TVShow) SetUpdatedAt(time time.Time) { tv.Added = time }

func (b *Book) GetID() int64                { return b.ID }
func (b *Book) SetID(id int64)              { b.ID = id }
func (b *Book) GetTableName() string        { return "books" }
func (b *Book) GetCreatedAt() time.Time     { return b.Added }
func (b *Book) SetCreatedAt(time time.Time) { b.Added = time }
func (b *Book) GetUpdatedAt() time.Time     { return b.Added }
func (b *Book) SetUpdatedAt(time time.Time) { b.Added = time }

// MarshalTags converts tags slice to JSON string for database storage
func (n *Note) MarshalTags() (string, error) {
	if len(n.Tags) == 0 {
		return "", nil
	}
	data, err := json.Marshal(n.Tags)
	return string(data), err
}

// UnmarshalTags converts JSON string from database to tags slice
func (n *Note) UnmarshalTags(data string) error {
	if data == "" {
		n.Tags = nil
		return nil
	}
	return json.Unmarshal([]byte(data), &n.Tags)
}

// IsArchived returns true if the note is archived
func (n *Note) IsArchived() bool {
	return n.Archived
}

func (n *Note) GetID() int64                { return n.ID }
func (n *Note) SetID(id int64)              { n.ID = id }
func (n *Note) GetTableName() string        { return "notes" }
func (n *Note) GetCreatedAt() time.Time     { return n.Created }
func (n *Note) SetCreatedAt(time time.Time) { n.Created = time }
func (n *Note) GetUpdatedAt() time.Time     { return n.Modified }
func (n *Note) SetUpdatedAt(time time.Time) { n.Modified = time }
