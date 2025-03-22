package processor

import (
    "errors"
    "net/url"
    "strings"
    "log"
	"time"
	"go.uber.org/zap"
	"github.com/pemistahl/lingua-go"
	// "github.com/microcosm-cc/bluemonday"
    // "golang.org/x/net/html"
	"indexer/internal/pkg/logger"
	"indexer/internal/pkg/deduplicator"
    "indexer/internal/pkg/models"
	"indexer/internal/pkg/metrics"
)

// Defines the high-level interface for processing page data.
type Processor interface {
	// Process runs the complete data processing pipeline.
	// It operates directly on the provided PageData and Document.
	Process(pageData *models.PageData, doc *models.Document) error
}

// The default implementation of Processor.
type processor struct {
	deduper  deduper.Deduper
	enricher Enricher
}

// Creates a new Processor instance and wires in the sub‑components.
func NewProcessor(deduper deduper.Deduper, nlpServiceURL string) Processor {
    return &processor{
        deduper:  deduper,
        enricher: NewNLPEnricher(nlpServiceURL),
    }
}

// Global language detector singleton to avoid repeated initialization
var languageDetector lingua.LanguageDetector

// init initializes the language detector once
func init() {
	// Build the detector with preloaded models for better performance
	languageDetector = lingua.NewLanguageDetectorBuilder().
	FromAllLanguages().
	WithPreloadedLanguageModels().
	Build()
}

// Runs the data processing pipeline:
// cleaning/normalization, deduplication, and enrichment.
func (processor *processor) Process(pageData *models.PageData, doc *models.Document) error {
    // Clean & normalize
    if err := cleanAndNormalize(pageData, doc); err != nil {
        return err
    }

    // Dedup check
    signature := deduper.GenerateSignature(pageData.VisibleText)
    if processor.deduper.IsDuplicate(signature) {
        return errors.New("duplicate page detected")
    }

    // Enrich doc
    if err := processor.enricher.Enrich(pageData, doc); err != nil {
        return err
    }

    // Store signature
    processor.deduper.StoreSignature(signature)

	// Increment metrics
	metrics.PagesProcessed.Inc()

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
	// Language detection with timing
    start := time.Now()
    lang, err := detectLanguage(pageData.VisibleText)
    metrics.LanguageDetectionLatency.Observe(time.Since(start).Seconds())
    
    if err != nil {
        if strings.Contains(err.Error(), "not an English page") {
            logger.Log.Info("Skipping non-English page", 
                zap.String("url", pageData.URL), 
                zap.String("detected_language", lang))
            return errors.New("not an English page, skipping")
        }
        logger.Log.Warn("Language detection failed", zap.Error(err))
        // Continue processing even if language detection fails
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
	return strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
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
    const minTextLength = 20
    if len(text) < minTextLength {
        return "unknown", nil
    }

    // Detect language and calculate confidence values
    detectedLang, exists := languageDetector.DetectLanguageOf(text)
    if !exists {
        metrics.LanguageDetectionFailures.Inc()
        return "", errors.New("language detection failed")
    }

    // Get confidence values for all languages
    confidenceValues := languageDetector.ComputeLanguageConfidenceValues(text)
    var englishConfidence float64

    // Find English confidence value
    for _, conf := range confidenceValues {
        if conf.Language() == lingua.English {
            englishConfidence = conf.Value()
            break
        }
    }

    logger.Log.Debug("Language detection result", 
        zap.String("detected_language", detectedLang.String()),
        zap.Float64("english_confidence", englishConfidence))

	if detectedLang == lingua.English {
		return "en", nil
	} else if englishConfidence > 0.33 {
		return detectedLang.IsoCode639_1().String(), nil
	}

    // If not English or low confidence, skip this document
    metrics.NonEnglishPagesSkipped.Inc()
    return detectedLang.IsoCode639_1().String(), errors.New("not an English page, skipping")
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