package models

import "time"

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
