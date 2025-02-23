package processor

import (
	"indexer/internal/pkg/models"
)

// Enricher defines the interface for adding additional metadata to a document.
type Enricher interface {
	Enrich(pd *models.PageData, doc *models.Document) error
}

// nlpEnricher is a naive implementation of Enricher.
type nlpEnricher struct{}

// NewNLPEnricher creates a new instance of an NLP-based Enricher.
func NewNLPEnricher() Enricher {
	return &nlpEnricher{}
}

// Enrich augments the document with entities, keywords, and a summary.
func (e *nlpEnricher) Enrich(pd *models.PageData, doc *models.Document) error {
	// Extract entities.
	doc.Entities = extractEntities(pd.VisibleText)
	// Extract keywords.
	doc.Keywords = extractKeywords(pd.VisibleText)
	// Generate summary.
	doc.Summary = summarize(pd.VisibleText)
	return nil
}

// extractEntities extracts entities from text (placeholder implementation).
func extractEntities(text string) []string {
	return []string{"entity1", "entity2"}
}

// extractKeywords extracts keywords from text (placeholder implementation).
func extractKeywords(text string) []string {
	return []string{"keyword1", "keyword2"}
}

// summarize generates a summary from text (placeholder implementation).
func summarize(text string) string {
	return "Placeholder summary"
}

// tokenize splits text into tokens (unused, placeholder).
func tokenize(text string) []string {
	return []string{"token1", "token2"}
}

// splitIntoSentences splits text into sentences (unused, placeholder).
func splitIntoSentences(text string) []string {
	return []string{"Sentence 1", "Sentence 2"}
}
