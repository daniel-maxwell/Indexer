package models

import (
	"time"
)

// IndexAction represents the metadata for a bulk index operation.
type IndexAction struct {
	Index struct {
		Index string `json:"_index"`
		ID    string `json:"_id"`
	} `json:"index"`
}

// Output document data to be indexed by Elasticsearch.
type Document struct {
	URL              string         `json:"url"`
	CanonicalURL     string         `json:"canonical_url"`
	Title            string         `json:"title"`
	MetaDescription  string         `json:"meta_description"`
	VisibleText      string         `json:"visible_text"`
	Entities         []string       `json:"entities"`
	Keywords         []string       `json:"keywords"`
	InternalLinks    []string       `json:"internal_links"`
	ExternalLinks    []string       `json:"external_links"`
	StructuredData   StructuredData `json:"structured_data"`
	OpenGraph        OpenGraph      `json:"open_graph"`
	DatePublished    time.Time      `json:"date_published"`
	DateModified     time.Time      `json:"date_modified"`
	Categories       []string       `json:"categories"`
	Tags             []string       `json:"tags"`
	SocialLinks      []string       `json:"social_links"`
	LoadTime         int64          `json:"load_time"`
	IsSecure         bool           `json:"is_secure"`
	QualityScore     int        	`json:"quality_score"` // Out of 100
	InboundLinkCount int            `json:"inbound_link_count"`
	LastCrawled      time.Time      `json:"last_crawled"`
}

// Structured data block.
type StructuredData struct {
	Context string `json:"@context"`
	Type    string `json:"@type"`
}

// OpenGraph metadata.
type OpenGraph struct {
	OGTitle       string `json:"og:title"`
	OGDescription string `json:"og:description"`
	OGImage       string `json:"og:image"`
}
