package models

import (
	"encoding/json"
	"fmt"
	"net/url"
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

// RRule represents a recurrence rule (RFC 5545).
// Example: "FREQ=DAILY;INTERVAL=1" or "FREQ=WEEKLY;BYDAY=MO,WE,FR".
type RRule string

// Model defines the common interface that all domain models must implement
type Model interface {
	GetID() int64             // GetID returns the primary key identifier
	SetID(id int64)           // SetID sets the primary key identifier
	GetTableName() string     // GetTableName returns the database table name for this model
	GetCreatedAt() time.Time  // GetCreatedAt returns when the model was created
	SetCreatedAt(t time.Time) // SetCreatedAt sets when the model was created
	GetUpdatedAt() time.Time  // GetUpdatedAt returns when the model was last updated
	SetUpdatedAt(t time.Time) // SetUpdatedAt sets when the model was last updated
}

// Stateful represents entities with status management behavior
//
// Implemented by: [Book], [Movie], [TVShow], [Task]
type Stateful interface {
	GetStatus() string
	ValidStatuses() []string
}

// Queueable represents media that can be queued for later consumption
//
// Implemented by: [Book], [Movie], [TVShow]
type Queueable interface {
	Stateful
	IsQueued() bool
}

// Completable represents media that can be marked as completed/finished/watched. It tracks completion timestamps for media consumption.
//
// Implemented by: [Book] (finished), [Movie] (watched), [TVShow] (watched)
type Completable interface {
	Stateful
	IsCompleted() bool
	GetCompletionTime() *time.Time
}

// Progressable represents media with measurable progress tracking
//
// Implemented by: [Book] (percentage-based reading progress)
type Progressable interface {
	Completable
	GetProgress() int
	SetProgress(progress int) error
}

// Compile-time interface checks
var (
	_ Stateful     = (*Task)(nil)
	_ Stateful     = (*Book)(nil)
	_ Stateful     = (*Movie)(nil)
	_ Stateful     = (*TVShow)(nil)
	_ Queueable    = (*Book)(nil)
	_ Queueable    = (*Movie)(nil)
	_ Queueable    = (*TVShow)(nil)
	_ Completable  = (*Book)(nil)
	_ Completable  = (*Movie)(nil)
	_ Completable  = (*TVShow)(nil)
	_ Progressable = (*Book)(nil)
)

// Task represents a task item with TaskWarrior-inspired fields
type Task struct {
	ID          int64      `json:"id"`
	UUID        string     `json:"uuid"`
	Description string     `json:"description"`
	Status      string     `json:"status"`             // pending, completed, deleted
	Priority    string     `json:"priority,omitempty"` // A-Z or empty
	Project     string     `json:"project,omitempty"`
	Context     string     `json:"context,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Due         *time.Time `json:"due,omitempty"`
	Entry       time.Time  `json:"entry"`
	Modified    time.Time  `json:"modified"`
	End         *time.Time `json:"end,omitempty"`   // Completion time
	Start       *time.Time `json:"start,omitempty"` // When the task was started
	Annotations []string   `json:"annotations,omitempty"`
	Recur       RRule      `json:"recur,omitempty"`
	Until       *time.Time `json:"until,omitempty"`       // End date for recurrence
	ParentUUID  *string    `json:"parent_uuid,omitempty"` // ID of parent/template task
	DependsOn   []string   `json:"depends_on,omitempty"`  // IDs of tasks this task depends on
}

// Movie represents a movie in the watch queue
type Movie struct {
	ID      int64      `json:"id"`
	Title   string     `json:"title"`
	Year    int        `json:"year,omitempty"`
	Status  string     `json:"status"` // queued, watched, removed
	Rating  float64    `json:"rating,omitempty"`
	Notes   string     `json:"notes,omitempty"`
	Added   time.Time  `json:"added"`
	Watched *time.Time `json:"watched,omitempty"`
}

// TVShow represents a TV show in the watch queue
type TVShow struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Season      int        `json:"season,omitempty"`
	Episode     int        `json:"episode,omitempty"`
	Status      string     `json:"status"` // queued, watching, watched, removed
	Rating      float64    `json:"rating,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	Added       time.Time  `json:"added"`
	LastWatched *time.Time `json:"last_watched,omitempty"`
}

// Book represents a book in the reading list
type Book struct {
	ID       int64      `json:"id"`
	Title    string     `json:"title"`
	Author   string     `json:"author,omitempty"`
	Status   string     `json:"status"`   // queued, reading, finished, removed
	Progress int        `json:"progress"` // percentage 0-100
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

// Article represents a parsed article from a web URL
type Article struct {
	ID           int64     `json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Author       string    `json:"author,omitempty"`
	Date         string    `json:"date,omitempty"`
	MarkdownPath string    `json:"markdown_path"`
	HTMLPath     string    `json:"html_path"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
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
func (t *Task) IsCompleted() bool { return t.Status == "completed" }

// IsPending returns true if the task is pending
func (t *Task) IsPending() bool { return t.Status == "pending" }

// IsDeleted returns true if the task is deleted
func (t *Task) IsDeleted() bool { return t.Status == "deleted" }

// HasPriority returns true if the task has a priority set
func (t *Task) HasPriority() bool { return t.Priority != "" }

func (t *Task) IsTodo() bool       { return t.Status == StatusTodo }
func (t *Task) IsInProgress() bool { return t.Status == StatusInProgress }
func (t *Task) IsBlocked() bool    { return t.Status == StatusBlocked }
func (t *Task) IsDone() bool       { return t.Status == StatusDone }
func (t *Task) IsAbandoned() bool  { return t.Status == StatusAbandoned }

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

// GetPriorityWeight returns a numeric weight for sorting priorities. A higher number = higher priority
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

// IsStarted returns true if the task has a start time set.
func (t *Task) IsStarted() bool { return t.Start != nil }

// IsOverdue returns true if the task is overdue.
func (t *Task) IsOverdue(now time.Time) bool {
	return t.Due != nil && now.After(*t.Due) && !t.IsCompleted()
}

// HasDueDate returns true if the task has a due date set.
func (t *Task) HasDueDate() bool { return t.Due != nil }

// IsRecurring returns true if the task has recurrence defined.
func (t *Task) IsRecurring() bool { return t.Recur != "" }

// IsRecurExpired checks if the recurrence has an end (until) date and is past it.
func (t *Task) IsRecurExpired(now time.Time) bool {
	return t.Until != nil && now.After(*t.Until)
}

// HasDependencies returns true if the task depends on other tasks.
func (t *Task) HasDependencies() bool { return len(t.DependsOn) > 0 }

// Blocks checks if this task blocks another given task.
func (t *Task) Blocks(other *Task) bool {
	return slices.Contains(other.DependsOn, t.UUID)
}

// Urgency computes a score based on priority, due date, and tags.
// This can be expanded later with weights.
func (t *Task) Urgency(now time.Time) float64 {
	score := 0.0
	if t.Priority != "" {
		score += 1.0
	}
	if t.IsOverdue(now) {
		score += 2.0
	}
	if len(t.Tags) > 0 {
		score += 0.5
	}
	return score
}

// GetStatus returns the current status of the task
func (t *Task) GetStatus() string { return t.Status }

// ValidStatuses returns all valid status values for a task
func (t *Task) ValidStatuses() []string {
	return []string{
		StatusTodo, StatusInProgress, StatusBlocked, StatusDone, StatusAbandoned,
		StatusPending, StatusCompleted, StatusDeleted,
	}
}

// IsWatched returns true if the movie has been watched
func (m *Movie) IsWatched() bool { return m.Status == "watched" }

// IsQueued returns true if the movie is in the queue
func (m *Movie) IsQueued() bool { return m.Status == "queued" }

// GetStatus returns the current status of the movie
func (m *Movie) GetStatus() string { return m.Status }

// ValidStatuses returns all valid status values for a movie
func (m *Movie) ValidStatuses() []string { return []string{"queued", "watched", "removed"} }

// IsCompleted returns true if the movie has been watched
func (m *Movie) IsCompleted() bool { return m.Status == "watched" }

// GetCompletionTime returns when the movie was watched
func (m *Movie) GetCompletionTime() *time.Time { return m.Watched }

// IsWatching returns true if the TV show is currently being watched
func (tv *TVShow) IsWatching() bool { return tv.Status == "watching" }

// IsWatched returns true if the TV show has been watched
func (tv *TVShow) IsWatched() bool { return tv.Status == "watched" }

// IsQueued returns true if the TV show is in the queue
func (tv *TVShow) IsQueued() bool { return tv.Status == "queued" }

// GetStatus returns the current status of the TV show
func (tv *TVShow) GetStatus() string { return tv.Status }

// ValidStatuses returns all valid status values for a TV show
func (tv *TVShow) ValidStatuses() []string {
	return []string{"queued", "watching", "watched", "removed"}
}

// IsCompleted returns true if the TV show has been watched
func (tv *TVShow) IsCompleted() bool { return tv.Status == "watched" }

// GetCompletionTime returns when the TV show was last watched
func (tv *TVShow) GetCompletionTime() *time.Time { return tv.LastWatched }

// IsReading returns true if the book is currently being read
func (b *Book) IsReading() bool { return b.Status == "reading" }

// IsFinished returns true if the book has been finished
func (b *Book) IsFinished() bool { return b.Status == "finished" }

// IsQueued returns true if the book is in the queue
func (b *Book) IsQueued() bool { return b.Status == "queued" }

// ProgressPercent returns the reading progress as a percentage
func (b *Book) ProgressPercent() int { return b.Progress }

// GetStatus returns the current status of the book
func (b *Book) GetStatus() string { return b.Status }

// ValidStatuses returns all valid status values for a book
func (b *Book) ValidStatuses() []string { return []string{"queued", "reading", "finished", "removed"} }

// IsCompleted returns true if the book has been finished
func (b *Book) IsCompleted() bool { return b.Status == "finished" }

// GetCompletionTime returns when the book was finished
func (b *Book) GetCompletionTime() *time.Time { return b.Finished }

// GetProgress returns the reading progress percentage (0-100)
func (b *Book) GetProgress() int { return b.Progress }

// SetProgress sets the reading progress percentage (0-100)
func (b *Book) SetProgress(progress int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100, got %d", progress)
	}
	b.Progress = progress
	return nil
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
func (a *Album) HasRating() bool { return a.Rating > 0 }

// IsValidRating returns true if the rating is between 1 and 5
func (a *Album) IsValidRating() bool { return a.Rating >= 1 && a.Rating <= 5 }

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

func (a *Article) GetID() int64                { return a.ID }
func (a *Article) SetID(id int64)              { a.ID = id }
func (a *Article) GetTableName() string        { return "articles" }
func (a *Article) GetCreatedAt() time.Time     { return a.Created }
func (a *Article) SetCreatedAt(time time.Time) { a.Created = time }
func (a *Article) GetUpdatedAt() time.Time     { return a.Modified }
func (a *Article) SetUpdatedAt(time time.Time) { a.Modified = time }

// IsValidURL returns true if the article has parseable URL
func (a *Article) IsValidURL() bool {
	_, err := url.ParseRequestURI(a.URL)
	return err == nil
}

// HasAuthor returns true if the article has an author
func (a *Article) HasAuthor() bool { return a.Author != "" }

// HasDate returns true if the article has a date
func (a *Article) HasDate() bool { return a.Date != "" }
