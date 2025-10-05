package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// mockTagRepository implements TagRepository for testing
type mockTagRepository struct {
	tags []repo.TagSummary
	err  error
}

func (m *mockTagRepository) GetTags(ctx context.Context) ([]repo.TagSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tags, nil
}

func TestTagAdapter(t *testing.T) {
	t.Run("TagSummaryRecord", func(t *testing.T) {
		summary := repo.TagSummary{
			Name:      "work",
			TaskCount: 8,
		}
		record := &TagSummaryRecord{summary: summary}

		t.Run("GetField", func(t *testing.T) {
			tests := []struct {
				field    string
				expected any
				name     string
			}{
				{"name", "work", "should return tag name"},
				{"task_count", 8, "should return task count"},
				{"unknown", "", "should return empty string for unknown field"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result := record.GetField(tt.field)
					if result != tt.expected {
						t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
					}
				})
			}
		})

		t.Run("ModelInterface", func(t *testing.T) {
			if record.GetID() != 8 {
				t.Errorf("GetID() = %d, want 8", record.GetID())
			}

			if record.GetTableName() != "tags" {
				t.Errorf("GetTableName() = %q, want 'tags'", record.GetTableName())
			}

			Expect.AssertZeroTime(t, record.GetCreatedAt, "GetCreatedAt")
			Expect.AssertZeroTime(t, record.GetUpdatedAt, "GetUpdatedAt")
		})
	})

	t.Run("TagDataSource", func(t *testing.T) {
		t.Run("Load", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "work", TaskCount: 5},
				{Name: "personal", TaskCount: 3},
			}
			repo := &mockTagRepository{tags: tags}
			source := &TagDataSource{repo: repo}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 2 {
				t.Errorf("Load() returned %d records, want 2", len(records))
			}

			if records[0].GetField("name") != "work" {
				t.Errorf("First record name = %v, want 'work'", records[0].GetField("name"))
			}

			if records[0].GetField("task_count") != 5 {
				t.Errorf("First record task_count = %v, want 5", records[0].GetField("task_count"))
			}
		})

		t.Run("Load_Error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockTagRepository{err: testErr}
			source := &TagDataSource{repo: repo}

			_, err := source.Load(context.Background(), DataOptions{})
			if err != testErr {
				t.Errorf("Load() error = %v, want %v", err, testErr)
			}
		})

		t.Run("Count", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "work", TaskCount: 5},
				{Name: "personal", TaskCount: 3},
				{Name: "urgent", TaskCount: 1},
			}
			repo := &mockTagRepository{tags: tags}
			source := &TagDataSource{repo: repo}

			count, err := source.Count(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Count() failed: %v", err)
			}

			if count != 3 {
				t.Errorf("Count() = %d, want 3", count)
			}
		})

		t.Run("Count_Error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockTagRepository{err: testErr}
			source := &TagDataSource{repo: repo}

			_, err := source.Count(context.Background(), DataOptions{})
			if err != testErr {
				t.Errorf("Count() error = %v, want %v", err, testErr)
			}
		})
	})

	t.Run("NewTagDataTable", func(t *testing.T) {
		repo := &mockTagRepository{
			tags: []repo.TagSummary{
				{Name: "work", TaskCount: 4},
			},
		}

		opts := DataTableOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		table := NewTagDataTable(repo, opts)
		if table == nil {
			t.Fatal("NewTagDataTable() returned nil")
		}

		err := table.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}
	})

	t.Run("NewTagListFromTable", func(t *testing.T) {
		repo := &mockTagRepository{
			tags: []repo.TagSummary{
				{Name: "urgent", TaskCount: 1},
			},
		}

		output := &bytes.Buffer{}
		input := strings.NewReader("q\n")

		table := NewTagListFromTable(repo, output, input, true)
		if table == nil {
			t.Fatal("NewTagListFromTable() returned nil")
		}

		err := table.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}

		outputStr := output.String()
		if !strings.Contains(outputStr, "Tags") {
			t.Error("Output should contain 'Tags' title")
		}
		if !strings.Contains(outputStr, "urgent") {
			t.Error("Output should contain tag name")
		}
		if !strings.Contains(outputStr, "1 task") {
			t.Error("Output should contain formatted task count")
		}
	})
}
