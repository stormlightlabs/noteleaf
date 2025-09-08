package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestModels(t *testing.T) {
	t.Run("Task Model", func(t *testing.T) {
		t.Run("Model Interface Implementation", func(t *testing.T) {
			task := &Task{
				ID:          1,
				UUID:        "test-uuid",
				Description: "Test task",
				Status:      "pending",
				Entry:       time.Now(),
				Modified:    time.Now(),
			}

			if task.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", task.GetID())
			}

			task.SetID(2)
			if task.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", task.GetID())
			}

			if task.GetTableName() != "tasks" {
				t.Errorf("Expected table name 'tasks', got '%s'", task.GetTableName())
			}

			createdAt := time.Now()
			task.SetCreatedAt(createdAt)
			if !task.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, task.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			task.SetUpdatedAt(updatedAt)
			if !task.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, task.GetUpdatedAt())
			}
		})

		t.Run("Status Methods", func(t *testing.T) {
			testCases := []struct {
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

			for _, tc := range testCases {
				task := &Task{Status: tc.status}

				if task.IsCompleted() != tc.isCompleted {
					t.Errorf("Status %s: expected IsCompleted %v, got %v", tc.status, tc.isCompleted, task.IsCompleted())
				}
				if task.IsPending() != tc.isPending {
					t.Errorf("Status %s: expected IsPending %v, got %v", tc.status, tc.isPending, task.IsPending())
				}
				if task.IsDeleted() != tc.isDeleted {
					t.Errorf("Status %s: expected IsDeleted %v, got %v", tc.status, tc.isDeleted, task.IsDeleted())
				}
			}
		})

		t.Run("New Status Tracking Methods", func(t *testing.T) {
			testCases := []struct {
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

			for _, tc := range testCases {
				task := &Task{Status: tc.status}

				if task.IsTodo() != tc.isTodo {
					t.Errorf("Status %s: expected IsTodo %v, got %v", tc.status, tc.isTodo, task.IsTodo())
				}
				if task.IsInProgress() != tc.isInProgress {
					t.Errorf("Status %s: expected IsInProgress %v, got %v", tc.status, tc.isInProgress, task.IsInProgress())
				}
				if task.IsBlocked() != tc.isBlocked {
					t.Errorf("Status %s: expected IsBlocked %v, got %v", tc.status, tc.isBlocked, task.IsBlocked())
				}
				if task.IsDone() != tc.isDone {
					t.Errorf("Status %s: expected IsDone %v, got %v", tc.status, tc.isDone, task.IsDone())
				}
				if task.IsAbandoned() != tc.isAbandoned {
					t.Errorf("Status %s: expected IsAbandoned %v, got %v", tc.status, tc.isAbandoned, task.IsAbandoned())
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
				testCases := []struct {
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

				for _, tc := range testCases {
					task := &Task{Priority: tc.priority}
					weight := task.GetPriorityWeight()
					if weight != tc.weight {
						t.Errorf("Priority %s: expected weight %d, got %d", tc.priority, tc.weight, weight)
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

		t.Run("JSON Marshaling", func(t *testing.T) {
			now := time.Now()
			due := now.Add(24 * time.Hour)
			task := &Task{
				ID:          1,
				UUID:        "test-uuid",
				Description: "Test task",
				Status:      "pending",
				Priority:    "A",
				Project:     "test-project",
				Tags:        []string{"work", "urgent"},
				Due:         &due,
				Entry:       now,
				Modified:    now,
				Annotations: []string{"Note 1"},
			}

			data, err := json.Marshal(task)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var unmarshaled Task
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			if unmarshaled.ID != task.ID {
				t.Errorf("Expected ID %d, got %d", task.ID, unmarshaled.ID)
			}
			if unmarshaled.UUID != task.UUID {
				t.Errorf("Expected UUID %s, got %s", task.UUID, unmarshaled.UUID)
			}
			if unmarshaled.Description != task.Description {
				t.Errorf("Expected description %s, got %s", task.Description, unmarshaled.Description)
			}
		})
	})

	t.Run("Movie Model", func(t *testing.T) {
		t.Run("Model Interface Implementation", func(t *testing.T) {
			movie := &Movie{
				ID:    1,
				Title: "Test Movie",
				Year:  2023,
				Added: time.Now(),
			}

			if movie.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", movie.GetID())
			}

			movie.SetID(2)
			if movie.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", movie.GetID())
			}

			if movie.GetTableName() != "movies" {
				t.Errorf("Expected table name 'movies', got '%s'", movie.GetTableName())
			}

			createdAt := time.Now()
			movie.SetCreatedAt(createdAt)
			if !movie.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, movie.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			movie.SetUpdatedAt(updatedAt)
			if !movie.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, movie.GetUpdatedAt())
			}
		})

		t.Run("Status Methods", func(t *testing.T) {
			testCases := []struct {
				status    string
				isWatched bool
				isQueued  bool
			}{
				{"queued", false, true},
				{"watched", true, false},
				{"removed", false, false},
				{"unknown", false, false},
			}

			for _, tc := range testCases {
				movie := &Movie{Status: tc.status}

				if movie.IsWatched() != tc.isWatched {
					t.Errorf("Status %s: expected IsWatched %v, got %v", tc.status, tc.isWatched, movie.IsWatched())
				}
				if movie.IsQueued() != tc.isQueued {
					t.Errorf("Status %s: expected IsQueued %v, got %v", tc.status, tc.isQueued, movie.IsQueued())
				}
			}
		})

		t.Run("JSON Marshaling", func(t *testing.T) {
			now := time.Now()
			watched := now.Add(-24 * time.Hour)
			movie := &Movie{
				ID:      1,
				Title:   "Test Movie",
				Year:    2023,
				Status:  "watched",
				Rating:  8.5,
				Notes:   "Great movie!",
				Added:   now,
				Watched: &watched,
			}

			data, err := json.Marshal(movie)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var unmarshaled Movie
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			if unmarshaled.ID != movie.ID {
				t.Errorf("Expected ID %d, got %d", movie.ID, unmarshaled.ID)
			}
			if unmarshaled.Title != movie.Title {
				t.Errorf("Expected title %s, got %s", movie.Title, unmarshaled.Title)
			}
			if unmarshaled.Rating != movie.Rating {
				t.Errorf("Expected rating %f, got %f", movie.Rating, unmarshaled.Rating)
			}
		})
	})

	t.Run("TV Show Model", func(t *testing.T) {
		t.Run("Model Interface Implementation", func(t *testing.T) {
			tvShow := &TVShow{
				ID:    1,
				Title: "Test Show",
				Added: time.Now(),
			}

			if tvShow.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", tvShow.GetID())
			}

			tvShow.SetID(2)
			if tvShow.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", tvShow.GetID())
			}

			if tvShow.GetTableName() != "tv_shows" {
				t.Errorf("Expected table name 'tv_shows', got '%s'", tvShow.GetTableName())
			}

			createdAt := time.Now()
			tvShow.SetCreatedAt(createdAt)
			if !tvShow.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, tvShow.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			tvShow.SetUpdatedAt(updatedAt)
			if !tvShow.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, tvShow.GetUpdatedAt())
			}
		})

		t.Run("Status Methods", func(t *testing.T) {
			testCases := []struct {
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

			for _, tc := range testCases {
				tvShow := &TVShow{Status: tc.status}

				if tvShow.IsWatching() != tc.isWatching {
					t.Errorf("Status %s: expected IsWatching %v, got %v", tc.status, tc.isWatching, tvShow.IsWatching())
				}
				if tvShow.IsWatched() != tc.isWatched {
					t.Errorf("Status %s: expected IsWatched %v, got %v", tc.status, tc.isWatched, tvShow.IsWatched())
				}
				if tvShow.IsQueued() != tc.isQueued {
					t.Errorf("Status %s: expected IsQueued %v, got %v", tc.status, tc.isQueued, tvShow.IsQueued())
				}
			}
		})

		t.Run("JSON Marshaling", func(t *testing.T) {
			now := time.Now()
			lastWatched := now.Add(-24 * time.Hour)
			tvShow := &TVShow{
				ID:          1,
				Title:       "Test Show",
				Season:      1,
				Episode:     5,
				Status:      "watching",
				Rating:      9.0,
				Notes:       "Amazing series!",
				Added:       now,
				LastWatched: &lastWatched,
			}

			data, err := json.Marshal(tvShow)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var unmarshaled TVShow
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			if unmarshaled.ID != tvShow.ID {
				t.Errorf("Expected ID %d, got %d", tvShow.ID, unmarshaled.ID)
			}
			if unmarshaled.Title != tvShow.Title {
				t.Errorf("Expected title %s, got %s", tvShow.Title, unmarshaled.Title)
			}
			if unmarshaled.Season != tvShow.Season {
				t.Errorf("Expected season %d, got %d", tvShow.Season, unmarshaled.Season)
			}
			if unmarshaled.Episode != tvShow.Episode {
				t.Errorf("Expected episode %d, got %d", tvShow.Episode, unmarshaled.Episode)
			}
		})
	})

	t.Run("Book Model", func(t *testing.T) {
		t.Run("Model Interface Implementation", func(t *testing.T) {
			book := &Book{
				ID:    1,
				Title: "Test Book",
				Added: time.Now(),
			}

			if book.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", book.GetID())
			}

			book.SetID(2)
			if book.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", book.GetID())
			}

			if book.GetTableName() != "books" {
				t.Errorf("Expected table name 'books', got '%s'", book.GetTableName())
			}

			createdAt := time.Now()
			book.SetCreatedAt(createdAt)
			if !book.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, book.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			book.SetUpdatedAt(updatedAt)
			if !book.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, book.GetUpdatedAt())
			}
		})

		t.Run("Status Methods", func(t *testing.T) {
			testCases := []struct {
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

			for _, tc := range testCases {
				book := &Book{Status: tc.status}

				if book.IsReading() != tc.isReading {
					t.Errorf("Status %s: expected IsReading %v, got %v", tc.status, tc.isReading, book.IsReading())
				}
				if book.IsFinished() != tc.isFinished {
					t.Errorf("Status %s: expected IsFinished %v, got %v", tc.status, tc.isFinished, book.IsFinished())
				}
				if book.IsQueued() != tc.isQueued {
					t.Errorf("Status %s: expected IsQueued %v, got %v", tc.status, tc.isQueued, book.IsQueued())
				}
			}
		})

		t.Run("Progress Methods", func(t *testing.T) {
			book := &Book{Progress: 75}

			if book.ProgressPercent() != 75 {
				t.Errorf("Expected progress 75%%, got %d%%", book.ProgressPercent())
			}
		})

		t.Run("JSON Marshaling", func(t *testing.T) {
			now := time.Now()
			started := now.Add(-7 * 24 * time.Hour)
			finished := now.Add(-24 * time.Hour)
			book := &Book{
				ID:       1,
				Title:    "Test Book",
				Author:   "Test Author",
				Status:   "finished",
				Progress: 100,
				Pages:    300,
				Rating:   4.5,
				Notes:    "Excellent read!",
				Added:    now,
				Started:  &started,
				Finished: &finished,
			}

			data, err := json.Marshal(book)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var unmarshaled Book
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			if unmarshaled.ID != book.ID {
				t.Errorf("Expected ID %d, got %d", book.ID, unmarshaled.ID)
			}
			if unmarshaled.Title != book.Title {
				t.Errorf("Expected title %s, got %s", book.Title, unmarshaled.Title)
			}
			if unmarshaled.Author != book.Author {
				t.Errorf("Expected author %s, got %s", book.Author, unmarshaled.Author)
			}
			if unmarshaled.Progress != book.Progress {
				t.Errorf("Expected progress %d, got %d", book.Progress, unmarshaled.Progress)
			}
			if unmarshaled.Pages != book.Pages {
				t.Errorf("Expected pages %d, got %d", book.Pages, unmarshaled.Pages)
			}
		})
	})

	t.Run("Note Model", func(t *testing.T) {
		t.Run("Model Interface Implementation", func(t *testing.T) {
			note := &Note{
				ID:      1,
				Title:   "Test Note",
				Content: "This is test content",
				Created: time.Now(),
			}

			if note.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", note.GetID())
			}

			note.SetID(2)
			if note.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", note.GetID())
			}

			if note.GetTableName() != "notes" {
				t.Errorf("Expected table name 'notes', got '%s'", note.GetTableName())
			}

			createdAt := time.Now()
			note.SetCreatedAt(createdAt)
			if !note.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, note.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			note.SetUpdatedAt(updatedAt)
			if !note.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, note.GetUpdatedAt())
			}
		})

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

		t.Run("JSON Marshaling", func(t *testing.T) {
			now := time.Now()
			modified := now.Add(time.Hour)
			note := &Note{
				ID:       1,
				Title:    "Test Note",
				Content:  "This is test content with **markdown**",
				Tags:     []string{"personal", "markdown"},
				Archived: false,
				Created:  now,
				Modified: modified,
				FilePath: "/path/to/note.md",
			}

			data, err := json.Marshal(note)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var unmarshaled Note
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			if unmarshaled.ID != note.ID {
				t.Errorf("Expected ID %d, got %d", note.ID, unmarshaled.ID)
			}
			if unmarshaled.Title != note.Title {
				t.Errorf("Expected title %s, got %s", note.Title, unmarshaled.Title)
			}
			if unmarshaled.Content != note.Content {
				t.Errorf("Expected content %s, got %s", note.Content, unmarshaled.Content)
			}
			if unmarshaled.Archived != note.Archived {
				t.Errorf("Expected archived %v, got %v", note.Archived, unmarshaled.Archived)
			}
			if unmarshaled.FilePath != note.FilePath {
				t.Errorf("Expected file path %s, got %s", note.FilePath, unmarshaled.FilePath)
			}
		})
	})

	t.Run("Album Model", func(t *testing.T) {
		t.Run("Model Interface Implementation", func(t *testing.T) {
			album := &Album{
				ID:      1,
				Title:   "Test Album",
				Artist:  "Test Artist",
				Created: time.Now(),
			}

			if album.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", album.GetID())
			}

			album.SetID(2)
			if album.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", album.GetID())
			}

			if album.GetTableName() != "albums" {
				t.Errorf("Expected table name 'albums', got '%s'", album.GetTableName())
			}

			createdAt := time.Now()
			album.SetCreatedAt(createdAt)
			if !album.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, album.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			album.SetUpdatedAt(updatedAt)
			if !album.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, album.GetUpdatedAt())
			}
		})

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

			testCases := []struct {
				rating  int
				isValid bool
			}{
				{0, false},
				{1, true},
				{3, true},
				{5, true},
				{6, false},
				{-1, false},
			}

			for _, tc := range testCases {
				album.Rating = tc.rating
				if album.IsValidRating() != tc.isValid {
					t.Errorf("Rating %d: expected IsValidRating %v, got %v", tc.rating, tc.isValid, album.IsValidRating())
				}
			}
		})

		t.Run("Tracks Marshaling", func(t *testing.T) {
			album := &Album{}

			result, err := album.MarshalTracks()
			if err != nil {
				t.Fatalf("MarshalTracks failed: %v", err)
			}
			if result != "" {
				t.Errorf("Expected empty string for empty tracks, got '%s'", result)
			}

			album.Tracks = []string{"Track 1", "Track 2", "Interlude"}
			result, err = album.MarshalTracks()
			if err != nil {
				t.Fatalf("MarshalTracks failed: %v", err)
			}

			expected := `["Track 1","Track 2","Interlude"]`
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			}

			newAlbum := &Album{}
			err = newAlbum.UnmarshalTracks(result)
			if err != nil {
				t.Fatalf("UnmarshalTracks failed: %v", err)
			}

			if len(newAlbum.Tracks) != 3 {
				t.Errorf("Expected 3 tracks, got %d", len(newAlbum.Tracks))
			}
			if newAlbum.Tracks[0] != "Track 1" || newAlbum.Tracks[1] != "Track 2" || newAlbum.Tracks[2] != "Interlude" {
				t.Errorf("Tracks not unmarshaled correctly: %v", newAlbum.Tracks)
			}

			emptyAlbum := &Album{}
			err = emptyAlbum.UnmarshalTracks("")
			if err != nil {
				t.Fatalf("UnmarshalTracks with empty string failed: %v", err)
			}
			if emptyAlbum.Tracks != nil {
				t.Error("Expected nil tracks for empty string")
			}
		})

		t.Run("JSON Marshaling", func(t *testing.T) {
			now := time.Now()
			modified := now.Add(time.Hour)
			album := &Album{
				ID:              1,
				Title:           "Test Album",
				Artist:          "Test Artist",
				Genre:           "Rock",
				ReleaseYear:     2023,
				Tracks:          []string{"Track 1", "Track 2"},
				DurationSeconds: 3600,
				AlbumArtPath:    "/path/to/art.jpg",
				Rating:          4,
				Created:         now,
				Modified:        modified,
			}

			data, err := json.Marshal(album)
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			var unmarshaled Album
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			if unmarshaled.ID != album.ID {
				t.Errorf("Expected ID %d, got %d", album.ID, unmarshaled.ID)
			}
			if unmarshaled.Title != album.Title {
				t.Errorf("Expected title %s, got %s", album.Title, unmarshaled.Title)
			}
			if unmarshaled.Artist != album.Artist {
				t.Errorf("Expected artist %s, got %s", album.Artist, unmarshaled.Artist)
			}
			if unmarshaled.Genre != album.Genre {
				t.Errorf("Expected genre %s, got %s", album.Genre, unmarshaled.Genre)
			}
			if unmarshaled.ReleaseYear != album.ReleaseYear {
				t.Errorf("Expected release year %d, got %d", album.ReleaseYear, unmarshaled.ReleaseYear)
			}
			if unmarshaled.DurationSeconds != album.DurationSeconds {
				t.Errorf("Expected duration %d, got %d", album.DurationSeconds, unmarshaled.DurationSeconds)
			}
			if unmarshaled.Rating != album.Rating {
				t.Errorf("Expected rating %d, got %d", album.Rating, unmarshaled.Rating)
			}
		})
	})

	t.Run("Interface Implementations", func(t *testing.T) {
		t.Run("All models implement Model interface", func(t *testing.T) {
			var models []Model

			task := &Task{}
			movie := &Movie{}
			tvShow := &TVShow{}
			book := &Book{}
			note := &Note{}
			album := &Album{}

			models = append(models, task, movie, tvShow, book, note, album)

			if len(models) != 6 {
				t.Errorf("Expected 6 models, got %d", len(models))
			}

			for i, model := range models {
				model.SetID(int64(i + 1))
				if model.GetID() != int64(i+1) {
					t.Errorf("Model %d: ID not set correctly", i)
				}

				tableName := model.GetTableName()
				if tableName == "" {
					t.Errorf("Model %d: table name should not be empty", i)
				}

				now := time.Now()
				model.SetCreatedAt(now)
				model.SetUpdatedAt(now)

				// NOTE: We don't test exact equality due to potential precision differences
				if model.GetCreatedAt().IsZero() {
					t.Errorf("Model %d: created at should not be zero", i)
				}
				if model.GetUpdatedAt().IsZero() {
					t.Errorf("Model %d: updated at should not be zero", i)
				}
			}
		})
	})

	t.Run("Errors & Edge cases", func(t *testing.T) {
		t.Run("Marshaling Errors", func(t *testing.T) {
			t.Run("UnmarshalTags handles invalid JSON", func(t *testing.T) {
				task := &Task{}
				err := task.UnmarshalTags(`{"invalid": "json"}`)
				if err == nil {
					t.Error("Expected error for invalid JSON, got nil")
				}
			})

			t.Run("UnmarshalAnnotations handles invalid JSON", func(t *testing.T) {
				task := &Task{}
				err := task.UnmarshalAnnotations(`{"invalid": "json"}`)
				if err == nil {
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

			tagsJSON, err := task.MarshalTags()
			if err != nil {
				t.Errorf("MarshalTags with nil slice failed: %v", err)
			}
			if tagsJSON != "" {
				t.Errorf("Expected empty string for nil tags, got '%s'", tagsJSON)
			}

			annotationsJSON, err := task.MarshalAnnotations()
			if err != nil {
				t.Errorf("MarshalAnnotations with nil slice failed: %v", err)
			}
			if annotationsJSON != "" {
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

		t.Run("Model Interface Implementation", func(t *testing.T) {
			now := time.Now()
			te := &TimeEntry{
				ID:       1,
				TaskID:   100,
				Created:  now,
				Modified: now,
			}

			if te.GetID() != 1 {
				t.Errorf("Expected ID 1, got %d", te.GetID())
			}

			te.SetID(2)
			if te.GetID() != 2 {
				t.Errorf("Expected ID 2 after SetID, got %d", te.GetID())
			}

			if te.GetTableName() != "time_entries" {
				t.Errorf("Expected table name 'time_entries', got '%s'", te.GetTableName())
			}

			createdAt := time.Now()
			te.SetCreatedAt(createdAt)
			if !te.GetCreatedAt().Equal(createdAt) {
				t.Errorf("Expected created at %v, got %v", createdAt, te.GetCreatedAt())
			}

			updatedAt := time.Now().Add(time.Hour)
			te.SetUpdatedAt(updatedAt)
			if !te.GetUpdatedAt().Equal(updatedAt) {
				t.Errorf("Expected updated at %v, got %v", updatedAt, te.GetUpdatedAt())
			}
		})
	})
}
