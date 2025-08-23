package models

import (
	"encoding/json"
	"time"
)

// Task represents a task item with TaskWarrior-inspired fields
type Task struct {
	ID          int64     `json:"id"`
	UUID        string    `json:"uuid"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, completed, deleted
	Priority    string    `json:"priority,omitempty"` // A-Z or empty
	Project     string    `json:"project,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Due         *time.Time `json:"due,omitempty"`
	Entry       time.Time `json:"entry"`
	Modified    time.Time `json:"modified"`
	End         *time.Time `json:"end,omitempty"` // completion time
	Start       *time.Time `json:"start,omitempty"` // when task was started
	Annotations []string  `json:"annotations,omitempty"`
}

// Movie represents a movie in the watch queue
type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Year        int       `json:"year,omitempty"`
	Status      string    `json:"status"` // queued, watched, removed
	Rating      float64   `json:"rating,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Added       time.Time `json:"added"`
	Watched     *time.Time `json:"watched,omitempty"`
}

// TVShow represents a TV show in the watch queue
type TVShow struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Season       int       `json:"season,omitempty"`
	Episode      int       `json:"episode,omitempty"`
	Status       string    `json:"status"` // queued, watching, watched, removed
	Rating       float64   `json:"rating,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	Added        time.Time `json:"added"`
	LastWatched  *time.Time `json:"last_watched,omitempty"`
}

// Book represents a book in the reading list
type Book struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Author       string    `json:"author,omitempty"`
	Status       string    `json:"status"` // queued, reading, finished, removed
	Progress     int       `json:"progress"` // percentage 0-100
	Pages        int       `json:"pages,omitempty"`
	Rating       float64   `json:"rating,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	Added        time.Time `json:"added"`
	Started      *time.Time `json:"started,omitempty"`
	Finished     *time.Time `json:"finished,omitempty"`
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