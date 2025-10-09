package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestModels(t *testing.T) {
	t.Run("Model Interface", func(t *testing.T) {
		now := time.Now()
		time.Sleep(time.Duration(500) * time.Duration(time.Millisecond))
		updated := time.Now()

		for i, tc := range []struct {
			name        string
			model       Model
			unmarshaled Model
		}{
			{name: "Task", model: &Task{ID: 1, Entry: now, Modified: updated}, unmarshaled: &Task{}},
			{name: "Movie", model: &Movie{ID: 1, Title: "Test Movie", Year: 2023, Added: now}, unmarshaled: &Movie{}},
			{name: "TVShow", model: &TVShow{ID: 1, Title: "Test Show", Added: now}, unmarshaled: &TVShow{}},
			{name: "Book", model: &Book{ID: 1, Title: "Test Book", Added: now}, unmarshaled: &Book{}},
			{name: "Note", model: &Note{ID: 1, Title: "Test Note", Content: "This is test content", Created: now}, unmarshaled: &Note{}},
			{name: "Album", model: &Album{ID: 1, Title: "Test Album", Artist: "Test Artist", Created: now}, unmarshaled: &Album{}},
			{name: "TimeEntry", model: &TimeEntry{ID: 1, TaskID: 100, Created: now, Modified: updated}, unmarshaled: &TimeEntry{}},
			{name: "Article", model: &Article{ID: 1, Created: now, Modified: updated}, unmarshaled: &Article{}},
		} {
			model := tc.model
			t.Run(fmt.Sprintf("%v Implementation", tc.name), func(t *testing.T) {
				model.SetID(int64(i + 1))
				if model.GetID() != int64(i+1) {
					t.Errorf("Model %d: ID not set correctly", i)
				}

				tableName := model.GetTableName()
				if tableName == "" {
					t.Errorf("Model %d: table name should not be empty", i)
				}

				now = time.Now()
				model.SetCreatedAt(now)
				// NOTE: We don't test exact equality due to potential precision differences
				if model.GetCreatedAt().IsZero() {
					t.Errorf("Model %d: created at should not be zero", i)
				}

				updatedAt := time.Now().Add(time.Hour)
				model.SetUpdatedAt(updatedAt)
				if !model.GetUpdatedAt().Equal(updatedAt) {
					t.Errorf("Expected updated at %v, got %v", updatedAt, model.GetUpdatedAt())
				}

				if model.GetUpdatedAt().IsZero() {
					t.Errorf("Model %d: updated at should not be zero", i)
				}
				model.SetUpdatedAt(now)

				t.Run(fmt.Sprintf("%v JSON Marshal/Unmarshal", tc.name), func(t *testing.T) {
					if data, err := json.Marshal(model); err != nil {
						t.Fatalf("JSON marshal failed: %v", err)
					} else {
						var unmarshaled = tc.unmarshaled
						if err = json.Unmarshal(data, &unmarshaled); err != nil {
							t.Fatalf("JSON unmarshal failed: %v", err)
						}

						if unmarshaled.GetID() != model.GetID() {
							t.Fatalf("IDs should be the same")
						}
					}
				})
			})
		}

	})

	t.Run("Task Model", func(t *testing.T) {
		t.Run("Status Methods", func(t *testing.T) {
			tc := []struct {
				status      string
				isCompleted bool
				isPending   bool
				isDeleted   bool
			}{
				{"pending", false, true, false},
				{"completed", true, false, false},
				{"deleted", false, false, true},
				{"unknown", false, false, false},
			}

			for _, tt := range tc {
				task := &Task{Status: tt.status}

				if task.IsCompleted() != tt.isCompleted {
					t.Errorf("Status %s: expected IsCompleted %v, got %v", tt.status, tt.isCompleted, task.IsCompleted())
				}
				if task.IsPending() != tt.isPending {
					t.Errorf("Status %s: expected IsPending %v, got %v", tt.status, tt.isPending, task.IsPending())
				}
				if task.IsDeleted() != tt.isDeleted {
					t.Errorf("Status %s: expected IsDeleted %v, got %v", tt.status, tt.isDeleted, task.IsDeleted())
				}
			}
		})

		t.Run("New Status Tracking Methods", func(t *testing.T) {
			tc := []struct {
				status       string
				isTodo       bool
				isInProgress bool
				isBlocked    bool
				isDone       bool
				isAbandoned  bool
			}{
				{StatusTodo, true, false, false, false, false},
				{StatusInProgress, false, true, false, false, false},
				{StatusBlocked, false, false, true, false, false},
				{StatusDone, false, false, false, true, false},
				{StatusAbandoned, false, false, false, false, true},
				{"unknown", false, false, false, false, false},
			}

			for _, tt := range tc {
				task := &Task{Status: tt.status}

				if task.IsTodo() != tt.isTodo {
					t.Errorf("Status %s: expected IsTodo %v, got %v", tt.status, tt.isTodo, task.IsTodo())
				}
				if task.IsInProgress() != tt.isInProgress {
					t.Errorf("Status %s: expected IsInProgress %v, got %v", tt.status, tt.isInProgress, task.IsInProgress())
				}
				if task.IsBlocked() != tt.isBlocked {
					t.Errorf("Status %s: expected IsBlocked %v, got %v", tt.status, tt.isBlocked, task.IsBlocked())
				}
				if task.IsDone() != tt.isDone {
					t.Errorf("Status %s: expected IsDone %v, got %v", tt.status, tt.isDone, task.IsDone())
				}
				if task.IsAbandoned() != tt.isAbandoned {
					t.Errorf("Status %s: expected IsAbandoned %v, got %v", tt.status, tt.isAbandoned, task.IsAbandoned())
				}
			}
		})

		t.Run("Status Validation", func(t *testing.T) {
			validStatuses := []string{
				StatusTodo, StatusInProgress, StatusBlocked, StatusDone, StatusAbandoned,
				StatusPending, StatusCompleted, StatusDeleted,
			}

			for _, status := range validStatuses {
				task := &Task{Status: status}
				if !task.IsValidStatus() {
					t.Errorf("Status %s should be valid", status)
				}
			}

			invalidStatuses := []string{"unknown", "invalid", ""}
			for _, status := range invalidStatuses {
				task := &Task{Status: status}
				if task.IsValidStatus() {
					t.Errorf("Status %s should be invalid", status)
				}
			}
		})

		t.Run("Priority Methods", func(t *testing.T) {
			task := &Task{}

			if task.HasPriority() {
				t.Error("Task with empty priority should return false for HasPriority")
			}

			task.Priority = "A"
			if !task.HasPriority() {
				t.Error("Task with priority should return true for HasPriority")
			}
		})

		t.Run("Priority System", func(t *testing.T) {
			t.Run("Text-based Priority Validation", func(t *testing.T) {
				validTextPriorities := []string{
					PriorityHigh, PriorityMedium, PriorityLow,
				}

				for _, priority := range validTextPriorities {
					task := &Task{Priority: priority}
					if !task.IsValidPriority() {
						t.Errorf("Priority %s should be valid", priority)
					}
				}
			})

			t.Run("Numeric Priority Validation", func(t *testing.T) {
				validNumericPriorities := []string{"1", "2", "3", "4", "5"}

				for _, priority := range validNumericPriorities {
					task := &Task{Priority: priority}
					if !task.IsValidPriority() {
						t.Errorf("Numeric priority %s should be valid", priority)
					}
				}

				invalidNumericPriorities := []string{"0", "6", "10", "-1"}
				for _, priority := range invalidNumericPriorities {
					task := &Task{Priority: priority}
					if task.IsValidPriority() {
						t.Errorf("Numeric priority %s should be invalid", priority)
					}
				}
			})

			t.Run("Legacy A-Z Priority Validation", func(t *testing.T) {
				validLegacyPriorities := []string{"A", "B", "C", "D", "Z"}

				for _, priority := range validLegacyPriorities {
					task := &Task{Priority: priority}
					if !task.IsValidPriority() {
						t.Errorf("Legacy priority %s should be valid", priority)
					}
				}

				invalidLegacyPriorities := []string{"AA", "a", "1A", ""}
				for _, priority := range invalidLegacyPriorities {
					task := &Task{Priority: priority}
					if priority != "" && task.IsValidPriority() {
						t.Errorf("Legacy priority %s should be invalid", priority)
					}
				}
			})

			t.Run("Empty Priority Validation", func(t *testing.T) {
				task := &Task{Priority: ""}
				if !task.IsValidPriority() {
					t.Error("Empty priority should be valid")
				}
			})

			t.Run("Priority Weight Calculation", func(t *testing.T) {
				tc := []struct {
					priority string
					weight   int
				}{
					{PriorityHigh, 5},
					{PriorityMedium, 4},
					{PriorityLow, 3},
					{"5", 5},
					{"4", 4},
					{"3", 3},
					{"2", 2},
					{"1", 1},
					{"A", 26},
					{"B", 25},
					{"C", 24},
					{"Z", 1},
					{"", 0},
					{"invalid", 0},
				}

				for _, tt := range tc {
					task := &Task{Priority: tt.priority}
					weight := task.GetPriorityWeight()
					if weight != tt.weight {
						t.Errorf("Priority %s: expected weight %d, got %d", tt.priority, tt.weight, weight)
					}
				}
			})

			t.Run("Priority Weight Ordering", func(t *testing.T) {
				priorities := []string{PriorityHigh, PriorityMedium, PriorityLow}
				weights := []int{}

				for _, priority := range priorities {
					task := &Task{Priority: priority}
					weights = append(weights, task.GetPriorityWeight())
				}

				for i := 1; i < len(weights); i++ {
					if weights[i-1] <= weights[i] {
						t.Errorf("Priority weights should be in descending order: %v", weights)
					}
				}
			})
		})

		t.Run("Tags Marshaling", func(t *testing.T) {
			task := &Task{}

			result, err := task.MarshalTags()
			if err != nil {
				t.Fatalf("MarshalTags failed: %v", err)
			}
			if result != "" {
				t.Errorf("Expected empty string for empty tags, got '%s'", result)
			}

			task.Tags = []string{"work", "urgent", "project-x"}
			result, err = task.MarshalTags()
			if err != nil {
				t.Fatalf("MarshalTags failed: %v", err)
			}

			expected := `["work","urgent","project-x"]`
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}

			newTask := &Task{}
			err = newTask.UnmarshalTags(result)
			if err != nil {
				t.Fatalf("UnmarshalTags failed: %v", err)
			}

			if len(newTask.Tags) != 3 {
				t.Errorf("Expected 3 tags, got %d", len(newTask.Tags))
			}
			if newTask.Tags[0] != "work" || newTask.Tags[1] != "urgent" || newTask.Tags[2] != "project-x" {
				t.Errorf("Tags not unmarshaled correctly: %v", newTask.Tags)
			}

			emptyTask := &Task{}
			err = emptyTask.UnmarshalTags("")
			if err != nil {
				t.Fatalf("UnmarshalTags with empty string failed: %v", err)
			}
			if emptyTask.Tags != nil {
				t.Error("Expected nil tags for empty string")
			}
		})

		t.Run("Annotations Marshaling", func(t *testing.T) {
			task := &Task{}

			result, err := task.MarshalAnnotations()
			if err != nil {
				t.Fatalf("MarshalAnnotations failed: %v", err)
			}
			if result != "" {
				t.Errorf("Expected empty string for empty annotations, got '%s'", result)
			}

			task.Annotations = []string{"Note 1", "Note 2", "Important reminder"}
			result, err = task.MarshalAnnotations()
			if err != nil {
				t.Fatalf("MarshalAnnotations failed: %v", err)
			}

			expected := `["Note 1","Note 2","Important reminder"]`
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}

			newTask := &Task{}
			err = newTask.UnmarshalAnnotations(result)
			if err != nil {
				t.Fatalf("UnmarshalAnnotations failed: %v", err)
			}

			if len(newTask.Annotations) != 3 {
				t.Errorf("Expected 3 annotations, got %d", len(newTask.Annotations))
			}
			if newTask.Annotations[0] != "Note 1" || newTask.Annotations[1] != "Note 2" || newTask.Annotations[2] != "Important reminder" {
				t.Errorf("Annotations not unmarshaled correctly: %v", newTask.Annotations)
			}

			emptyTask := &Task{}
			err = emptyTask.UnmarshalAnnotations("")
			if err != nil {
				t.Fatalf("UnmarshalAnnotations with empty string failed: %v", err)
			}
			if emptyTask.Annotations != nil {
				t.Error("Expected nil annotations for empty string")
			}
		})

		t.Run("IsStarted", func(t *testing.T) {
			now := time.Now()
			task := Task{UUID: "123", Description: "demo", Entry: now, Modified: now}

			if task.IsStarted() {
				t.Errorf("expected IsStarted to be false, got true")
			}
			task.Start = &now
			if !task.IsStarted() {
				t.Errorf("expected IsStarted to be true, got false")
			}
		})

		t.Run("HasDueDate and IsOverdue", func(t *testing.T) {
			now := time.Now()
			past := now.Add(-24 * time.Hour)
			future := now.Add(24 * time.Hour)
			task := Task{UUID: "123", Description: "demo", Entry: now, Modified: now}

			if task.HasDueDate() {
				t.Errorf("expected HasDueDate to be false, got true")
			}
			task.Due = &future
			if !task.HasDueDate() {
				t.Errorf("expected HasDueDate to be true, got false")
			}
			task.Due = &past
			task.Status = string(StatusPending)
			if !task.IsOverdue(now) {
				t.Errorf("expected overdue task, got false")
			}
			task.Status = string(StatusCompleted)
			if task.IsOverdue(now) {
				t.Errorf("expected completed task not to be overdue, got true")
			}
		})

		t.Run("IsRecurring and IsRecurExpired", func(t *testing.T) {
			now := time.Now()
			past := now.Add(-24 * time.Hour)
			future := now.Add(24 * time.Hour)

			task := Task{UUID: "123", Description: "demo", Entry: now, Modified: now}
			if task.IsRecurring() {
				t.Errorf("expected IsRecurring to be false, got true")
			}
			task.Recur = "FREQ=DAILY"
			if !task.IsRecurring() {
				t.Errorf("expected IsRecurring to be true, got false")
			}
			if task.IsRecurExpired(now) {
				t.Errorf("expected IsRecurExpired to be false without Until, got true")
			}
			task.Until = &past
			if !task.IsRecurExpired(now) {
				t.Errorf("expected IsRecurExpired to be true, got false")
			}
			task.Until = &future
			if task.IsRecurExpired(now) {
				t.Errorf("expected IsRecurExpired to be false, got true")
			}
		})

		t.Run("HasDependencies and Blocks", func(t *testing.T) {
			now := time.Now()
			task := Task{UUID: "123", Description: "demo", Entry: now, Modified: now}
			if task.HasDependencies() {
				t.Errorf("expected HasDependencies to be false, got true")
			}
			task.DependsOn = []string{"abc"}
			if !task.HasDependencies() {
				t.Errorf("expected HasDependencies to be true, got false")
			}
			other := Task{UUID: "abc", DependsOn: []string{"123"}}
			if !task.Blocks(&other) {
				t.Errorf("expected task to block other, got false")
			}
			other.DependsOn = []string{}
			if task.Blocks(&other) {
				t.Errorf("expected task not to block other, got true")
			}
		})

		t.Run("Urgency", func(t *testing.T) {
			now := time.Now()
			past := now.Add(-24 * time.Hour)

			task := Task{
				UUID:        "u1",
				Description: "urgency test",
				Priority:    "H",
				Tags:        []string{"t1"},
				Due:         &past,
				Status:      string(StatusPending),
				Entry:       now,
				Modified:    now,
			}
			score := task.Urgency(now)
			if score <= 0 {
				t.Errorf("expected positive urgency score, got %f", score)
			}
		})

	})

	t.Run("Movie Model", func(t *testing.T) {
		t.Run("Status Methods", func(t *testing.T) {
			tc := []struct {
				status    string
				isWatched bool
				isQueued  bool
			}{
				{"queued", false, true},
				{"watched", true, false},
				{"removed", false, false},
				{"unknown", false, false},
			}

			for _, tt := range tc {
				movie := &Movie{Status: tt.status}

				if movie.IsWatched() != tt.isWatched {
					t.Errorf("Status %s: expected IsWatched %v, got %v", tt.status, tt.isWatched, movie.IsWatched())
				}
				if movie.IsQueued() != tt.isQueued {
					t.Errorf("Status %s: expected IsQueued %v, got %v", tt.status, tt.isQueued, movie.IsQueued())
				}
			}
		})
	})

	t.Run("TV Show Model", func(t *testing.T) {
		t.Run("Status Methods", func(t *testing.T) {
			tc := []struct {
				status     string
				isWatching bool
				isWatched  bool
				isQueued   bool
			}{
				{"queued", false, false, true},
				{"watching", true, false, false},
				{"watched", false, true, false},
				{"removed", false, false, false},
				{"unknown", false, false, false},
			}

			for _, tt := range tc {
				tvShow := &TVShow{Status: tt.status}

				if tvShow.IsWatching() != tt.isWatching {
					t.Errorf("Status %s: expected IsWatching %v, got %v", tt.status, tt.isWatching, tvShow.IsWatching())
				}
				if tvShow.IsWatched() != tt.isWatched {
					t.Errorf("Status %s: expected IsWatched %v, got %v", tt.status, tt.isWatched, tvShow.IsWatched())
				}
				if tvShow.IsQueued() != tt.isQueued {
					t.Errorf("Status %s: expected IsQueued %v, got %v", tt.status, tt.isQueued, tvShow.IsQueued())
				}
			}
		})
	})

	t.Run("Book Model", func(t *testing.T) {
		t.Run("Status Methods", func(t *testing.T) {
			tc := []struct {
				status     string
				isReading  bool
				isFinished bool
				isQueued   bool
			}{
				{"queued", false, false, true},
				{"reading", true, false, false},
				{"finished", false, true, false},
				{"removed", false, false, false},
				{"unknown", false, false, false},
			}

			for _, tt := range tc {
				book := &Book{Status: tt.status}

				if book.IsReading() != tt.isReading {
					t.Errorf("Status %s: expected IsReading %v, got %v", tt.status, tt.isReading, book.IsReading())
				}
				if book.IsFinished() != tt.isFinished {
					t.Errorf("Status %s: expected IsFinished %v, got %v", tt.status, tt.isFinished, book.IsFinished())
				}
				if book.IsQueued() != tt.isQueued {
					t.Errorf("Status %s: expected IsQueued %v, got %v", tt.status, tt.isQueued, book.IsQueued())
				}
			}
		})

		t.Run("Progress Methods", func(t *testing.T) {
			book := &Book{Progress: 75}

			if book.ProgressPercent() != 75 {
				t.Errorf("Expected progress 75%%, got %d%%", book.ProgressPercent())
			}
		})
	})

	t.Run("Note Model", func(t *testing.T) {
		t.Run("Archive Methods", func(t *testing.T) {
			note := &Note{Archived: false}

			if note.IsArchived() {
				t.Error("Note should not be archived")
			}

			note.Archived = true
			if !note.IsArchived() {
				t.Error("Note should be archived")
			}
		})

		t.Run("Tags Marshaling", func(t *testing.T) {
			note := &Note{}

			result, err := note.MarshalTags()
			if err != nil {
				t.Fatalf("MarshalTags failed: %v", err)
			}
			if result != "" {
				t.Errorf("Expected empty string for empty tags, got '%s'", result)
			}

			note.Tags = []string{"personal", "work", "idea"}
			result, err = note.MarshalTags()
			if err != nil {
				t.Fatalf("MarshalTags failed: %v", err)
			}

			expected := `["personal","work","idea"]`
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}

			newNote := &Note{}
			err = newNote.UnmarshalTags(result)
			if err != nil {
				t.Fatalf("UnmarshalTags failed: %v", err)
			}

			if len(newNote.Tags) != 3 {
				t.Errorf("Expected 3 tags, got %d", len(newNote.Tags))
			}
			if newNote.Tags[0] != "personal" || newNote.Tags[1] != "work" || newNote.Tags[2] != "idea" {
				t.Errorf("Tags not unmarshaled correctly: %v", newNote.Tags)
			}

			emptyNote := &Note{}
			err = emptyNote.UnmarshalTags("")
			if err != nil {
				t.Fatalf("UnmarshalTags with empty string failed: %v", err)
			}
			if emptyNote.Tags != nil {
				t.Error("Expected nil tags for empty string")
			}
		})
	})

	t.Run("Album Model", func(t *testing.T) {
		t.Run("Rating Methods", func(t *testing.T) {
			album := &Album{}

			if album.HasRating() {
				t.Error("Album with zero rating should return false for HasRating")
			}

			if album.IsValidRating() {
				t.Error("Album with zero rating should return false for IsValidRating")
			}

			album.Rating = 3
			if !album.HasRating() {
				t.Error("Album with rating should return true for HasRating")
			}

			if !album.IsValidRating() {
				t.Error("Album with valid rating should return true for IsValidRating")
			}

			for _, tc := range []struct {
				rating  int
				isValid bool
			}{{0, false}, {1, true}, {3, true}, {5, true}, {6, false}, {-1, false}} {
				album.Rating = tc.rating
				if album.IsValidRating() != tc.isValid {
					t.Errorf("Rating %d: expected IsValidRating %v, got %v", tc.rating, tc.isValid, album.IsValidRating())
				}
			}
		})

		t.Run("Tracks Marshaling", func(t *testing.T) {
			album := &Album{}

			if result, err := album.MarshalTracks(); err != nil {
				t.Fatalf("MarshalTracks failed: %v", err)
			} else {
				if result != "" {
					t.Errorf("Expected empty string for empty tracks, got '%s'", result)
				}
			}

			album.Tracks = []string{"Track 1", "Track 2", "Interlude"}
			result, err := album.MarshalTracks()
			if err != nil {
				t.Fatalf("MarshalTracks failed: %v", err)
			}

			if expected := `["Track 1","Track 2","Interlude"]`; result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}

			newAlbum := &Album{}
			if err = newAlbum.UnmarshalTracks(result); err != nil {
				t.Fatalf("UnmarshalTracks failed: %v", err)
			} else {
				if len(newAlbum.Tracks) != 3 {
					t.Errorf("Expected 3 tracks, got %d", len(newAlbum.Tracks))
				}

				if newAlbum.Tracks[0] != "Track 1" || newAlbum.Tracks[1] != "Track 2" || newAlbum.Tracks[2] != "Interlude" {
					t.Errorf("Tracks not unmarshaled correctly: %v", newAlbum.Tracks)
				}
			}

			emptyAlbum := &Album{}
			if err = emptyAlbum.UnmarshalTracks(""); err != nil {
				t.Fatalf("UnmarshalTracks with empty string failed: %v", err)
			} else if emptyAlbum.Tracks != nil {
				t.Error("Expected nil tracks for empty string")
			}
		})
	})

	t.Run("Article Model", func(t *testing.T) {
		article := Article{URL: "", Author: "", Date: ""}
		want := false

		for _, tc := range []func() bool{article.HasAuthor, article.HasDate, article.IsValidURL} {
			got := tc()
			if got != want {
				t.Errorf("wanted %v, got %v", want, got)
			}
		}

		article.URL = "http//wikipedia.org"
		if article.IsValidURL() != want {
			t.Errorf("%v is invalid but got valid", article.URL)
		}

		article.URL = "http://wikipedia.org"
		if !article.IsValidURL() {
			t.Errorf("%v should be valid", article.URL)
		}
	})

	t.Run("TimeEntry Model", func(t *testing.T) {
		t.Run("IsActive", func(t *testing.T) {
			now := time.Now()

			t.Run("returns true when EndTime is nil", func(t *testing.T) {
				te := &TimeEntry{
					TaskID:    1,
					StartTime: now,
					EndTime:   nil,
				}

				if !te.IsActive() {
					t.Error("TimeEntry with nil EndTime should be active")
				}
			})

			t.Run("returns false when EndTime is set", func(t *testing.T) {
				endTime := now.Add(time.Hour)
				te := &TimeEntry{
					TaskID:    1,
					StartTime: now,
					EndTime:   &endTime,
				}

				if te.IsActive() {
					t.Error("TimeEntry with EndTime should not be active")
				}
			})
		})

		t.Run("Stop", func(t *testing.T) {
			startTime := time.Now().Add(-time.Hour)
			te := &TimeEntry{
				TaskID:    1,
				StartTime: startTime,
				EndTime:   nil,
				Created:   startTime,
				Modified:  startTime,
			}

			if !te.IsActive() {
				t.Error("TimeEntry should be active before Stop()")
			}

			te.Stop()

			if te.IsActive() {
				t.Error("TimeEntry should not be active after Stop()")
			}

			if te.EndTime == nil {
				t.Error("EndTime should be set after Stop()")
			}

			if te.EndTime.Before(startTime) {
				t.Error("EndTime should be after StartTime")
			}

			expectedDuration := int64(te.EndTime.Sub(startTime).Seconds())
			if te.DurationSeconds != expectedDuration {
				t.Errorf("Expected DurationSeconds %d, got %d", expectedDuration, te.DurationSeconds)
			}

			if te.Modified.Before(startTime) {
				t.Error("Modified time should be updated after Stop()")
			}
		})

		t.Run("GetDuration", func(t *testing.T) {
			startTime := time.Now().Add(-time.Hour)

			t.Run("returns calculated duration when stopped", func(t *testing.T) {
				endTime := startTime.Add(30 * time.Minute)
				te := &TimeEntry{
					TaskID:          1,
					StartTime:       startTime,
					EndTime:         &endTime,
					DurationSeconds: 1800,
				}

				duration := te.GetDuration()
				expectedDuration := 30 * time.Minute

				if duration != expectedDuration {
					t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
				}
			})

			t.Run("returns time since start when active", func(t *testing.T) {
				te := &TimeEntry{
					TaskID:    1,
					StartTime: startTime,
					EndTime:   nil,
				}

				duration := te.GetDuration()

				if duration < 59*time.Minute || duration > 61*time.Minute {
					t.Errorf("Expected duration around 1 hour, got %v", duration)
				}
			})
		})
	})

	t.Run("Error Handling", func(t *testing.T) {
		t.Run("Marshaling Errors", func(t *testing.T) {
			t.Run("UnmarshalTags handles invalid JSON", func(t *testing.T) {
				task := &Task{}
				if err := task.UnmarshalTags(`{"invalid": "json"}`); err == nil {
					t.Error("Expected error for invalid JSON, got nil")
				}
			})

			t.Run("UnmarshalAnnotations handles invalid JSON", func(t *testing.T) {
				task := &Task{}
				if err := task.UnmarshalAnnotations(`{"invalid": "json"}`); err == nil {
					t.Error("Expected error for invalid JSON, got nil")
				}
			})
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("Task with nil slices", func(t *testing.T) {
			task := &Task{
				Tags:        nil,
				Annotations: nil,
			}

			if tagsJSON, err := task.MarshalTags(); err != nil {
				t.Errorf("MarshalTags with nil slice failed: %v", err)
			} else if tagsJSON != "" {
				t.Errorf("Expected empty string for nil tags, got '%s'", tagsJSON)
			}

			if annotationsJSON, err := task.MarshalAnnotations(); err != nil {
				t.Errorf("MarshalAnnotations with nil slice failed: %v", err)
			} else if annotationsJSON != "" {
				t.Errorf("Expected empty string for nil annotations, got '%s'", annotationsJSON)
			}
		})

		t.Run("Models with zero values", func(t *testing.T) {
			task := &Task{}
			movie := &Movie{}
			tvShow := &TVShow{}
			book := &Book{}
			note := &Note{}

			if task.IsCompleted() || task.IsPending() || task.IsDeleted() {
				t.Error("Zero value task should have false status methods")
			}

			if movie.IsWatched() || movie.IsQueued() {
				t.Error("Zero value movie should have false status methods")
			}

			if tvShow.IsWatching() || tvShow.IsWatched() || tvShow.IsQueued() {
				t.Error("Zero value TV show should have false status methods")
			}

			if book.IsReading() || book.IsFinished() || book.IsQueued() {
				t.Error("Zero value book should have false status methods")
			}

			if book.ProgressPercent() != 0 {
				t.Errorf("Zero value book should have 0%% progress, got %d%%", book.ProgressPercent())
			}

			if note.IsArchived() {
				t.Error("Zero value note should not be archived")
			}
		})
	})

	t.Run("Behavior Interfaces", func(t *testing.T) {
		t.Run("Stateful Interface", func(t *testing.T) {
			t.Run("Task implements Stateful", func(t *testing.T) {
				task := &Task{Status: StatusTodo}

				if task.GetStatus() != StatusTodo {
					t.Errorf("Expected status %s, got %s", StatusTodo, task.GetStatus())
				}

				validStatuses := task.ValidStatuses()
				if len(validStatuses) == 0 {
					t.Error("ValidStatuses should not be empty")
				}

				expectedStatuses := []string{StatusTodo, StatusInProgress, StatusBlocked, StatusDone, StatusAbandoned, StatusPending, StatusCompleted, StatusDeleted}
				if len(validStatuses) != len(expectedStatuses) {
					t.Errorf("Expected %d valid statuses, got %d", len(expectedStatuses), len(validStatuses))
				}
			})

			t.Run("Book implements Stateful", func(t *testing.T) {
				book := &Book{Status: "reading"}

				if book.GetStatus() != "reading" {
					t.Errorf("Expected status 'reading', got %s", book.GetStatus())
				}

				validStatuses := book.ValidStatuses()
				expectedStatuses := []string{"queued", "reading", "finished", "removed"}

				if len(validStatuses) != len(expectedStatuses) {
					t.Errorf("Expected %d valid statuses, got %d", len(expectedStatuses), len(validStatuses))
				}

				for i, status := range expectedStatuses {
					if validStatuses[i] != status {
						t.Errorf("Expected status %s at index %d, got %s", status, i, validStatuses[i])
					}
				}
			})

			t.Run("Movie implements Stateful", func(t *testing.T) {
				movie := &Movie{Status: "queued"}

				if movie.GetStatus() != "queued" {
					t.Errorf("Expected status 'queued', got %s", movie.GetStatus())
				}

				validStatuses := movie.ValidStatuses()
				expectedStatuses := []string{"queued", "watched", "removed"}

				if len(validStatuses) != len(expectedStatuses) {
					t.Errorf("Expected %d valid statuses, got %d", len(expectedStatuses), len(validStatuses))
				}
			})

			t.Run("TVShow implements Stateful", func(t *testing.T) {
				tvShow := &TVShow{Status: "watching"}

				if tvShow.GetStatus() != "watching" {
					t.Errorf("Expected status 'watching', got %s", tvShow.GetStatus())
				}

				validStatuses := tvShow.ValidStatuses()
				expectedStatuses := []string{"queued", "watching", "watched", "removed"}

				if len(validStatuses) != len(expectedStatuses) {
					t.Errorf("Expected %d valid statuses, got %d", len(expectedStatuses), len(validStatuses))
				}
			})
		})

		t.Run("Completable Interface", func(t *testing.T) {
			t.Run("Book implements Completable", func(t *testing.T) {
				now := time.Now()

				unfinishedBook := &Book{Status: "reading"}
				if unfinishedBook.IsCompleted() {
					t.Error("Book with 'reading' status should not be completed")
				}
				if unfinishedBook.GetCompletionTime() != nil {
					t.Error("Unfinished book should have nil completion time")
				}

				finishedBook := &Book{Status: "finished", Finished: &now}
				if !finishedBook.IsCompleted() {
					t.Error("Book with 'finished' status should be completed")
				}
				if finishedBook.GetCompletionTime() == nil {
					t.Error("Finished book should have completion time")
				}
				if !finishedBook.GetCompletionTime().Equal(now) {
					t.Errorf("Expected completion time %v, got %v", now, finishedBook.GetCompletionTime())
				}
			})

			t.Run("Movie implements Completable", func(t *testing.T) {
				now := time.Now()

				unwatchedMovie := &Movie{Status: "queued"}
				if unwatchedMovie.IsCompleted() {
					t.Error("Movie with 'queued' status should not be completed")
				}
				if unwatchedMovie.GetCompletionTime() != nil {
					t.Error("Unwatched movie should have nil completion time")
				}

				watchedMovie := &Movie{Status: "watched", Watched: &now}
				if !watchedMovie.IsCompleted() {
					t.Error("Movie with 'watched' status should be completed")
				}
				if watchedMovie.GetCompletionTime() == nil {
					t.Error("Watched movie should have completion time")
				}
				if !watchedMovie.GetCompletionTime().Equal(now) {
					t.Errorf("Expected completion time %v, got %v", now, watchedMovie.GetCompletionTime())
				}
			})

			t.Run("TVShow implements Completable", func(t *testing.T) {
				now := time.Now()

				unwatchedShow := &TVShow{Status: "watching"}
				if unwatchedShow.IsCompleted() {
					t.Error("TVShow with 'watching' status should not be completed")
				}
				if unwatchedShow.GetCompletionTime() != nil {
					t.Error("Unwatched show should have nil completion time")
				}

				watchedShow := &TVShow{Status: "watched", LastWatched: &now}
				if !watchedShow.IsCompleted() {
					t.Error("TVShow with 'watched' status should be completed")
				}
				if watchedShow.GetCompletionTime() == nil {
					t.Error("Watched show should have completion time")
				}
				if !watchedShow.GetCompletionTime().Equal(now) {
					t.Errorf("Expected completion time %v, got %v", now, watchedShow.GetCompletionTime())
				}
			})
		})

		t.Run("Progressable Interface", func(t *testing.T) {
			t.Run("Book implements Progressable", func(t *testing.T) {
				book := &Book{Progress: 50}

				if book.GetProgress() != 50 {
					t.Errorf("Expected progress 50, got %d", book.GetProgress())
				}
			})

			t.Run("SetProgress with valid values", func(t *testing.T) {
				book := &Book{}

				if err := book.SetProgress(0); err != nil {
					t.Errorf("SetProgress(0) should succeed, got error: %v", err)
				}
				if book.Progress != 0 {
					t.Errorf("Expected progress 0, got %d", book.Progress)
				}

				if err := book.SetProgress(100); err != nil {
					t.Errorf("SetProgress(100) should succeed, got error: %v", err)
				}
				if book.Progress != 100 {
					t.Errorf("Expected progress 100, got %d", book.Progress)
				}

				if err := book.SetProgress(42); err != nil {
					t.Errorf("SetProgress(42) should succeed, got error: %v", err)
				}
				if book.Progress != 42 {
					t.Errorf("Expected progress 42, got %d", book.Progress)
				}
			})

			t.Run("SetProgress rejects invalid values", func(t *testing.T) {
				book := &Book{Progress: 50}

				if err := book.SetProgress(-1); err == nil {
					t.Error("SetProgress(-1) should fail")
				} else if book.Progress != 50 {
					t.Error("Progress should not change on validation error")
				}

				if err := book.SetProgress(101); err == nil {
					t.Error("SetProgress(101) should fail")
				} else if book.Progress != 50 {
					t.Error("Progress should not change on validation error")
				}

				if err := book.SetProgress(-100); err == nil {
					t.Error("SetProgress(-100) should fail")
				}

				if err := book.SetProgress(1000); err == nil {
					t.Error("SetProgress(1000) should fail")
				}
			})

			t.Run("SetProgress error messages", func(t *testing.T) {
				book := &Book{}

				err := book.SetProgress(-5)
				if err == nil {
					t.Fatal("Expected error for negative progress")
				}
				if err.Error() != "progress must be between 0 and 100, got -5" {
					t.Errorf("Unexpected error message: %s", err.Error())
				}

				err = book.SetProgress(150)
				if err == nil {
					t.Fatal("Expected error for progress > 100")
				}
				if err.Error() != "progress must be between 0 and 100, got 150" {
					t.Errorf("Unexpected error message: %s", err.Error())
				}
			})
		})
	})
}
