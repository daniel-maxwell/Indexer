package models

import "time"

// Data structure to organize and store relevant information from the page
type PageData struct {
    URL             string              `json:"url"`
    CanonicalURL    string              `json:"canonical_url"`
    Title           string              `json:"title,omitempty"`
    Charset         string              `json:"charset,omitempty"`
    MetaDescription string              `json:"meta_description,omitempty"`
    MetaKeywords    string              `json:"meta_keywords,omitempty"`
    Language        string              `json:"language,omitempty"`
    Headings        map[string][]string `json:"headings,omitempty"`
    AltTexts        []string            `json:"alt_texts,omitempty"`
    AnchorTexts     []string            `json:"anchor_texts,omitempty"`
    InternalLinks   []string            `json:"internal_links,omitempty"`
    ExternalLinks   []string            `json:"external_links,omitempty"`
    StructuredData  []string            `json:"structured_data,omitempty"`
    OpenGraph       map[string]string   `json:"open_graph,omitempty"`
    DatePublished   time.Time           `json:"date_published,omitempty"`
    DateModified    time.Time           `json:"date_modified,omitempty"`
    SocialLinks     []string            `json:"social_links,omitempty"`
    VisibleText     string              `json:"visible_text,omitempty"`
    LoadTime        time.Duration       `json:"load_time,omitempty"`
    IsSecure        bool                `json:"is_secure,omitempty"`
    FetchError      string              `json:"fetch_error"`
}
