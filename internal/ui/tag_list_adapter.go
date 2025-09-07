package ui

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// TagRepository interface for dependency injection in tests
type TagRepository interface {
	GetTags(ctx context.Context) ([]repo.TagSummary, error)
}


// TagSummaryRecord adapts repo.TagSummary to work with DataTable
type TagSummaryRecord struct {
	summary repo.TagSummary
}

func (t *TagSummaryRecord) GetField(name string) any {
	switch name {
	case "name":
		return t.summary.Name
	case "task_count":
		return t.summary.TaskCount
	default:
		return ""
	}
}

func (t *TagSummaryRecord) GetTableName() string {
	return "tags"
}

// Use task count as pseudo-ID since tags don't have IDs
func (t *TagSummaryRecord) GetID() int64                { return int64(t.summary.TaskCount) }
func (t *TagSummaryRecord) SetID(id int64)              {}
func (t *TagSummaryRecord) GetCreatedAt() time.Time     { return time.Time{} }
func (t *TagSummaryRecord) SetCreatedAt(time time.Time) {}
func (t *TagSummaryRecord) GetUpdatedAt() time.Time     { return time.Time{} }
func (t *TagSummaryRecord) SetUpdatedAt(time time.Time) {}

// TagDataSource adapts TagRepository to work with DataTable
type TagDataSource struct {
	repo TagRepository
}

func (t *TagDataSource) Load(ctx context.Context, opts DataOptions) ([]DataRecord, error) {
	tags, err := t.repo.GetTags(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]DataRecord, len(tags))
	for i, tag := range tags {
		records[i] = &TagSummaryRecord{summary: tag}
	}

	return records, nil
}

func (t *TagDataSource) Count(ctx context.Context, opts DataOptions) (int, error) {
	tags, err := t.repo.GetTags(ctx)
	if err != nil {
		return 0, err
	}
	return len(tags), nil
}

// NewTagDataTable creates a new DataTable for browsing tags
func NewTagDataTable(repo TagRepository, opts DataTableOptions) *DataTable {
	if opts.Title == "" {
		opts.Title = "Tags"
	}

	if len(opts.Fields) == 0 {
		opts.Fields = []Field{
			{
				Name:  "name",
				Title: "Tag Name",
				Width: 25,
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

	source := &TagDataSource{repo: repo}
	return NewDataTable(source, opts)
}

// NewTagListFromTable creates a TagList-compatible interface using DataTable
func NewTagListFromTable(repo TagRepository, output io.Writer, input io.Reader, static bool) *DataTable {
	opts := DataTableOptions{
		Output: output,
		Input:  input,
		Static: static,
		Title:  "Tags",
	}
	return NewTagDataTable(repo, opts)
}
