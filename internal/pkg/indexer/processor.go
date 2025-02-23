package indexer

import (
    "errors"
    "net/url"
    "strings"
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

// Does two main things for now:
// 1) Basic "HTML cleanup" on text fields
// 2) URL normalization (for canonical_url, external_links, internal_links, etc.)
func (p *processor) CleanAndNormalize(pd models.PageData) (models.PageData, error) {
    // 1) Basic HTML cleanup:
    cleanedText := basicHTMLCleanup(pd.VisibleText)
    pd.VisibleText = cleanedText

    // 2) URL normalization:
    //    - pd.URL
    //    - pd.CanonicalURL
    //    - pd.InternalLinks
    //    - pd.ExternalLinks

    normalizedURL, err := normalizeURL(pd.URL)
    if err != nil {
        return pd, err
    }
    pd.URL = normalizedURL

    normalizedCanonical, err := normalizeURL(pd.CanonicalURL)
    if err == nil { // if CanonicalURL is empty or invalid, skip
        pd.CanonicalURL = normalizedCanonical
    }

    var normalizedInternal []string
    for _, link := range pd.InternalLinks {
        normalized, err := normalizeURL(link)
        if err == nil {
            normalizedInternal = append(normalizedInternal, normalized)
        }
        // If err != nil, skip that link
    }
    pd.InternalLinks = normalizedInternal

    var normalizedExternal []string
    for _, link := range pd.ExternalLinks {
        normalized, err := normalizeURL(link)
        if err == nil {
            normalizedExternal = append(normalizedExternal, normalized)
        }
    }
    pd.ExternalLinks = normalizedExternal

    return pd, nil
}

// For now, this is a naive placeholder that just trims and collapses whitespace.
// In future steps, we will do more sophisticated HTML sanitization, remove scripts, tags, etc.
func basicHTMLCleanup(input string) string {
    cleaned := strings.TrimSpace(input)
    cleaned = strings.Join(strings.Fields(cleaned), " ")
    return cleaned
}

// Uses net/url to parse and normalize a given URL/string.
// If invalid, we return an error so we can skip or handle gracefully in the caller.
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

    // We could do more advanced normalization: removing default ports, removing tracking params, etc.
    // For now, let's keep it minimal.
    return u.String(), nil
}
