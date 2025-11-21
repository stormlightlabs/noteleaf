package documents

import (
	"testing"
	"time"
)

func TestTokenizer(t *testing.T) {
	tokenizer := NewTokenizer()

	t.Run("Basic tokenization", func(t *testing.T) {
		t.Run("tokenizes simple text", func(t *testing.T) {
			tokens := tokenizer.Tokenize("Hello World")
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens, got %d", len(tokens))
			}
			if tokens[0] != "hello" || tokens[1] != "world" {
				t.Errorf("expected [hello world], got %v", tokens)
			}
		})

		t.Run("lowercases all tokens", func(t *testing.T) {
			tokens := tokenizer.Tokenize("UPPERCASE MiXeD lowercase")
			if len(tokens) != 3 {
				t.Fatalf("expected 3 tokens, got %d", len(tokens))
			}
			for _, token := range tokens {
				if token != "uppercase" && token != "mixed" && token != "lowercase" {
					t.Errorf("unexpected token: %s", token)
				}
			}
		})

		t.Run("handles punctuation", func(t *testing.T) {
			tokens := tokenizer.Tokenize("Hello, world! How are you?")
			expected := []string{"hello", "world", "how", "are", "you"}
			if len(tokens) != len(expected) {
				t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
			}
			for i, token := range tokens {
				if token != expected[i] {
					t.Errorf("token %d: expected %s, got %s", i, expected[i], token)
				}
			}
		})
	})

	t.Run("Unicode support", func(t *testing.T) {
		t.Run("tokenizes unicode characters", func(t *testing.T) {
			tokens := tokenizer.Tokenize("café résumé naïve")
			if len(tokens) != 3 {
				t.Fatalf("expected 3 tokens, got %d", len(tokens))
			}
		})

		t.Run("handles emoji and special characters", func(t *testing.T) {
			tokens := tokenizer.Tokenize("hello 😀 world")
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens (emoji excluded), got %d", len(tokens))
			}
			if tokens[0] != "hello" || tokens[1] != "world" {
				t.Errorf("expected [hello world], got %v", tokens)
			}
		})

		t.Run("tokenizes CJK characters", func(t *testing.T) {
			tokens := tokenizer.Tokenize("你好 世界")
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens, got %d", len(tokens))
			}
		})
	})

	t.Run("Numbers", func(t *testing.T) {
		t.Run("tokenizes numbers", func(t *testing.T) {
			tokens := tokenizer.Tokenize("test 123 456")
			if len(tokens) != 3 {
				t.Fatalf("expected 3 tokens, got %d", len(tokens))
			}
			if tokens[1] != "123" || tokens[2] != "456" {
				t.Errorf("expected numbers to be tokenized, got %v", tokens)
			}
		})

		t.Run("handles mixed alphanumeric", func(t *testing.T) {
			tokens := tokenizer.Tokenize("version 2 released")
			if len(tokens) != 3 {
				t.Fatalf("expected 3 tokens, got %d", len(tokens))
			}
		})
	})

	t.Run("Edge cases", func(t *testing.T) {
		t.Run("handles empty string", func(t *testing.T) {
			tokens := tokenizer.Tokenize("")
			if len(tokens) != 0 {
				t.Errorf("expected 0 tokens for empty string, got %d", len(tokens))
			}
		})

		t.Run("handles whitespace only", func(t *testing.T) {
			tokens := tokenizer.Tokenize("   \t\n  ")
			if len(tokens) != 0 {
				t.Errorf("expected 0 tokens for whitespace, got %d", len(tokens))
			}
		})

		t.Run("handles punctuation only", func(t *testing.T) {
			tokens := tokenizer.Tokenize("!@#$%^&*()")
			if len(tokens) != 0 {
				t.Errorf("expected 0 tokens for punctuation only, got %d", len(tokens))
			}
		})
	})
}

func TestTokenFrequency(t *testing.T) {
	t.Run("counts term frequencies", func(t *testing.T) {
		tokens := []string{"hello", "world", "hello", "test"}
		freq := TokenFrequency(tokens)

		if freq["hello"] != 2 {
			t.Errorf("expected hello frequency 2, got %d", freq["hello"])
		}
		if freq["world"] != 1 {
			t.Errorf("expected world frequency 1, got %d", freq["world"])
		}
		if freq["test"] != 1 {
			t.Errorf("expected test frequency 1, got %d", freq["test"])
		}
	})

	t.Run("handles empty token list", func(t *testing.T) {
		freq := TokenFrequency([]string{})
		if len(freq) != 0 {
			t.Errorf("expected empty frequency map, got %d entries", len(freq))
		}
	})

	t.Run("handles single token", func(t *testing.T) {
		freq := TokenFrequency([]string{"single"})
		if freq["single"] != 1 {
			t.Errorf("expected frequency 1, got %d", freq["single"])
		}
	})
}

func TestBuildIndex(t *testing.T) {
	now := time.Now()

	t.Run("builds index from documents", func(t *testing.T) {
		docs := []Document{
			{ID: 1, Title: "Go Programming", Body: "Go is a great language", CreatedAt: now, DocKind: int64(NoteDoc)},
			{ID: 2, Title: "Python Guide", Body: "Python is versatile", CreatedAt: now, DocKind: int64(ArticleDoc)},
		}

		idx := BuildIndex(docs)

		if idx.NumDocs != 2 {
			t.Errorf("expected NumDocs 2, got %d", idx.NumDocs)
		}

		if len(idx.DocLengths) != 2 {
			t.Errorf("expected 2 document lengths, got %d", len(idx.DocLengths))
		}

		if idx.DocLengths[1] <= 0 || idx.DocLengths[2] <= 0 {
			t.Error("document lengths should be positive")
		}

		if _, exists := idx.Postings["go"]; !exists {
			t.Error("expected 'go' to be in postings")
		}
		if _, exists := idx.Postings["python"]; !exists {
			t.Error("expected 'python' to be in postings")
		}
	})

	t.Run("handles empty document list", func(t *testing.T) {
		idx := BuildIndex([]Document{})
		if idx.NumDocs != 0 {
			t.Errorf("expected NumDocs 0, got %d", idx.NumDocs)
		}
		if len(idx.Postings) != 0 {
			t.Errorf("expected empty postings, got %d entries", len(idx.Postings))
		}
	})

	t.Run("calculates term frequencies correctly", func(t *testing.T) {
		docs := []Document{
			{ID: 1, Title: "test", Body: "test test test", CreatedAt: now, DocKind: int64(NoteDoc)},
		}

		idx := BuildIndex(docs)

		postings := idx.Postings["test"]
		if len(postings) != 1 {
			t.Fatalf("expected 1 posting for 'test', got %d", len(postings))
		}

		if postings[0].TF != 4 {
			t.Errorf("expected TF 4 (title + 3 in body), got %d", postings[0].TF)
		}
	})

	t.Run("builds postings for multiple documents with same term", func(t *testing.T) {
		docs := []Document{
			{ID: 1, Title: "Go", Body: "Go is great", CreatedAt: now, DocKind: int64(NoteDoc)},
			{ID: 2, Title: "Go Tutorial", Body: "Learn Go", CreatedAt: now, DocKind: int64(NoteDoc)},
		}

		idx := BuildIndex(docs)

		postings := idx.Postings["go"]
		if len(postings) != 2 {
			t.Fatalf("expected 2 postings for 'go', got %d", len(postings))
		}
	})
}

func TestIndexSearch(t *testing.T) {
	now := time.Now()

	t.Run("Search functionality", func(t *testing.T) {
		t.Run("returns empty results for empty query", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "Test", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("", 10)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != 0 {
				t.Errorf("expected 0 results for empty query, got %d", len(results))
			}
		})

		t.Run("finds matching documents", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "Go Programming", Body: "Learn Go language", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 2, Title: "Python Guide", Body: "Python is versatile", CreatedAt: now, DocKind: int64(ArticleDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("go", 10)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}

			if results[0].DocID != 1 {
				t.Errorf("expected DocID 1, got %d", results[0].DocID)
			}

			if results[0].Score <= 0 {
				t.Error("expected positive score")
			}
		})

		t.Run("ranks documents by relevance", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "Go", Body: "tutorial python rust", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 2, Title: "Go Programming", Body: "advanced go tutorial", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 3, Title: "Python", Body: "different language", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("go", 10)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 2 {
				t.Fatalf("expected 2 results, got %d", len(results))
			}

			if results[0].DocID != 2 {
				t.Errorf("expected document 2 to rank higher (has more 'go' terms)")
			}

			if results[0].Score <= results[1].Score {
				t.Errorf("expected first result to have higher score, got %f <= %f", results[0].Score, results[1].Score)
			}
		})

		t.Run("respects limit parameter", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "test one", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 2, Title: "test two", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 3, Title: "test three", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("test", 2)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("expected 2 results with limit=2, got %d", len(results))
			}
		})

		t.Run("handles multi-term queries", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "Go Programming", Body: "advanced tutorial", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 2, Title: "Go Basics", Body: "beginner tutorial", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 3, Title: "Python", Body: "different language", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("go tutorial", 10)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("expected 2 results, got %d", len(results))
			}
		})

		t.Run("returns no results for non-matching query", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "Go", Body: "programming", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("rust", 10)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 0 {
				t.Errorf("expected 0 results for non-matching query, got %d", len(results))
			}
		})

		t.Run("handles zero limit", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "test", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("test", 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("expected all results with limit=0, got %d", len(results))
			}
		})

		t.Run("tie-breaking uses DocID", func(t *testing.T) {
			docs := []Document{
				{ID: 1, Title: "test", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
				{ID: 2, Title: "test", Body: "content", CreatedAt: now, DocKind: int64(NoteDoc)},
			}
			idx := BuildIndex(docs)

			results, err := idx.Search("test", 10)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != 2 {
				t.Fatalf("expected 2 results, got %d", len(results))
			}

			if results[0].DocID <= results[1].DocID {
				t.Error("expected higher DocID first when scores are equal")
			}
		})
	})
}
