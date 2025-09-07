package ui

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// ProjectRepository interface for dependency injection in tests
type ProjectRepository interface {
	GetProjects(ctx context.Context) ([]repo.ProjectSummary, error)
}


func pluralizeCount(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// ProjectSummaryRecord adapts repo.ProjectSummary to work with DataTable
type ProjectSummaryRecord struct {
	summary repo.ProjectSummary
}

func (p *ProjectSummaryRecord) GetField(name string) any {
	switch name {
	case "name":
		return p.summary.Name
	case "task_count":
		return p.summary.TaskCount
	default:
		return ""
	}
}

func (p *ProjectSummaryRecord) GetTableName() string {
	return "projects"
}

// Use task count as pseudo-ID since projects don't have IDs
func (p *ProjectSummaryRecord) GetID() int64             { return int64(p.summary.TaskCount) }
func (p *ProjectSummaryRecord) SetID(id int64)           {}
func (p *ProjectSummaryRecord) GetCreatedAt() time.Time  { return time.Time{} }
func (p *ProjectSummaryRecord) SetCreatedAt(t time.Time) {}
func (p *ProjectSummaryRecord) GetUpdatedAt() time.Time  { return time.Time{} }
func (p *ProjectSummaryRecord) SetUpdatedAt(t time.Time) {}

// ProjectDataSource adapts ProjectRepository to work with DataTable
type ProjectDataSource struct {
	repo ProjectRepository
}

func (p *ProjectDataSource) Load(ctx context.Context, opts DataOptions) ([]DataRecord, error) {
	projects, err := p.repo.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]DataRecord, len(projects))
	for i, project := range projects {
		records[i] = &ProjectSummaryRecord{summary: project}
	}

	return records, nil
}

func (p *ProjectDataSource) Count(ctx context.Context, opts DataOptions) (int, error) {
	projects, err := p.repo.GetProjects(ctx)
	if err != nil {
		return 0, err
	}
	return len(projects), nil
}

// NewProjectDataTable creates a new DataTable for browsing projects
func NewProjectDataTable(repo ProjectRepository, opts DataTableOptions) *DataTable {
	if opts.Title == "" {
		opts.Title = "Projects"
	}

	if len(opts.Fields) == 0 {
		opts.Fields = []Field{
			{
				Name:  "name",
				Title: "Project Name",
				Width: 30,
			},
			{
				Name:  "task_count",
				Title: "Task Count",
				Width: 15,
				Formatter: func(value any) string {
					if count, ok := value.(int); ok {
						return fmt.Sprintf("%d task%s", count, pluralizeCount(count))
					}
					return fmt.Sprintf("%v", value)
				},
			},
		}
	}

	source := &ProjectDataSource{repo: repo}
	return NewDataTable(source, opts)
}

// NewProjectListFromTable creates a ProjectList-compatible interface using DataTable
func NewProjectListFromTable(repo ProjectRepository, output io.Writer, input io.Reader, static bool) *DataTable {
	opts := DataTableOptions{
		Output: output,
		Input:  input,
		Static: static,
		Title:  "Projects",
	}
	return NewProjectDataTable(repo, opts)
}
