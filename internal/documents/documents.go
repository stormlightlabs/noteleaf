// Term Frequency-Inverse Document Frequency search model for notes
package documents

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

type DocKind int64

const (
	NoteDoc DocKind = iota
	ArticleDoc
	MovieDoc
	BookDoc
	TVDoc
)

type Document struct {
	ID        int64
	Title     string
	Body      string
	CreatedAt time.Time
	DocKind   int64
}

type Posting struct {
	DocID int64
	TF    int
}

type Index struct {
	Postings   map[string][]Posting
	DocLengths map[int64]int
	NumDocs    int
}

type Result struct {
	DocID int64
	Score float64
}

type Searchable interface {
	Search(query string, limit int) ([]Result, error)
}

// Tokenizer handles text tokenization and normalization
type Tokenizer struct {
	pattern *regexp.Regexp
}

// NewTokenizer creates a new tokenizer with Unicode-aware word/number matching
func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		pattern: regexp.MustCompile(`\p{L}+\p{M}*|\p{N}+`),
	}
}

// Tokenize splits text into normalized tokens (lowercase words and numbers)
func (t *Tokenizer) Tokenize(text string) []string {
	lowered := strings.ToLower(text)
	return t.pattern.FindAllString(lowered, -1)
}

// TokenFrequency computes term frequency map for tokens
func TokenFrequency(tokens []string) map[string]int {
	freq := make(map[string]int)
	for _, token := range tokens {
		freq[token]++
	}
	return freq
}

// BuildIndex constructs a TF-IDF index from a collection of documents
func BuildIndex(docs []Document) *Index {
	idx := &Index{
		Postings:   make(map[string][]Posting),
		DocLengths: make(map[int64]int),
		NumDocs:    0,
	}

	tokenizer := NewTokenizer()

	for _, doc := range docs {
		text := doc.Title + " " + doc.Body
		tokens := tokenizer.Tokenize(text)

		idx.NumDocs++
		idx.DocLengths[doc.ID] = len(tokens)

		freq := TokenFrequency(tokens)

		for term, tf := range freq {
			idx.Postings[term] = append(idx.Postings[term], Posting{
				DocID: doc.ID,
				TF:    tf,
			})
		}
	}

	return idx
}

// Search performs TF-IDF ranked search on the index
func (idx *Index) Search(query string, limit int) ([]Result, error) {
	tokenizer := NewTokenizer()
	queryTokens := tokenizer.Tokenize(query)

	if len(queryTokens) == 0 {
		return []Result{}, nil
	}

	scores := make(map[int64]float64)

	for _, term := range queryTokens {
		postings, exists := idx.Postings[term]
		if !exists {
			continue
		}

		df := len(postings)
		idf := math.Log(float64(idx.NumDocs) / float64(df))

		for _, posting := range postings {
			tf := float64(posting.TF)
			scores[posting.DocID] += tf * idf
		}
	}

	results := make([]Result, 0, len(scores))
	for docID, score := range scores {
		results = append(results, Result{
			DocID: docID,
			Score: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].DocID > results[j].DocID
	})

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
