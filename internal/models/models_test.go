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

	t.Run("Interface Implementations", func(t *testing.T) {
		t.Run("All models implement Model interface", func(t *testing.T) {
			var models []Model

			task := &Task{}
			movie := &Movie{}
			tvShow := &TVShow{}
			book := &Book{}

			models = append(models, task, movie, tvShow, book)

			if len(models) != 4 {
				t.Errorf("Expected 4 models, got %d", len(models))
			}

			// Test that all models have the required methods
			for i, model := range models {
				// Test ID methods
				model.SetID(int64(i + 1))
				if model.GetID() != int64(i+1) {
					t.Errorf("Model %d: ID not set correctly", i)
				}

				// Test table name method
				tableName := model.GetTableName()
				if tableName == "" {
					t.Errorf("Model %d: table name should not be empty", i)
				}

				// Test timestamp methods
				now := time.Now()
				model.SetCreatedAt(now)
				model.SetUpdatedAt(now)

				// Note: We don't test exact equality due to potential precision differences
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

			// Test that zero values don't cause panics
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
		})
	})
}
