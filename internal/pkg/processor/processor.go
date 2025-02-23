package processor

import (
    "errors"
    "net/url"
    "strings"
    "log"
	// "github.com/microcosm-cc/bluemonday"
    // "golang.org/x/net/html"
    "indexer/internal/pkg/models"

)

// Defines the high-level interface for processing page data.
type Processor interface {
	// Process runs the complete data processing pipeline.
	// It operates directly on the provided PageData and Document.
	Process(pd *models.PageData, doc *models.Document) error
}

// The default implementation of Processor.
type processor struct {
	deduper  Deduper
	enricher Enricher
}

// Creates a new Processor instance and wires in the sub‑components.
func NewProcessor() Processor {
	return &processor{
		deduper:  NewDeduper(),
		enricher: NewNLPEnricher(),
	}
}

// Runs the data processing pipeline:
// cleaning/normalization, deduplication, and enrichment.
func (processor *processor) Process(pageData *models.PageData, doc *models.Document) error {
	// Clean and normalize the data.
	if err := cleanAndNormalize(pageData, doc); err != nil {
		return err
	}

	// Deduplication: generate a signature from the visible text.
	signature := GenerateSignature(pageData.VisibleText)
	if processor.deduper.IsDuplicate(signature) {
		return errors.New("duplicate page detected")
	}

	// Enrich the document.
	if err := processor.enricher.Enrich(pageData, doc); err != nil {
		return err
	}

	// Mark the page as processed.
	processor.deduper.StoreSignature(signature)
	return nil
}

// cleanAndNormalize applies cleaning, URL normalization, language detection,
// and spam filtering. It updates the PageData and Document in place.
func cleanAndNormalize(pageData *models.PageData, doc *models.Document) error {
	// Basic HTML cleanup.
	doc.VisibleText = basicHTMLCleanup(pageData.VisibleText)

	// Normalize primary URL.
	var err error
	doc.URL, err = normalizeURL(pageData.URL)
	if err != nil {
		log.Printf("invalid URL %q: %v", pageData.URL, err)
		return err
	}

	// Normalize canonical URL if valid.
	if canonical, err := normalizeURL(pageData.CanonicalURL); err == nil {
		pageData.CanonicalURL = canonical
	}

	// Normalize internal and external links.
	pageData.InternalLinks = normalizeURLs(pageData.InternalLinks)
	pageData.ExternalLinks = normalizeURLs(pageData.ExternalLinks)

	// Language detection.
	lang, err := detectLanguage(pageData.VisibleText)
	if err != nil {
		log.Printf("language detection failed: %v", err)
	}
	if lang != "en" {
		return errors.New("not an English page, skipping")
	}
	pageData.Language = lang

	// Spam filtering.
	if isSpam(pageData.VisibleText) {
		return errors.New("spam detected, skipping")
	}

	return nil
}

// Removes extra whitespace and newlines.
func basicHTMLCleanup(input string) string {
	cleaned := strings.TrimSpace(input)
	return strings.Join(strings.Fields(cleaned), " ")
}

// Trims, parses, and normalizes a URL.
func normalizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty URL")
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	return parsedURL.String(), nil
}

// Processes a slice of URLs and returns only those that are valid.
func normalizeURLs(urls []string) []string {
	var result []string
	for _, link := range urls {
		if normalized, err := normalizeURL(link); err == nil {
			result = append(result, normalized)
		}
	}
	return result
}

// Placeholder. Returns "en" if the text is long enough.
func detectLanguage(text string) (string, error) {
	if len(text) < 10 {
		return "", errors.New("text too short for language detection")
	}
	return "en", nil
}

// Placeholder. Implements a naive spam check.
func isSpam(text string) bool {
	if len(text) < 30 {
		return true
	}
	suspiciousKeywords := []string{"buy now", "cheap pills", "viagra", "bitcoin scam"}
	lower := strings.ToLower(text)
	for _, keyword := range suspiciousKeywords {
		if strings.Count(lower, keyword) > 3 {
			return true
		}
	}
	return false
}
