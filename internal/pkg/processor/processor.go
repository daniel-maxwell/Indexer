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

// Defines all the steps needed to transform raw PageData
// into a cleaned, normalized version that’s ready for further checks.
type Processor interface {
    CleanAndNormalize(pd models.PageData) (models.PageData, error)
}

// Default implementation of Processor.
type processor struct{}

// Creates a new Processor instance.
func NewProcessor() Processor {
    return &processor{}
}

func (p *processor) CleanAndNormalize(pd models.PageData) (models.PageData, error) {
    // Basic HTML cleanup
    pd.VisibleText = basicHTMLCleanup(pd.VisibleText)

    // URL normalization for primary URL, canonical URL, internal/external links
    var err error
    pd.URL, err = normalizeURL(pd.URL)
    if err != nil {
        // We can choose to skip if URL is invalid, or just log the error
        log.Printf("invalid URL %q: %v", pd.URL, err)
        return pd, err
    }

    canonical, err := normalizeURL(pd.CanonicalURL)
    if err == nil {
        pd.CanonicalURL = canonical
    }

    pd.InternalLinks = normalizeURLs(pd.InternalLinks)
    pd.ExternalLinks = normalizeURLs(pd.ExternalLinks)

    // Language detection
    lang, err := detectLanguage(pd.VisibleText)
    if err != nil {
        log.Printf("language detection failed: %v", err)
        return pd, err
    }
    if lang != "en" {
        // Not English, we skip
        return pd, errors.New("not an English page, skipping")
    }
    pd.Language = lang

    // Spam filtering
    if isSpam(pd.VisibleText) {
        return pd, errors.New("spam detected, skipping")
    }

    // If we reach here, we have a cleaned, normalized, English, non-spam page
    return pd, nil
}

// Basic HTML cleanup: removes extra whitespace and newlines
func basicHTMLCleanup(input string) string {
    cleaned := strings.TrimSpace(input)
    cleaned = strings.Join(strings.Fields(cleaned), " ")
    return cleaned
}

// Normalizes a URL, returning an error if it fails parsing
func normalizeURL(raw string) (string, error) {
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return "", errors.New("empty URL")
    }

    u, err := url.Parse(raw)
    if err != nil {
        return "", err
    }

    u.Scheme = strings.ToLower(u.Scheme)
    u.Host = strings.ToLower(u.Host)
    return u.String(), nil
}

// Normalizes a slice of URLs, skipping those that fail parsing
func normalizeURLs(urls []string) []string {
    var result []string
    for _, link := range urls {
        normalized, err := normalizeURL(link)
        if err == nil {
            result = append(result, normalized)
        }
    }
    return result
}

// For now, this is a placeholder that always returns "en" or an error.
func detectLanguage(text string) (string, error) {
    if len(text) < 10 {
        return "", errors.New("text too short for language detection")
    }
    return "en", nil
}

// Very naive spam checker for now. Does a simple check for either 
// extremely short text or certain suspicious keywords repeated too often.
func isSpam(text string) bool {
    // Too short
    if len(text) < 30 {
        return true
    }

    // Suspicious repeated keywords
    suspiciousKeywords := []string{"buy now", "cheap pills", "viagra", "bitcoin scam"}
    lower := strings.ToLower(text)
    for _, keyword := range suspiciousKeywords {
        if strings.Count(lower, keyword) > 3 {
            return true
        }
    }
    return false
}
