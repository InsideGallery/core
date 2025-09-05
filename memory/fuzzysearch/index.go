package fuzzysearch

import (
	"strings"
	"unicode"

	"github.com/InsideGallery/core/memory/set"
)

// Index search index
type Index map[string][]int

// NewIndex return new index
func NewIndex() Index {
	return Index{}
}

// Document document
type Document struct {
	Text string
	ID   int
}

// NewDocument return new document
func NewDocument(id int, text string) Document {
	return Document{
		ID:   id,
		Text: text,
	}
}

func (d Document) Terms() []string {
	return Analyze(d.Text)
}

// Add add documents to index
func (idx Index) Add(docs ...Document) {
	for _, doc := range docs {
		for _, token := range Analyze(doc.Text) {
			ids := idx[token]
			if ids != nil && ids[len(ids)-1] == doc.ID {
				// Don't add same ID twice.
				continue
			}

			idx[token] = append(ids, doc.ID)
		}
	}
}

// Remove remove documents from index
func (idx Index) Remove(docs ...Document) {
	for _, doc := range docs {
		for _, token := range Analyze(doc.Text) {
			ids := idx[token]
			if ids == nil {
				continue
			}

			for i := range ids {
				if i == doc.ID {
					idx[token] = append(idx[token][:i], idx[token][i+1:]...)
					break
				}
			}
		}
	}
}

// Search search documents
func (idx Index) Search(text string) []int {
	var r []int

	for _, token := range Analyze(text) {
		if ids, ok := idx[token]; ok {
			if r == nil {
				r = ids
			} else {
				r = Intersection(r, ids)
			}
		} else {
			// Token doesn't exist.
			return nil
		}
	}

	return r
}

// Intersection intersect two slices
func Intersection(a []int, b []int) []int {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	r := make([]int, 0, maxLen)

	var i, j int

	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			i++
		case a[i] > b[j]:
			j++
		default:
			r = append(r, a[i])
			i++
			j++
		}
	}

	return r
}

// Analyze analyze string
func Analyze(text string) []string {
	tokens := Tokenize(text)
	tokens = LowercaseFilter(tokens)

	return tokens
}

// Tokenize tokenize string
func Tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		// Split on any character that is not a letter or a number.
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

// LowercaseFilter lowercase all tokens
func LowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}

	return r
}

// StopwordFilter filer of stopworld
func StopwordFilter(tokens []string, stopwords set.GenericDataSet[string]) []string {
	r := make([]string, 0, len(tokens))

	for _, token := range tokens {
		if !stopwords.Contains(token) {
			r = append(r, token)
		}
	}

	return r
}
