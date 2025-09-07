package models

import (
	"encoding/json"
	"slices"
	"time"
)

type TaskStatus string
type TaskPriority string
type TaskWeight int

// TODO: Use [TaskStatus]
const (
	StatusTodo       = "todo"
	StatusInProgress = "in-progress"
	StatusBlocked    = "blocked"
	StatusDone       = "done"
	StatusAbandoned  = "abandoned"
	StatusPending    = "pending"
	StatusCompleted  = "completed"
	StatusDeleted    = "deleted"
)

// TODO: Use [TaskPriority]
const (
	PriorityHigh   = "High"
	PriorityMedium = "Medium"
	PriorityLow    = "Low"
)

// TODO: Use [TaskWeight]
const (
	PriorityNumericMin = 1
	PriorityNumericMax = 5
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
	Context  string     `json:"context,omitempty"`
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

// Album represents a music album
type Album struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Artist          string    `json:"artist"`
	Genre           string    `json:"genre,omitempty"`
	ReleaseYear     int       `json:"release_year,omitempty"`
	Tracks          []string  `json:"tracks,omitempty"`
	DurationSeconds int       `json:"duration_seconds,omitempty"`
	AlbumArtPath    string    `json:"album_art_path,omitempty"`
	Rating          int       `json:"rating,omitempty"`
	Created         time.Time `json:"created"`
	Modified        time.Time `json:"modified"`
}

// TimeEntry represents a time tracking entry for a task
type TimeEntry struct {
	ID              int64      `json:"id"`
	TaskID          int64      `json:"task_id"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	DurationSeconds int64      `json:"duration_seconds,omitempty"`
	Description     string     `json:"description,omitempty"`
	Created         time.Time  `json:"created"`
	Modified        time.Time  `json:"modified"`
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

// New status tracking methods
func (t *Task) IsTodo() bool {
	return t.Status == StatusTodo
}

func (t *Task) IsInProgress() bool {
	return t.Status == StatusInProgress
}

func (t *Task) IsBlocked() bool {
	return t.Status == StatusBlocked
}

func (t *Task) IsDone() bool {
	return t.Status == StatusDone
}

func (t *Task) IsAbandoned() bool {
	return t.Status == StatusAbandoned
}

// IsValidStatus returns true if the status is one of the defined valid statuses
func (t *Task) IsValidStatus() bool {
	validStatuses := []string{
		StatusTodo, StatusInProgress, StatusBlocked, StatusDone, StatusAbandoned,
		StatusPending, StatusCompleted, StatusDeleted, // legacy support
	}
	return slices.Contains(validStatuses, t.Status)
}

// IsValidPriority returns true if the priority is valid (text-based or numeric string)
func (t *Task) IsValidPriority() bool {
	if t.Priority == "" {
		return true
	}

	textPriorities := []string{PriorityHigh, PriorityMedium, PriorityLow}
	if slices.Contains(textPriorities, t.Priority) {
		return true
	}

	if len(t.Priority) == 1 && t.Priority >= "A" && t.Priority <= "Z" {
		return true
	}

	switch t.Priority {
	case "1", "2", "3", "4", "5":
		return true
	}

	return false
}

// GetPriorityWeight returns a numeric weight for sorting priorities
//
//	Higher numbers = higher priority
func (t *Task) GetPriorityWeight() int {
	switch t.Priority {
	case PriorityHigh, "5":
		return 5
	case PriorityMedium, "4":
		return 4
	case PriorityLow, "3":
		return 3
	case "2":
		return 2
	case "1":
		return 1
	case "A":
		return 26
	case "B":
		return 25
	case "C":
		return 24
	default:
		if len(t.Priority) == 1 && t.Priority >= "A" && t.Priority <= "Z" {
			return int('Z' - t.Priority[0] + 1)
		}
		return 0
	}
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

// MarshalTracks converts tracks slice to JSON string for database storage
func (a *Album) MarshalTracks() (string, error) {
	if len(a.Tracks) == 0 {
		return "", nil
	}
	data, err := json.Marshal(a.Tracks)
	return string(data), err
}

// UnmarshalTracks converts JSON string from database to tracks slice
func (a *Album) UnmarshalTracks(data string) error {
	if data == "" {
		a.Tracks = nil
		return nil
	}
	return json.Unmarshal([]byte(data), &a.Tracks)
}

// HasRating returns true if the album has a rating set
func (a *Album) HasRating() bool {
	return a.Rating > 0
}

// IsValidRating returns true if the rating is between 1 and 5
func (a *Album) IsValidRating() bool {
	return a.Rating >= 1 && a.Rating <= 5
}

func (a *Album) GetID() int64                { return a.ID }
func (a *Album) SetID(id int64)              { a.ID = id }
func (a *Album) GetTableName() string        { return "albums" }
func (a *Album) GetCreatedAt() time.Time     { return a.Created }
func (a *Album) SetCreatedAt(time time.Time) { a.Created = time }
func (a *Album) GetUpdatedAt() time.Time     { return a.Modified }
func (a *Album) SetUpdatedAt(time time.Time) { a.Modified = time }

// IsActive returns true if the time entry is currently active (not stopped)
func (te *TimeEntry) IsActive() bool {
	return te.EndTime == nil
}

// Stop stops the time entry and calculates duration
func (te *TimeEntry) Stop() {
	now := time.Now()
	te.EndTime = &now
	te.DurationSeconds = int64(now.Sub(te.StartTime).Seconds())
	te.Modified = now
}

// GetDuration returns the duration of the time entry
func (te *TimeEntry) GetDuration() time.Duration {
	if te.EndTime != nil {
		return time.Duration(te.DurationSeconds) * time.Second
	}
	return time.Since(te.StartTime)
}

func (te *TimeEntry) GetID() int64                { return te.ID }
func (te *TimeEntry) SetID(id int64)              { te.ID = id }
func (te *TimeEntry) GetTableName() string        { return "time_entries" }
func (te *TimeEntry) GetCreatedAt() time.Time     { return te.Created }
func (te *TimeEntry) SetCreatedAt(time time.Time) { te.Created = time }
func (te *TimeEntry) GetUpdatedAt() time.Time     { return te.Modified }
func (te *TimeEntry) SetUpdatedAt(time time.Time) { te.Modified = time }
