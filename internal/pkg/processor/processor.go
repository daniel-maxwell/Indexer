package processor

import (
    "errors"
    "net/url"
    "strings"
    "log"
	"time"
	"go.uber.org/zap"
	"github.com/pemistahl/lingua-go"
	"indexer/internal/pkg/logger"
	"indexer/internal/pkg/deduplicator"
	"indexer/internal/pkg/processor/languagedetector"
	"indexer/internal/pkg/processor/spamdetector"
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
	spamDetector *spamdetector.SpamDetector
}

// Creates a new Processor instance and wires in the sub‑components.
func NewProcessor(deduper deduper.Deduper, nlpServiceURL string, spamThreshold int) Processor {
    return &processor{
        deduper:  deduper,
        enricher: NewNLPEnricher(nlpServiceURL),
		spamDetector: spamdetector.NewSpamDetector(spamThreshold),
    }
}

// Global language detector singleton to avoid repeated initialization
var languageDetector lingua.LanguageDetector

// Initializes the language detector once
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

	// Store signature
	processor.deduper.StoreSignature(signature)

	// Language detection
	if err := detectLanguage(pageData); err != nil {
		return err
	}
	
	// Spam detection
	if err := processor.detectSpam(pageData, doc); err != nil {
		return err
	}
	// Record spam score metrics
	metrics.SpamScoreHistogram.Observe(float64(doc.SpamScore))
	
    // Enrich doc
    if err := processor.enricher.Enrich(pageData, doc); err != nil {
        return err
    }

	// Update quality score based on spam score
	// Higher spam score means lower quality
	if doc.SpamScore > 0 {
		qualityPenalty := doc.SpamScore * 2
		if doc.QualityScore > qualityPenalty {
			doc.QualityScore -= qualityPenalty
		} else {
			doc.QualityScore = 0
		}
	}

	// Increment metrics
	metrics.PagesProcessed.Inc()

    return nil
}

// Applies cleaning, URL normalization, language detection,
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
    
    // Handle relative URLs
    if !strings.Contains(rawURL, "://") && !strings.HasPrefix(rawURL, "//") {
        return "", errors.New("relative URL without base")
    }
    
    // Handle scheme-relative URLs (starting with //)
    if strings.HasPrefix(rawURL, "//") {
        rawURL = "https:" + rawURL
    }
    
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return "", err
    }
    
    // Ensure scheme is set
    if parsedURL.Scheme == "" {
        parsedURL.Scheme = "https"
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

// Detects the language of the visible text and updates the PageData.
func detectLanguage(pageData *models.PageData) error {
    start := time.Now()

	lang, err := languagedetector.DetectLanguage(languageDetector, pageData.VisibleText)

    metrics.LanguageDetectionLatency.Observe(time.Since(start).Seconds())
    
	if err != nil {
		if strings.Contains(err.Error(), "not an English page") {
			logger.Log.Info("Skipping non-English page", 
				zap.String("url", pageData.URL), 
				zap.String("detected_language", lang))
			return errors.New("not an English page, skipping")
		}
		logger.Log.Warn("Language detection failed", zap.Error(err))
		metrics.LanguageDetectionFailures.Inc()
		pageData.Language = "unknown"
	} else {
		pageData.Language = lang
	}

	return nil
}

// Detects spam content in the visible text and updates the Document.
func (processor *processor) detectSpam(pageData *models.PageData, doc *models.Document) error {
	// Spam detection with timing
	spamStart := time.Now()
	spamResult := processor.spamDetector.DetectSpam(pageData.VisibleText)
	metrics.SpamDetectionLatency.Observe(time.Since(spamStart).Seconds())
	
	// Store spam score and matched phrases in the document
	doc.SpamScore = spamResult.Score
	
	logger.Log.Debug("Spam detection result", 
		zap.String("url", pageData.URL),
		zap.Int("spam_score", spamResult.Score),
		zap.Bool("is_high_spam", spamResult.IsHighSpam))
	
	// If high spam, abort processing
	if spamResult.IsHighSpam {
		metrics.HighSpamPagesSkipped.Inc()
		logger.Log.Info("Skipping high spam content", 
			zap.String("url", pageData.URL), 
			zap.Int("spam_score", spamResult.Score))
		return errors.New("high spam content detected, skipping")
	}

	return nil
}