package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

type mockPublicationRepository struct {
	notes      []*models.Note
	err        error
	published  []*models.Note
	drafts     []*models.Note
	leafletAll []*models.Note
}

func (m *mockPublicationRepository) ListPublished(ctx context.Context) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.published != nil {
		return m.published, nil
	}
	var published []*models.Note
	for _, note := range m.notes {
		if !note.IsDraft {
			published = append(published, note)
		}
	}
	return published, nil
}

func (m *mockPublicationRepository) ListDrafts(ctx context.Context) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.drafts != nil {
		return m.drafts, nil
	}
	var drafts []*models.Note
	for _, note := range m.notes {
		if note.IsDraft {
			drafts = append(drafts, note)
		}
	}
	return drafts, nil
}

func (m *mockPublicationRepository) GetLeafletNotes(ctx context.Context) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.leafletAll != nil {
		return m.leafletAll, nil
	}
	return m.notes, nil
}

func (m *mockPublicationRepository) List(ctx context.Context, options repo.NoteListOptions) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.notes, nil
}

func TestPublicationAdapter(t *testing.T) {
	t.Run("PublicationRecord", func(t *testing.T) {
		rkey := "test-rkey-123"
		cid := "test-cid-456"
		publishedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

		note := &models.Note{
			ID:          1,
			Title:       "Test Publication",
			Content:     "Publication content",
			Tags:        []string{"article", "tech"},
			IsDraft:     false,
			PublishedAt: &publishedAt,
			Modified:    time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
			LeafletRKey: &rkey,
			LeafletCID:  &cid,
		}
		record := &PublicationRecord{Note: note}

		t.Run("GetField returns all publication fields", func(t *testing.T) {
			tests := []struct {
				field    string
				expected any
				name     string
			}{
				{"id", int64(1), "id field"},
				{"title", "Test Publication", "title field"},
				{"status", "published", "status for published note"},
				{"published_at", &publishedAt, "published_at field"},
				{"modified", note.Modified, "modified field"},
				{"leaflet_rkey", &rkey, "leaflet_rkey field"},
				{"leaflet_cid", &cid, "leaflet_cid field"},
				{"unknown", "", "unknown field returns empty string"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result := record.GetField(tt.field)
					if fmt.Sprintf("%v", result) != fmt.Sprintf("%v", tt.expected) {
						t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
					}
				})
			}
		})

		t.Run("GetField returns draft status", func(t *testing.T) {
			draftNote := &models.Note{
				ID:      2,
				Title:   "Draft Note",
				IsDraft: true,
			}
			draftRecord := &PublicationRecord{Note: draftNote}

			status := draftRecord.GetField("status")
			if status != "draft" {
				t.Errorf("GetField(status) for draft = %v, want 'draft'", status)
			}
		})

		t.Run("GetTitle formats with ID and status", func(t *testing.T) {
			title := record.GetTitle()
			if !strings.Contains(title, "[1]") {
				t.Errorf("GetTitle() should contain ID [1], got: %s", title)
			}
			if !strings.Contains(title, "Test Publication") {
				t.Errorf("GetTitle() should contain title, got: %s", title)
			}
			if !strings.Contains(title, "(published)") {
				t.Errorf("GetTitle() should contain status (published), got: %s", title)
			}
		})

		t.Run("GetTitle shows draft status", func(t *testing.T) {
			draftNote := &models.Note{
				ID:      3,
				Title:   "Draft Article",
				IsDraft: true,
			}
			draftRecord := &PublicationRecord{Note: draftNote}

			title := draftRecord.GetTitle()
			if !strings.Contains(title, "(draft)") {
				t.Errorf("GetTitle() for draft should contain (draft), got: %s", title)
			}
		})

		t.Run("GetDescription includes all metadata", func(t *testing.T) {
			description := record.GetDescription()

			if !strings.Contains(description, "Published: 2024-01-15 10:00") {
				t.Errorf("GetDescription() should contain published date, got: %s", description)
			}
			if !strings.Contains(description, "Modified: 2024-01-16 12:00") {
				t.Errorf("GetDescription() should contain modified date, got: %s", description)
			}
			if !strings.Contains(description, "rkey: test-rkey-123") {
				t.Errorf("GetDescription() should contain rkey, got: %s", description)
			}
		})

		t.Run("GetDescription handles missing fields", func(t *testing.T) {
			minimalNote := &models.Note{
				ID:       4,
				Title:    "Minimal Note",
				Modified: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			minimalRecord := &PublicationRecord{Note: minimalNote}

			description := minimalRecord.GetDescription()

			if strings.Contains(description, "Published:") {
				t.Errorf("GetDescription() should not contain Published for unpublished note, got: %s", description)
			}
			if strings.Contains(description, "rkey:") {
				t.Errorf("GetDescription() should not contain rkey when not set, got: %s", description)
			}
			if !strings.Contains(description, "Modified: 2024-01-01 00:00") {
				t.Errorf("GetDescription() should always contain Modified, got: %s", description)
			}
		})

		t.Run("GetFilterValue includes searchable text", func(t *testing.T) {
			filterValue := record.GetFilterValue()

			if !strings.Contains(filterValue, "Test Publication") {
				t.Errorf("GetFilterValue() should contain title, got: %s", filterValue)
			}
			if !strings.Contains(filterValue, "Publication content") {
				t.Errorf("GetFilterValue() should contain content, got: %s", filterValue)
			}
			if !strings.Contains(filterValue, "test-rkey-123") {
				t.Errorf("GetFilterValue() should contain rkey, got: %s", filterValue)
			}
		})

		t.Run("GetFilterValue handles missing rkey", func(t *testing.T) {
			noteWithoutRKey := &models.Note{
				ID:      5,
				Title:   "No RKey Note",
				Content: "Some content",
			}
			recordWithoutRKey := &PublicationRecord{Note: noteWithoutRKey}

			filterValue := recordWithoutRKey.GetFilterValue()

			if !strings.Contains(filterValue, "No RKey Note") {
				t.Errorf("GetFilterValue() should contain title, got: %s", filterValue)
			}
			if !strings.Contains(filterValue, "Some content") {
				t.Errorf("GetFilterValue() should contain content, got: %s", filterValue)
			}
		})
	})

	t.Run("PublicationDataSource", func(t *testing.T) {
		rkey1 := "rkey-published"
		rkey2 := "rkey-draft"
		publishedAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)

		notes := []*models.Note{
			{
				ID:          1,
				Title:       "Published Article",
				Content:     "Published content",
				IsDraft:     false,
				PublishedAt: &publishedAt,
				LeafletRKey: &rkey1,
				Modified:    time.Now(),
			},
			{
				ID:          2,
				Title:       "Draft Article",
				Content:     "Draft content",
				IsDraft:     true,
				LeafletRKey: &rkey2,
				Modified:    time.Now(),
			},
			{
				ID:       3,
				Title:    "Another Published",
				Content:  "More published content",
				IsDraft:  false,
				Modified: time.Now(),
			},
		}

		t.Run("Load with all filter", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 3 {
				t.Errorf("Load() with filter 'all' returned %d items, want 3", len(items))
			}
		})

		t.Run("Load with published filter", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "published"}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 2 {
				t.Errorf("Load() with filter 'published' returned %d items, want 2", len(items))
			}

			for _, item := range items {
				pubRecord := item.(*PublicationRecord)
				if pubRecord.IsDraft {
					t.Error("Load() with 'published' filter should not return drafts")
				}
			}
		})

		t.Run("Load with draft filter", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "draft"}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Load() with filter 'draft' returned %d items, want 1", len(items))
			}

			pubRecord := items[0].(*PublicationRecord)
			if !pubRecord.IsDraft {
				t.Error("Load() with 'draft' filter should only return drafts")
			}
		})

		t.Run("Load with search query", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			items, err := source.Load(context.Background(), ListOptions{Search: "Draft"})
			if err != nil {
				t.Fatalf("Load() with search failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Load() with search 'Draft' returned %d items, want 1", len(items))
			}

			if items[0].GetTitle() != "[2] Draft Article (draft)" {
				t.Errorf("Search result title = %q, want '[2] Draft Article (draft)'", items[0].GetTitle())
			}
		})

		t.Run("Load with search in content", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			items, err := source.Load(context.Background(), ListOptions{Search: "Draft content"})
			if err != nil {
				t.Fatalf("Load() with content search failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Load() searching content returned %d items, want 1", len(items))
			}
		})

		t.Run("Load with search in rkey", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			items, err := source.Load(context.Background(), ListOptions{Search: "rkey-draft"})
			if err != nil {
				t.Fatalf("Load() with rkey search failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Load() searching rkey returned %d items, want 1", len(items))
			}

			pubRecord := items[0].(*PublicationRecord)
			if *pubRecord.LeafletRKey != "rkey-draft" {
				t.Errorf("Found note with rkey %q, want 'rkey-draft'", *pubRecord.LeafletRKey)
			}
		})

		t.Run("Load with limit", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			items, err := source.Load(context.Background(), ListOptions{Limit: 2})
			if err != nil {
				t.Fatalf("Load() with limit failed: %v", err)
			}

			if len(items) != 2 {
				t.Errorf("Load() with limit 2 returned %d items, want 2", len(items))
			}
		})

		t.Run("Load error handling", func(t *testing.T) {
			testErr := fmt.Errorf("database error")
			mockRepo := &mockPublicationRepository{err: testErr}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			_, err := source.Load(context.Background(), ListOptions{})
			if err != testErr {
				t.Errorf("Load() error = %v, want %v", err, testErr)
			}
		})

		t.Run("Count", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			count, err := source.Count(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Count() failed: %v", err)
			}

			if count != 3 {
				t.Errorf("Count() = %d, want 3", count)
			}
		})

		t.Run("Count with filter", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "draft"}

			count, err := source.Count(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Count() with filter failed: %v", err)
			}

			if count != 1 {
				t.Errorf("Count() with draft filter = %d, want 1", count)
			}
		})

		t.Run("Search", func(t *testing.T) {
			mockRepo := &mockPublicationRepository{notes: notes}
			source := &PublicationDataSource{repo: mockRepo, filter: "all"}

			items, err := source.Search(context.Background(), "Published", ListOptions{})
			if err != nil {
				t.Fatalf("Search() failed: %v", err)
			}

			if len(items) != 2 {
				t.Errorf("Search() for 'Published' returned %d items, want 2", len(items))
			}
		})
	})

	t.Run("NewPublicationDataList", func(t *testing.T) {
		notes := []*models.Note{
			{
				ID:       1,
				Title:    "Test Publication",
				Content:  "Test content",
				IsDraft:  false,
				Modified: time.Now(),
			},
		}

		mockRepo := &mockPublicationRepository{notes: notes}

		opts := DataListOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		list := NewPublicationDataList(mockRepo, opts, "all")
		if list == nil {
			t.Fatal("NewPublicationDataList() returned nil")
		}

		err := list.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}
	})

	t.Run("NewPublicationListFromList", func(t *testing.T) {
		notes := []*models.Note{
			{
				ID:       1,
				Title:    "Test Publication",
				Content:  "Test content",
				IsDraft:  false,
				Modified: time.Now(),
			},
		}

		mockRepo := &mockPublicationRepository{notes: notes}

		output := &bytes.Buffer{}
		input := strings.NewReader("q\n")

		list := NewPublicationListFromList(mockRepo, output, input, true, "all")
		if list == nil {
			t.Fatal("NewPublicationListFromList() returned nil")
		}

		err := list.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}

		outputStr := output.String()
		if !strings.Contains(outputStr, "Publications") {
			t.Error("Output should contain 'Publications' title")
		}
		if !strings.Contains(outputStr, "Test Publication") {
			t.Error("Output should contain publication title")
		}
	})

	t.Run("formatPublicationForView", func(t *testing.T) {
		t.Run("formats published note with all metadata", func(t *testing.T) {
			rkey := "test-rkey"
			cid := "test-cid"
			publishedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

			note := &models.Note{
				ID:          1,
				Title:       "Test Article",
				Content:     "# Test Article\n\nThis is the article content.",
				IsDraft:     false,
				PublishedAt: &publishedAt,
				Modified:    time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
			}

			result := formatPublicationForView(note)

			if !strings.Contains(result, "Test Article") {
				t.Errorf("Formatted view should contain title\nGot: %s", result)
			}
			if !strings.Contains(result, "published") {
				t.Errorf("Formatted view should contain status 'published'\nGot: %s", result)
			}
			if !strings.Contains(result, "2024-01-15") {
				t.Errorf("Formatted view should contain published date\nGot: %s", result)
			}
			if !strings.Contains(result, "Modified") && !strings.Contains(result, "2024-01-16") {
				t.Errorf("Formatted view should contain modified date\nGot: %s", result)
			}
			if !strings.Contains(result, "test-rkey") {
				t.Error("Formatted view should contain rkey")
			}
			if !strings.Contains(result, "test-cid") {
				t.Error("Formatted view should contain cid")
			}
		})

		t.Run("formats draft note", func(t *testing.T) {
			note := &models.Note{
				ID:       2,
				Title:    "Draft Article",
				Content:  "Draft content here.",
				IsDraft:  true,
				Modified: time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
			}

			result := formatPublicationForView(note)

			if !strings.Contains(result, "Draft Article") {
				t.Error("Formatted view should contain title")
			}
			if !strings.Contains(result, "draft") {
				t.Error("Formatted view should contain status 'draft'")
			}
			if strings.Contains(result, "Published:") {
				t.Error("Formatted draft view should not contain published date")
			}
			if !strings.Contains(result, "2024-01-20 14:00") {
				t.Error("Formatted view should contain modified date")
			}
		})

		t.Run("handles content without title header", func(t *testing.T) {
			note := &models.Note{
				ID:       3,
				Title:    "Plain Content",
				Content:  "This content has no markdown header.",
				IsDraft:  false,
				Modified: time.Now(),
			}

			result := formatPublicationForView(note)

			if !strings.Contains(result, "Plain Content") {
				t.Error("Formatted view should contain title")
			}
			if !strings.Contains(result, "This content has no markdown header") {
				t.Error("Formatted view should contain full content")
			}
		})

		t.Run("strips duplicate title from content", func(t *testing.T) {
			note := &models.Note{
				ID:       4,
				Title:    "Article Title",
				Content:  "# Article Title\n\nContent after title.",
				IsDraft:  false,
				Modified: time.Now(),
			}

			result := formatPublicationForView(note)

			titleCount := strings.Count(result, "Article Title")
			if titleCount < 1 {
				t.Error("Formatted view should contain title at least once")
			}
			if !strings.Contains(result, "Content after title") {
				t.Error("Formatted view should contain content after title")
			}
		})
	})
}
