package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// mockProjectRepository implements ProjectRepository for testing
type mockProjectRepository struct {
	projects []repo.ProjectSummary
	err      error
}

func (m *mockProjectRepository) GetProjects(ctx context.Context) ([]repo.ProjectSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.projects, nil
}

func TestProjectAdapter(t *testing.T) {
	t.Run("ProjectSummaryRecord", func(t *testing.T) {
		summary := repo.ProjectSummary{
			Name:      "Test Project",
			TaskCount: 5,
		}
		record := &ProjectSummaryRecord{summary: summary}

		t.Run("GetField", func(t *testing.T) {
			tests := []struct {
				field    string
				expected any
				name     string
			}{
				{"name", "Test Project", "should return project name"},
				{"task_count", 5, "should return task count"},
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
			if record.GetID() != 5 {
				t.Errorf("GetID() = %d, want 5", record.GetID())
			}

			if record.GetTableName() != "projects" {
				t.Errorf("GetTableName() = %q, want 'projects'", record.GetTableName())
			}

			if !record.GetCreatedAt().IsZero() {
				t.Error("GetCreatedAt() should return zero time")
			}

			if !record.GetUpdatedAt().IsZero() {
				t.Error("GetUpdatedAt() should return zero time")
			}
		})
	})

	t.Run("ProjectDataSource", func(t *testing.T) {
		t.Run("Load", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "Project 1", TaskCount: 5},
				{Name: "Project 2", TaskCount: 3},
			}
			repo := &mockProjectRepository{projects: projects}
			source := &ProjectDataSource{repo: repo}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 2 {
				t.Errorf("Load() returned %d records, want 2", len(records))
			}

			if records[0].GetField("name") != "Project 1" {
				t.Errorf("First record name = %v, want 'Project 1'", records[0].GetField("name"))
			}

			if records[0].GetField("task_count") != 5 {
				t.Errorf("First record task_count = %v, want 5", records[0].GetField("task_count"))
			}
		})

		t.Run("Load_Error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockProjectRepository{err: testErr}
			source := &ProjectDataSource{repo: repo}

			_, err := source.Load(context.Background(), DataOptions{})
			if err != testErr {
				t.Errorf("Load() error = %v, want %v", err, testErr)
			}
		})

		t.Run("Count", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "Project 1", TaskCount: 5},
				{Name: "Project 2", TaskCount: 3},
				{Name: "Project 3", TaskCount: 1},
			}
			repo := &mockProjectRepository{projects: projects}
			source := &ProjectDataSource{repo: repo}

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
			repo := &mockProjectRepository{err: testErr}
			source := &ProjectDataSource{repo: repo}

			_, err := source.Count(context.Background(), DataOptions{})
			if err != testErr {
				t.Errorf("Count() error = %v, want %v", err, testErr)
			}
		})
	})

	t.Run("NewProjectDataTable", func(t *testing.T) {
		repo := &mockProjectRepository{
			projects: []repo.ProjectSummary{
				{Name: "Test Project", TaskCount: 2},
			},
		}

		opts := DataTableOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		table := NewProjectDataTable(repo, opts)
		if table == nil {
			t.Fatal("NewProjectDataTable() returned nil")
		}

		emptyOpts := DataTableOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		table2 := NewProjectDataTable(repo, emptyOpts)
		if table2 == nil {
			t.Fatal("NewProjectDataTable() with empty opts returned nil")
		}

		err := table.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}
	})

	t.Run("NewProjectListFromTable", func(t *testing.T) {
		repo := &mockProjectRepository{
			projects: []repo.ProjectSummary{
				{Name: "Test Project", TaskCount: 1},
			},
		}

		output := &bytes.Buffer{}
		input := strings.NewReader("q\n")

		table := NewProjectListFromTable(repo, output, input, true)
		if table == nil {
			t.Fatal("NewProjectListFromTable() returned nil")
		}

		err := table.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}

		outputStr := output.String()
		if !strings.Contains(outputStr, "Projects") {
			t.Error("Output should contain 'Projects' title")
		}
		if !strings.Contains(outputStr, "Test Project") {
			t.Error("Output should contain project name")
		}
		if !strings.Contains(outputStr, "1 task") {
			t.Error("Output should contain formatted task count")
		}
	})
}
